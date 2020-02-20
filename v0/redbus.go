package redbus

import (
	"fmt"
	"sync"

	"github.com/go-redis/redis/v7"
)
//https://github.com/grpc/grpc-go/blob/master/stream.go#L122:2
//stream read write
type Endpoint interface {
	Write(data Event) error
	Read(data *[]Event) error
	Dettach()
	Proble() bool
}

type EventBus interface {
	Attach() Endpoint
}


type Event struct {
	PID string
}

func (p *Event) Type() string { return p.PID }

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
				_ = ep.Write(data)
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

func NewEndpoint(descriptor string)  *endpoint {
	return &endpoint{Descriptor:descriptor,ch: make(chan Event,100),done: make(chan struct{})}
}

func (ep *endpoint) Write(e Event) (err error) {
	//add time out control
	ep.ch <- e
	return nil
}

func (ep *endpoint) Read(data *[]Event) (err error) {
	*data = append(*data,  <- ep.ch)
	fmt.Println(data)
	return nil
}

func (ep *endpoint) Dettach() { //	close channel on EventBus
	ep.done <- struct{}{}
	close(ep.ch)
}

func (ep *endpoint) Proble()bool {
	if len(ep.done) ==0 {
		return true
	}
	return false
}
