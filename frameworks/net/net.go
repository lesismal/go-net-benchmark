package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
)

var port = flag.Int("p", 8000, "server addr")
var rpcPort = flag.Int("r", 9000, "rpc server addr")

func handle(conn net.Conn) {
	buf := make([]byte, 1024*32)
	for {
		nread, err := conn.Read(buf)
		if err != nil {
			return
		}
		nwrite, err := conn.Write(buf[:nread])
		if err != nil {
			return
		}
		if nwrite != nread {
			return
		}
	}
}
func main() {
	alog.SetLevel(alog.LevelNone)

	go func() {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%v", *port))
		if err != nil {
			log.Fatalf("listen failed: %v", err)
		}
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatalf("Accept failed: %v", err)
				return
			}
			go handle(conn)
		}
	}()

	recorder := perf.NewRecorder("server@net")

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
