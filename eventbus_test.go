package eventbus_test

import (
	"testing"
	"time"

	evb "muses.service/service/eventbus"
)

func TestEventBus(t *testing.T) {
	e := evb.NewDefault()
	channel1 := e.Subscribe("room1")
	channel2 := e.Subscribe("room2")
	channel3 := e.Subscribe("room3")

	// should be wrapped to a function to get go func
	// or add callback listen to event
	go func() {
		for msg := range channel1 {
			t.Log("receive: ", msg.Channel, msg.Pattern, msg.Payload)
		}
	}()
	go func() {
		for msg := range channel2 {
			t.Log("receive: ", msg.Channel, msg.Pattern, msg.Payload)
		}
	}()
	go func() {
		for msg := range channel3 {
			t.Log("receive: ", msg.Channel, msg.Pattern, msg.Payload)
		}
	}()

	for i := 0; i < 10; i++ {
		if i == 5 {
			e.UnSubscribe("room1")
		}
		e.PPublish("room*", "hello", "world")
	}
	// using this to log messages
	// t.Fail()
	time.Sleep(time.Second * 2)

}

func TestEventBusUrl(t *testing.T) {
	e, err := evb.New()
	if err != nil {
		t.Error(err)
	}
	channel1 := e.PSubscribe("room*")

	flag := 0
	// should be wrapped to a function to get go func
	// or add callback listen to event
	go func() {
		for msg := range channel1 {
			t.Log("receive: ", msg.Channel, msg.Pattern, msg.Payload)
			if msg.Channel == "room11" && msg.Pattern == "room*" {
				flag = 1
			}
		}
	}()

	for i := 0; i < 10; i++ {
		if i == 5 {
			if err := e.UnSubscribe("room1", "room2", ""); err != nil {
				t.Error(err)
			}
		}
		e.Broadcast("hello", "world")
		e.Publish("room11", "hello")
	}
	// using this to log messages
	// t.Fail()
	time.Sleep(time.Second * 2)
	if flag != 1 {
		t.Error("can't close PSub")
	}

}

// a out-of-data test
// func TestEvent(t *testing.T) {
// 	b, err := evb.New()
// 	if err != nil {
// 		t.Fatal("Couldn't read env url to create evb")
// 	}
// 	go b.Register("log", func(a string) { fmt.Println(a) })
// 	time.Sleep(time.Second * 1)
// 	b.Publish("log", "hello")

// 	b.Publish("log", "/exit")
// 	time.Sleep(time.Second * 1)
// }
