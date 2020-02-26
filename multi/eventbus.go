package eventbus

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)
//https://github.com/grpc/grpc-go/blob/master/stream.go#L122:2
//stream read write
type Endpoint interface {
	Descriptor()  string
	Receive() (Event , error)
	Detach()
	Probe() bool
}

//type EventBus interface {
//	Attach(topic string) Endpoint
//	Send(topic string, e Event)
//}

type Event interface {
	Type() string
	Payload() []byte
}

type Pong struct {
	pID string
	payload []byte
}
func NewPong() Pong { return Pong{pID: "pong",payload: []byte("pong")}}
func (p Pong) Type() string { return p.pID }
func (p Pong) Payload() []byte {return p.payload}

type EventBus struct {
	epcount int
	eps map[string]map[int]chan Event
	rm    sync.RWMutex

	filename string
}

func (eb *EventBus) Serve(){
	lis, err := net.Listen("unix", eb.filename)
	if err != nil {
		fmt.Println(err)
	}
	defer lis.Close()

	for	{
		conn, err:= lis.Accept()


		in handle(conn)
		{
			defer conn.Close()
			for {
				reader := bufio.NewReader(conn)
				msg, err := reader.ReadString('\n')
				if err == io.EOF { //当对端退出后会报这么一个错误
					fmt.Println("go : 对端已接 收全部数据")
					remove eb.eps.conn
					break
				} else if err != nil { //处理完客户端关闭的错误正常错误还是要处理的
					log.Println(err)
					break
				}
				
				if attach {
					attach
					the
					conn
				}
				if pub {
					conn.Write([]byte("服务端已接收数\n"))
					pub
					msg
					to
					conn
				}
			}
		}
	}
}

func New() *EventBus{ return &EventBus{
	epcount: 0,
	eps: make(map[string]map[int]chan Event),
	rm: sync.RWMutex{},
}}

func (eb *EventBus) Attach(topic string) Endpoint {
	eb.rm.Lock()
	ep := eb.Newendpoint(topic + " "+ strconv.Itoa(eb.epcount))
	if _, found := eb.eps[topic]; !found {
		eb.eps[topic] = make(map[int]chan Event)
	}
	eb.eps[topic][eb.epcount] =  ep.ch
	eb.epcount ++
	eb.rm.Unlock()
	return ep
}

func (eb *EventBus) Send(topic string, data Event) {
	eb.rm.RLock()
	if eps, found := eb.eps[topic]; found {
		go func(data Event, eps map[int]chan Event ) {
			for _, ep := range eps {
				ep <- data.(Event)
			}
		}(data, eps)
	}
	eb.rm.RUnlock()
}

type endpoint struct {
	eb EventBus
	descriptor string
	ch         chan Event
}

func (eb *EventBus)Newendpoint(descriptor string)  *endpoint {
	return &endpoint{eb: *eb,descriptor:descriptor,ch: make(chan Event,100)}
}

func (ep *endpoint) Send(topic string, e Event) ( err error) {
	//add time out control
	ep.eb.Send(topic,e)
	return nil
}

// a block method until receive a event from the eventbus
func (ep *endpoint) Receive()(data Event, err error) {
	data, haddata := <- ep.ch
	if !haddata {
		return data,errors.New("channel no data and closed")
	}
	return data ,nil
}

func (ep *endpoint) Detach() { //	close channel on EventBus
	desc := strings.Split(ep.descriptor," ")
	epcount,_ := strconv.Atoi(desc[1])
	delete(ep.eb.eps[desc[0]],epcount)
	close(ep.ch)
}

func (ep *endpoint) Probe()bool {
	desc := strings.Split(ep.descriptor," ")
	epcount,_ := strconv.Atoi(desc[1])
	if _,found := ep.eb.eps[desc[0]][epcount]; !found {
		return false
	}
	return true
}

func(ep *endpoint) Descriptor() string { return ep.descriptor}