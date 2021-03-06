package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/cloudwego/netpoll"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
)

var port = flag.Int("p", 8000, "server addr")
var rpcPort = flag.Int("r", 9000, "rpc server addr")

func main() {
	flag.Parse()

	alog.SetLevel(alog.LevelNone)

	// 创建 listener
	listener, err := netpoll.CreateListener("tcp", fmt.Sprintf(":%v", *port))
	if err != nil {
		panic("create netpoll listener fail")
	}

	// handle: 连接读数据和处理逻辑
	var onRequest netpoll.OnRequest = handler

	// options: EventLoop 初始化自定义配置项
	var opts = []netpoll.Option{
		netpoll.WithReadTimeout(5 * time.Second),
		netpoll.WithIdleTimeout(10 * time.Minute),
		netpoll.WithOnPrepare(nil),
	}

	// 创建 EventLoop
	eventLoop, err := netpoll.NewEventLoop(onRequest, opts...)
	if err != nil {
		log.Fatalf("create netpoll event-loop fail: %v", err)
	}

	// 运行 Server
	go func() {
		err = eventLoop.Serve(listener)
		if err != nil {
			panic("netpoll server exit")
		}
	}()

	recorder := perf.NewRecorder("server@netpoll")

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

// 读事件处理
func handler(ctx context.Context, connection netpoll.Connection) error {
	reader := connection.Reader()
	l := reader.Len()
	buf, err := reader.Next(l)
	if err != nil {
		return err
	}

	_, err = connection.Write(buf)

	return err
}
