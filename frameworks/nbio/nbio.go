package main

import (
	"log"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
	"github.com/lesismal/nbio"
	nlog "github.com/lesismal/nbio/logging"
)

var port = ":8002"
var rpcPort = ":9002"

func main() {
	alog.SetLevel(alog.LevelNone)
	nlog.SetLevel(nlog.LevelNone)

	g := nbio.NewGopher(nbio.Config{
		Network: "tcp",
		Addrs:   []string{port},
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

	log.Fatal(svr.Run(rpcPort))
}
