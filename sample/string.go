package main

import (
	"encoding/json"
	"log"
)

func main() {

	var str = "hello"
	// log.Println(str, "--", []byte(str+stri{des: str}.String()))

	jstr, _ := json.Marshal(stri{des: str})
	log.Println(string(jstr))
}

type stri struct {
	des string
}

func (s stri) String() string {
	return s.des
}
