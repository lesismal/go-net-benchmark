package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
	"github.com/lesismal/nbio"
	nlog "github.com/lesismal/nbio/logging"
)

var port = flag.Int("p", 8000, "server addr")
var rpcPort = flag.Int("r", 9000, "rpc server addr")

func main() {
	flag.Parse()

	alog.SetLevel(alog.LevelNone)
	nlog.SetLevel(nlog.LevelNone)

	g := nbio.NewGopher(nbio.Config{
		Network: "tcp",
		Addrs:   []string{fmt.Sprintf(":%v", *port)},
	})

	g.OnData(func(c *nbio.Conn, data []byte) {
		c.Write(data)
	})

	err := g.Start()
	if err != nil {
		log.Fatalf("nbio.Start failed: %v\n", err)
		return
	}

	recorder := perf.NewRecorder("server@nbio")

	svr := arpc.NewServer()
	svr.Handler.Handle("action", func(ctx *arpc.Context) {
		cmd := ""
		ctx.Bind(&cmd)
		switch cmd {
		case "begin":
			recorder.Begin()
			ctx.Write(nil)
		case "end":
			recorder.End()
			ctx.Write(recorder.ReportString())
		}
	})
	defer svr.Stop()

	log.Fatal(svr.Run(fmt.Sprintf(":%v", *rpcPort)))
}
