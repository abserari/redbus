package eventbus

import (
	"errors"
	"strconv"
	"strings"
	"sync"
)

//if need stream read write, read about https://github.com/grpc/grpc-go/blob/master/stream.go#L122:2

// Endpoint means a middleware which does nothing more but send and receive Event.
type Endpoint interface {
	Descriptor() string
	Receive() (Event, error)
	Detach()
	Probe() bool
}

// EventBus manager the register receiver and send Event to them.
type EventBus interface {
	Attach(topic string) Endpoint
	Send(topic string, e Event)
}

// Event is messages' type.
type Event interface {
	Type() string
	Payload() []byte
}

// Pong realized Event.
type Pong struct {
	pID     string
	payload []byte
}

// NewPong new a pong struct.
func NewPong() Pong { return Pong{pID: "pong", payload: []byte("pong")} }

// Type return Event type.
func (p Pong) Type() string { return p.pID }

// Payload return data of Event.
func (p Pong) Payload() []byte { return p.payload }

// Bus relized Eventbus
type Bus struct {
	epcount int
	eps     map[string]map[int]chan Event
	rm      sync.RWMutex
}

// New return Bus struct pointer.
func New() *Bus {
	return &Bus{
		epcount: 0,
		eps:     make(map[string]map[int]chan Event),
		rm:      sync.RWMutex{},
	}
}

// Attach return Endpoint to receive events from the specified topic.
func (eb *Bus) Attach(topic string) Endpoint {
	eb.rm.Lock()
	ep := eb.Newendpoint(topic + " " + strconv.Itoa(eb.epcount))
	if _, found := eb.eps[topic]; !found {
		eb.eps[topic] = make(map[int]chan Event)
	}
	eb.eps[topic][eb.epcount] = ep.Ch
	eb.epcount++
	eb.rm.Unlock()
	return ep
}

// Send send event from bus to target topic's receiver.
func (eb *Bus) Send(topic string, data Event) {
	eb.rm.RLock()
	if eps, found := eb.eps[topic]; found {
		go func(data Event, eps map[int]chan Event) {
			for _, ep := range eps {
				ep <- data.(Event)
			}
		}(data, eps)
	}
	eb.rm.RUnlock()
}

type endpoint struct {
	eb         Bus
	descriptor string
	Ch         chan Event
}

// Newendpoint return endpoint struct pointer.
func (eb *Bus) Newendpoint(descriptor string) *endpoint {
	return &endpoint{eb: *eb, descriptor: descriptor, Ch: make(chan Event, 100)}
}

// Send endpoint could also send event to other endpoint by bus.
func (ep *endpoint) Send(topic string, e Event) (err error) {
	//add time out control
	ep.eb.Send(topic, e)
	return nil
}

// Receice is a block method until receive a event from the eventbus.
func (ep *endpoint) Receive() (data Event, err error) {
	data, haddata := <-ep.Ch
	if !haddata {
		return data, errors.New("channel no data and closed")
	}
	return data, nil
}

// Detach detach current endpoint from bus.
func (ep *endpoint) Detach() { //	close channel on EventBus
	desc := strings.Split(ep.descriptor, " ")
	epcount, _ := strconv.Atoi(desc[1])
	delete(ep.eb.eps[desc[0]], epcount)
	close(ep.Ch)
}

// Probe detects the bus if have this endpoint. In case to know if avialable to send or receive Event.
func (ep *endpoint) Probe() bool {
	desc := strings.Split(ep.descriptor, " ")
	epcount, _ := strconv.Atoi(desc[1])
	if _, found := ep.eb.eps[desc[0]][epcount]; !found {
		return false
	}
	return true
}

// Descriptor return the info of the endpoint.
func (ep *endpoint) Descriptor() string { return ep.descriptor }
