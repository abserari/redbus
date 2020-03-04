# Redbus

a eventbus which attempt to use redis pub/sub to communicate.

## Usage
### v1
```go
import "github.com/yhyddr/redbus"

func add(e *Event) {
   a := strconv.Atoi(e.Get("a"))
   b := strconv.Atoi(e.Get("b"))
   fmt.Println(a+b)
}

func main() {
    redbus.HandleFunc("/add", add)
    go redbus.Serve()
    msg := make(http.Header)
    msg.Add("a", "1")
    msg.Add("b", "2")
    redbus.Publish("/add",msg)
}
```

### single
```go
import eventbus "github.com/yhyddr/redbus/single"

func main() {
	eb := eventbus.New()

	ep := eb.Attach("keyboard")

	fmt.Println(ep.Descriptor())
	fmt.Println(ep.Probe())

	p := eventbus.NewPong()
	eb.Send("keyboard",p)

	e, err := ep.Receive()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(e.Type(),e.Payload())

	ep.Detach()

	fmt.Println(ep.Probe())

	e, err = ep.Receive()

    fmt.Println(e,err)
}
```
### multiple
#### server
```go
import (
	eventbus "github.com/yhyddr/redbus/multi"
)

func main() {
	eb := eventbus.New()
	eb.Serve()
}

```
#### suber
```go
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

```
### udp

### sidecar