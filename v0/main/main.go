package main

import (
	"fmt"
	"time"
)

func main() {
	x := make(chan struct{},1)
	fmt.Println(len(x))
	x <- struct{}{}
	fmt.Println(len(x))
	time.Sleep(time.Second)
}
