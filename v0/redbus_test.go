package eventbus_test

import (
	"fmt"
	"github.com/yhyddr/redbus/v0"
	"testing"
)

func TestRedbus(t *testing.T) {
	eb := eventbus.New()

	ep := eb.Attach("keyboard")

	fmt.Println(ep.Descriptor())
	fmt.Println(ep.Probe())

	p := eventbus.NewPong()
	eb.Send("keyboard",p)

	e, err := ep.Receive()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(e.Type(),e.Payload())

	ep.Detach()

	fmt.Println(ep.Probe())

	e, err = ep.Receive()

	fmt.Println(e,err)
}
