package main

import (
	"bufio"
	"fmt"
	"net"
	 "gopkg.in/natefinch/npipe.v2"

)

func main() {
	ln, err := npipe.Listen(`\\.\pipe\mypipe`)
	if err != nil {
		// handle error
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			continue
		}

		// handle connection like any other net.Conn
		go func(conn net.Conn) {
			r := bufio.NewReader(conn)
			msg, err := r.ReadString('\n')
			if err != nil {
				// handle error
				return
			}
			fmt.Println(msg)
		}(conn)
	}
}
