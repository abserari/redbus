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

type Event struct{}
type Redbus struct {
	channels map[string][]chan Event
	rm       sync.RWMutex
	redis    *redis.Client
}
