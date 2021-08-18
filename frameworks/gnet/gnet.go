package main

import (
	"fmt"
	"log"
	"net"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
	"github.com/panjf2000/gnet"
)

var port = 8003
var rpcPort = ":9003"

type echoServer struct {
	*gnet.EventServer
}

func (es *echoServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	out = append([]byte{}, frame...)
	return
}

func main() {
	alog.SetLevel(alog.LevelNone)

	echo := new(echoServer)

	go func() {
		log.Fatal(gnet.Serve(echo, fmt.Sprintf("tcp://:%d", port), gnet.WithMulticore(true), gnet.WithReusePort(false)))
	}()

	recorder := perf.NewRecorder("server@gnet")

	svr := arpc.NewServer()
	svr.Handler.Handle("Hello", func(ctx *arpc.Context) {
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

	ln, err := net.Listen("tcp", rpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	svr.Serve(ln)
}
