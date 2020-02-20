package main

import (
	"fmt"
	"github.com/yhyddr/redbus/v0"
)

func main() {
	b := redbus.New()

	ep := b.Attach("room")

	var ev []redbus.Event
	b.Dispatch("room", redbus.Event{PID: "Hello World"})
	ep.Read(&ev)

	fmt.Println(ev)
	fmt.Println(ev[0].Type())
}
