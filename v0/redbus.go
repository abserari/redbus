package redbus

import (
	"sync"

	"github.com/go-redis/redis/v7"
)
//https://github.com/grpc/grpc-go/blob/master/stream.go#L122:2
//stream read write
type Endpoint interface {
	Info() (descriptor string)
	Receive() (e Event ,err  error)
	Write(e Event) ( err error)
	Detach()
	Probe() bool
}

type EventBus interface {
	Attach() Endpoint
	Request(topic string)
}

type Event interface {
	Type() string
	Payload() string
}

type Packet struct {
	pID string
	payload string
}

func (p *Packet) Type() string { return p.pID }

func (p *Packet) Payload() string {return p.payload}

type Hub map[string][]Endpoint

type Redbus struct {
	hubs  Hub
	rm    sync.RWMutex
	redis *redis.Client
}

func New() *Redbus{ return &Redbus{
	hubs: make(map[string][]Endpoint),
	rm: sync.RWMutex{},
	redis: redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}),
}}

func (eb *Redbus) Attach(topic string) Endpoint {
	eb.rm.Lock()
	ep := NewEndpoint("device")
	if prev, found := eb.hubs[topic]; found {
		eb.hubs[topic] = append(prev, ep)
	} else {
		eb.hubs[topic] = append([]Endpoint{}, ep)
	}
	eb.rm.Unlock()
	return ep
}

func (eb *Redbus) Dispatch(topic string, data Event) {
	eb.rm.RLock()
	if eps, found := eb.hubs[topic]; found {
		endpoints := append([]Endpoint{}, eps...)
		go func(data Event, endpoints []Endpoint) {
			for _, ep := range endpoints {
				_  = ep.Write(data)
			}
		}(data, endpoints)
	}
	eb.rm.RUnlock()
}

type endpoint struct {
	Descriptor string
	ch         chan Event
	done       chan struct{}
}

func NewEndpoint(descriptor string)  Endpoint {
	return &endpoint{Descriptor:descriptor,ch: make(chan Event,100),done: make(chan struct{},1)}
}

func (ep *endpoint) Write(e Event) ( err error) {
	//add time out control
	ep.ch <- e
	return nil
}

// a block method until receive a event from the eventbus
func (ep *endpoint) Receive()(data Event, err error) {
	return <- ep.ch, nil
}

func (ep *endpoint) Detach() { //	close channel on EventBus
	ep.done <- struct{}{}
	close(ep.ch)
}

func (ep *endpoint) Probe()bool {
	if len(ep.done) ==0 {
		return true
	}
	return false
}

func(ep *endpoint) Info() string{ return ep.Descriptor}