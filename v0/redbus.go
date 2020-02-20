package redbus

import (
	"sync"

	"github.com/go-redis/redis/v7"
)

type Endpoint interface {
	Read()
	Write()
	Dettach()
}

type EventBus interface {
	Attach() Endpoint
	Write()
	Read()
}

type Event interface {
	Type() string
}
type Data struct{
	PID string
}
func (p *Data) Type() {return p.PID }

type Setup struct{}
func (p *Setup) Type() {return p.PID}

type ACK struct{}
func (p *ACK) Type() {return p.PID}

type Redbus struct {
	channels map[string][]chan Event
	rm       sync.RWMutex
	redis    *redis.Client
}

func (eb *EventBus) publish(topic string, data interface{}) {
	eb.rm.RLock()
	if chans, found := eb.channels[topic]; found {
		channels := append(map[string][]chan Event{}, chans...)
		go func(data Event, channels map[string][]chan Event) {
			for _, ch := range channels {
				ch <- data
			}
		}(data, channels)
	}
	eb.rm.RUnlock()
}

func (eb *EventBus)subscribe(topic string, ch []chan Event])  {
	eb.rm.Lock()
	if prev, found := eb.channels[topic]; found {
	   eb.channels[topic] = append(prev, ch)
	} else {
	   eb.channels[topic] = append([]chan Event{}, ch)
	}
	eb.rm.Unlock()
 }
 