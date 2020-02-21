package main

import (
	"fmt"
	eventbus "github.com/yhyddr/redbus/v0"
)

func main() {
	eb := eventbus.New()

	ep := eb.Attach("keyboard")

	fmt.Println(ep.Descriptor(),ep.Probe())

	p := eventbus.NewPong()
	eb.Send("keyboard",p)

	e, err := ep.Receive()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(e.Type(),e.Payload())
}
