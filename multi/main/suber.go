package main

import (
	"fmt"
	"log"
	"time"

	eventbus "github.com/yhyddr/redbus/multi"
)

func main() {
	eb := eventbus.New()

	pong := eventbus.NewPong()

	ep := eb.Attach("room")
	defer ep.Detach()
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
