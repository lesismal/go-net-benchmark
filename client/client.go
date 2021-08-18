package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/cloudwego/kitex-benchmark/runner"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
)

var (
	port          = flag.Int("p", 8000, "server addr")
	rpcPort       = flag.Int("r", 9000, "rpc server addr")
	framework     = flag.String("f", "none", "framework name")
	connectionNum = flag.Int("c", 100, "connection num")
	total         = flag.Int64("n", 10000000, "total test time")
	bufsize       = flag.Int("b", 1024, "buffer size")
)

func main() {
	flag.Parse()

	alog.SetLevel(alog.LevelNone)

	client, err := arpc.NewClient(func() (net.Conn, error) {
		return net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%v", *rpcPort), time.Second*3)
	})
	if err != nil {
		log.Fatalf("NewClient failed: %v", err)
	}
	defer client.Stop()

	chTask := make(chan chan error, *connectionNum)

	conns := make([]net.Conn, *connectionNum)
	for i := 0; i < *connectionNum; i++ {
		c, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%v", *port), time.Second*3)
		if err != nil {
			log.Fatalf("dial failed: %v", err)
		}
		conns = append(conns, c)

		go func() {
			defer c.Close()

			request := make([]byte, *bufsize)
			response := make([]byte, *bufsize)
			rand.Read(request)
			for waitting := range chTask {
				nwrite, err := c.Write(request)
				if err != nil || nwrite != *bufsize {
					log.Fatalf("write failed: %v, %v", nwrite, err)
				}
				nread, err := io.ReadFull(c, response)
				if err != nil {
					log.Fatalf("read failed: %v", err)
				}
				if nread != nwrite || !bytes.Equal(request, response) {
					log.Fatalf("not equal: %v / %v", nread, nwrite)
				}
				waitting <- err
			}
		}()
	}

	r := runner.NewRunner()

	handler := func() error {
		waitting := make(chan error, 1)
		chTask <- waitting
		return <-waitting
	}

	r.Warmup(handler, *connectionNum, 100*1000)

	err = client.Call("action", "begin", nil, time.Second)
	if err != nil {
		log.Fatalf("call begain failed: %v", err)
	}

	recorder := perf.NewRecorder(fmt.Sprintf("client@%v", *framework))
	recorder.Begin()

	r.Run(*framework, handler, *connectionNum, *total, *bufsize, 0)

	recorder.End()

	serverReport := ""
	err = client.Call("action", "end", &serverReport, time.Second)
	if err != nil {
		log.Fatalf("call begain failed: %v", err)
	}
	fmt.Print(serverReport)

	recorder.Report()
	fmt.Printf("\n\n")
}
