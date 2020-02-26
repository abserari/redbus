package eventbus

import (
	"bufio"
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
	Send(topic string, payload Event) error
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
	eps map[string]map[int]net.Conn
	rm    sync.RWMutex

	filepath string
}

func (eb *EventBus) Serve(){
	lis, err := net.Listen("unix", eb.filepath)
	if err != nil {
		fmt.Println(err)
	}
	defer lis.Close()

	for	{
		conn, err:= lis.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go eb.handle(conn)
	}
}

func (eb *EventBus)handle(conn net.Conn) {
	defer conn.Close()

	for {
		reader := bufio.NewReader(conn)
		msg, err := reader.ReadString('\n')
		if err == io.EOF { //当对端退出后会报这么一个错误
			fmt.Println("go : 对端已接 收全部数据")
			//remove eb.eps.conn
			break
		} else if err != nil { //处理完客户端关闭的错误正常错误还是要处理的
			log.Println(err)
			break
		}

		args := strings.Split(msg," ")

		switch args[0] {
		case "/attach" :{
			//attach the conn
			conn.Write([]byte("/number " + strconv.Itoa(eb.epcount)+ "\n") )
			if _, found := eb.eps[args[0]] ; !found {
				eb.eps[args[0]] = make(map[int]net.Conn)
			}
			eb.eps[args[1]][eb.epcount] = conn
			eb.epcount++
		}
		case "/pub" : {
			conn.Write([]byte("服务端已接收数\n"))
			eb.Send(args[1], Pong{pID:"data", payload: []byte(args[2])})
			//pub msg to conn
		}
		case "/exit" : {
			i, _ := strconv.Atoi(args[2])
			delete(eb.eps[args[1]],i)
		}
		}

	}
}

func New() *EventBus{ return &EventBus{
	epcount: 0,
	eps: make(map[string]map[int]net.Conn),
	rm: sync.RWMutex{},
	filepath: "test.sock",
}}

func (eb *EventBus) Attach(topic string) Endpoint {
	eb.rm.Lock()
	conn , err := net.Dial("unix",eb.filepath)
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Write([]byte("/attach " + topic + "\n"))
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	args := strings.Split(msg," ")
	if args[0] == "/number" {
		// new endpoint
		ep := eb.Newendpoint(args[1],conn)

		// response to register info on eb
		_, err = conn.Write([]byte("/info " + ep.Descriptor() + "\n"))
		if err != nil {
			log.Fatal(err)
		}

		eb.rm.Unlock()
		return ep
	}
	return nil
}

func (eb *EventBus) Send(topic string, data Event) {
	eb.rm.RLock()
	if eps, found := eb.eps[topic]; found {
		go func(topic string , data Event, eps map[int]net.Conn ) {
			for _, ep := range eps {
				var sendmsg = []byte("/pub " + topic + " " + string(data.Payload()) )
				ep.Write(sendmsg)
			}
		}(topic, data, eps)
	}
	eb.rm.RUnlock()
}

type endpoint struct {
	conn net.Conn
	descriptor string
	ch         chan Event
}

func (eb *EventBus)Newendpoint(descriptor string,conn net.Conn)  *endpoint {
	return &endpoint{conn: conn,descriptor:descriptor,ch: make(chan Event,100)}
}

func (ep *endpoint) Send(topic string, e Event) ( err error) {
	//add time out control
	_, err = ep.conn.Write([]byte("/pub " + topic +" " + string(e.Payload())))
	return
}

// a block method until receive a event from the eventbus
func (ep *endpoint) Receive()(data Event, err error) {
	reader := bufio.NewReader(ep.conn)
	msg, err := reader.ReadString('\n')
	args := strings.Split(msg," ")
	data = Pong{pID:args[0],payload:[]byte(args[1])}
	return data ,nil
}

func (ep *endpoint) Detach() { //	close channel on EventBus
	ep.conn.Write([]byte("/exit " + ep.Descriptor()))
	ep.conn.Close()
}

func (ep *endpoint) Probe()bool {
	ep.conn.Write([]byte("/exit " + ep.Descriptor()))
	reader := bufio.NewReader(ep.conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
		return false
	}
	args := strings.Split(msg," ")
	if args[0] == "/pong" {
		return true
	}
	return false
}

func(ep *endpoint) Descriptor() string { return ep.descriptor}