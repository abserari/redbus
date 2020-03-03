package main

import "log"

func main() {

	var str = "hello"
	log.Println(str, "--", []byte(str+stri{des: str}.String()))
}

type stri struct {
	des string
}

func (s stri) String() string {
	return s.des
}
