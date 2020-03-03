package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	eventbus "github.com/yhyddr/redbus/multi"
)

func main() {
	eb := eventbus.New()

	pong := eventbus.NewPong()

	ep := eb.Attach("room")

	gracefullyShutdown(ep.Detach)
	for {
		ep.Send("room", pong)

		e, err := ep.Receive()
		if err != nil {
			fmt.Println(err)
		}
		log.Println("event:", e, "\n")
		time.Sleep(time.Second)
	}
}

func gracefullyShutdown(fn func()) {
	c := make(chan os.Signal)
	// 监听信号
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)

	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
				fmt.Println("退出:", s)
				fn()
				os.Exit(0)
			case syscall.SIGUSR1:
				fmt.Println("usr1", s)
			case syscall.SIGUSR2:
				fmt.Println("usr2", s)
			default:
				fmt.Println("其他信号:", s)
			}
		}
	}()
}
