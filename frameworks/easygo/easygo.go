package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"

	//"time"

	"github.com/cloudwego/kitex-benchmark/perf"
	"github.com/lesismal/arpc"
	alog "github.com/lesismal/arpc/log"
	"github.com/mosn/easygo/netpoll"
	"golang.org/x/sys/unix"
)

var port = flag.Int("p", 8000, "server addr")
var rpcPort = flag.Int("r", 9000, "rpc server addr")

var _EPOLLCLOSED netpoll.EpollEvent = 0x20
var pool = NewWorkerPool(runtime.NumCPU() * 4)

func main() {
	alog.SetLevel(alog.LevelNone)

	ep, err := netpoll.EpollCreate(epollConfig())
	if err != nil {
		log.Fatal(err)
	}

	ep2, err := netpoll.EpollCreate(epollConfig())
	if err != nil {
		log.Fatal(err)
	}

	ln, err := listen(*port)
	if err != nil {
		log.Fatal(err)
	}
	defer unix.Close(ln)

	ep.Add(ln, netpoll.EPOLLIN, func(evt netpoll.EpollEvent) {
		if evt&_EPOLLCLOSED != 0 {
			return
		}

		conn, _, err := unix.Accept(ln)
		if err != nil {
			log.Fatalf("could not accept: %s", err)
		}

		pool.ScheduleAuto(func() {
			unix.SetNonblock(conn, true)

			readBuf := make([]byte, 1024*16)

			ep2.Add(conn, netpoll.EPOLLIN|netpoll.EPOLLET|netpoll.EPOLLHUP|netpoll.EPOLLRDHUP, func(evt netpoll.EpollEvent) {
				if evt&_EPOLLCLOSED != 0 {
					return
				}

				pool.ScheduleAuto(func() {
					n, err := unix.Read(conn, readBuf[:])
					if err != nil {
						return
					}
					n, err = unix.Write(conn, readBuf[:n])
				})
			})
		})
	})

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

func epollConfig() *netpoll.EpollConfig {
	return &netpoll.EpollConfig{
		OnWaitError: func(err error) {
			log.Fatal(err)
		},
	}
}

func listen(port int) (ln int, err error) {
	ln, err = unix.Socket(unix.AF_INET, unix.O_NONBLOCK|unix.SOCK_STREAM, 0)
	if err != nil {
		return
	}

	unix.SetsockoptInt(ln, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)

	addr := &unix.SockaddrInet4{
		Port: port,
		Addr: [4]byte{0x7f, 0, 0, 1}, // 127.0.0.1
	}

	if err = unix.Bind(ln, addr); err != nil {
		return
	}
	err = unix.Listen(ln, 4)

	return
}

type workerPool struct {
	work chan func()
	sem  chan struct{}
}

func NewWorkerPool(size int) WorkerPool {
	return &workerPool{
		work: make(chan func()),
		sem:  make(chan struct{}, size),
	}
}

func (p *workerPool) Schedule(task func()) {
	select {
	case p.work <- task:
	case p.sem <- struct{}{}:
		go p.spawnWorker(task)
	}
}

func (p *workerPool) ScheduleAlways(task func()) {
	select {
	case p.work <- task:
	case p.sem <- struct{}{}:
		go p.spawnWorker(task)
	default:
		// new temp goroutine for task execution
		// log.Printf("[syncpool] workerpool new goroutine")
		go task()
	}
}

func (p *workerPool) ScheduleAuto(task func()) {
	select {
	case p.work <- task:
		return
	default:
	}
	select {
	case p.work <- task:
	case p.sem <- struct{}{}:
		go p.spawnWorker(task)
	default:
		// new temp goroutine for task execution
		// log.Printf("[syncpool] workerpool new goroutine")
		go task()
	}
}

func (p *workerPool) spawnWorker(task func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("syncpool", "[syncpool] panic %v\n%s", p, string(debug.Stack()))
		}
		<-p.sem
	}()
	for {
		task()
		task = <-p.work
	}
}

type WorkerPool interface {
	Schedule(task func())
	ScheduleAlways(task func())
	ScheduleAuto(task func())
}
