package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
	"github.com/panjf2000/gnet"
)

var port = flag.Int("p", 8000, "server addr")
var rpcPort = flag.Int("r", 9000, "rpc server addr")

type echoServer struct {
	*gnet.EventServer
}

func (es *echoServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	out = frame
	return
}

func main() {
	flag.Parse()

	alog.SetLevel(alog.LevelNone)

	echo := new(echoServer)

	go func() {
		log.Fatal(gnet.Serve(echo, fmt.Sprintf("tcp://:%d", *port), gnet.WithMulticore(true), gnet.WithReusePort(false)))
	}()

	recorder := perf.NewRecorder("server@gnet")

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
