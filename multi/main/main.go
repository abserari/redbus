package main

import (
	eventbus "github.com/yhyddr/redbus/multi"
)

func main() {
	eb := eventbus.New()
	eb.Serve()
}
