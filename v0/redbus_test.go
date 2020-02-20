package redbus_test

import (
	"fmt"
	"github.com/yhyddr/redbus/v0"
	"testing"
)

func TestRedbus(t *testing.T) {
	b := redbus.New()

	ep:= b.Attach("room")

	var ev redbus.Event
	go ep.Read(ev)

	b.Dispatch("room",&redbus.Data{PID:"Hello World"})
	for(ev == nil)  {}
	fmt.Println(ev.Type())
}
