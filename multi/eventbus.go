package eventbus

type Endpoint interface {
	Descriptor()  string
	Receive() (Event , error)
	Detach()
	Probe() bool
}

type LocalEventBus interface {
	Attach() Endpoint
	Send(topic string, e Event)
}

type Event interface {
	Type() string
	Payload() []byte
}