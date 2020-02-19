package eventbus

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis/v7"
)

type EventBus interface {
	Subscribe(channel string) chan *redis.Message
	Publish(channel string, msg ...interface{})
	Receive(channel string) *redis.Message
}

type Redbus struct {
	red    *redis.Client
	subBus map[string]*redis.PubSub
}

func NewDefault() *Redbus {
	return &Redbus{red: redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}), subBus: make(map[string]*redis.PubSub)}
}

func (b *Redbus) Receive(channel string) (interface{}, error) {
	if b.subBus[channel] == nil {
		return nil, errors.New("no channel")
	}

	return b.subBus[channel].Receive()
}

func (b *Redbus) Subscribe(channel string) <-chan *redis.Message {
	pubsub := b.red.Subscribe(channel)
	iface, err := pubsub.Receive()
	if err != nil {
		return nil
	}
	switch iface.(type) {
	case *redis.Subscription:
		// subscribe succeeded
	case *redis.Message:
		// received first message
	case *redis.Pong:
		// pong received
	default:
		// handle error
	}

	b.subBus[channel] = pubsub

	return pubsub.Channel()
}

func (b *Redbus) UnSubscribe(channels ...string) error {
	for _, channel := range channels {
		if b.subBus[channel] == nil {
			continue
		}
		err := b.subBus[channel].Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Close closes the client, releasing any open resources.
//
// It is rare to Close a Client, as the Client is meant to be
// long-lived and shared between many goroutines.
// don't need because of pubsub is in pool of red
// for _, subC := range b.subC {
// 	subC.Close()
// }
func (b *Redbus) Close() error {
	return b.red.Close()
}

//------------------------------------------------------------------------------
// Publish posts the message to the channel.
// http://redisdoc.com/pubsub/pubsub.html
func (b *Redbus) Broadcast(msg ...interface{}) {
	b.PPublish("", msg...)
}

func (b *Redbus) PPublish(pattern string, msg ...interface{}) {
	channels := b.red.PubSubChannels(pattern).Val()
	for _, channel := range channels {
		b.red.Publish(channel, fmt.Sprintln(msg...))
	}
}

func (b *Redbus) Publish(channel string, msg ...interface{}) {
	b.red.Publish(channel, fmt.Sprint(msg...))
}
