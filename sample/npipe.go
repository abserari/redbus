package main

import (
	"bufio"
	"fmt"
	npipe "gopkg.in/natefinch/npipe.v2"
)


func main() {
	endpoint:=`\.pipetest-multi-22244-5577006791947779410`
	conn, err := npipe.Dial(endpoint)
	if err != nil {
		// handle error
	}
	if _, err := fmt.Fprintln(conn, "Hi server!"); err != nil {
		// handle error
		fmt.Println(err)
	}
	r := bufio.NewReader(conn)
	msg, err := r.ReadString('n')
	if err != nil {
		// handle eror
		fmt.Println(err)
	}
	fmt.Println(msg)
}


