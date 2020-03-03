package eventbus

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

//https://github.com/grpc/grpc-go/blob/master/stream.go#L122:2
//stream read write
type Endpoint interface {
	Descriptor() string
	Receive() (Event, error)
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
	pID     string
	payload []byte
}

func NewPong() Pong            { return Pong{pID: "pong", payload: []byte("pong")} }
func (p Pong) Type() string    { return p.pID }
func (p Pong) Payload() []byte { return p.payload }

type EventBus struct {
	epcount int
	eps     map[string]map[int]net.Conn
	rm      sync.RWMutex

	filepath string
}

func (eb *EventBus) Serve() {
start:
	lis, err := net.Listen("unix", eb.filepath)
	if err != nil {
		log.Println("UNIX Domain Socket 创 建失败，正在尝试重新创建 -> ", err)
		err = os.Remove(eb.filepath)
		if err != nil {
			log.Fatalln("删除 sock 文件失败！程序退出 -> ", err)
		}
		goto start
	} else {
		fmt.Println("创建 UNIX Domain Socket 成功")
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go eb.handle(conn)
	}
}

func (eb *EventBus) handle(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err == io.EOF { //当对端退出后会报这么一个错误
			fmt.Println("对端已退出")
			//remove eb.eps.conn
			break
		} else if err != nil { //处理完客户端关闭的错误正常错误还是要处理的
			log.Println(err)
			break
		}
		log.Print("receive: " + msg)

		args := strings.Split(msg, " ")
		switch args[0] {
		case "/attach":
			{
				//attach the conn
				conn.Write([]byte("/number " + strconv.Itoa(eb.epcount) + " \n"))
				if _, found := eb.eps[args[1]]; !found {
					eb.eps[args[1]] = make(map[int]net.Conn)
				}
				eb.eps[args[1]][eb.epcount] = conn
				eb.epcount++
			}
		case "/pub":
			{
				eb.Send(args[1], Pong{pID: "data", payload: []byte(args[2])})
				//pub msg to conn
			}
		case "/exit":
			{
				i, _ := strconv.Atoi(args[1])
				delete(eb.eps[args[1]], i)
				eb.epcount--
				log.Println("delete", i, eb.epcount)
			}
		default:
		}

	}
}

func New() *EventBus {
	return &EventBus{
		epcount:  0,
		eps:      make(map[string]map[int]net.Conn),
		rm:       sync.RWMutex{},
		filepath: "test.sock",
	}
}

func (eb *EventBus) Attach(topic string) Endpoint {
	eb.rm.Lock()
	conn, err := net.Dial("unix", eb.filepath)
	if err != nil {
		log.Fatal(err)
	}
	_, err = conn.Write([]byte("/attach " + topic + " \n"))
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(conn)
	msg, err := reader.ReadString('\n')
	log.Print("get number", msg)
	if err != nil {
		log.Fatal(err)
	}
	args := strings.Split(msg, " ")
	if args[0] == "/number" {
		// new endpoint
		ep := eb.Newendpoint(args[1], conn)

		// response to register info on eb
		_, err = ep.conn.Write([]byte("/info " + ep.Descriptor() + " \n"))
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
	log.Println("send " + topic + " " + data.Type())
	if eps, found := eb.eps[topic]; found {
		go func(topic string, data Event, eps map[int]net.Conn) {
			for number, ep := range eps {
				var sendmsg = []byte("/pub " + topic + " " + string(data.Payload()) + " \n")
				log.Println("send to", number, topic, data.Payload())
				ep.Write(sendmsg)
			}
		}(topic, data, eps)
	}
	eb.rm.RUnlock()
}

type endpoint struct {
	conn       net.Conn
	rd         *bufio.Reader
	descriptor string
	ch         chan Event
}

func (eb *EventBus) Newendpoint(descriptor string, conn net.Conn) *endpoint {
	return &endpoint{conn: conn, rd: bufio.NewReader(conn), descriptor: descriptor, ch: make(chan Event, 100)}
}

func (ep *endpoint) Send(topic string, e Event) (err error) {
	//add time out control
	log.Println("send: " + topic + " " + e.Type())
	_, err = ep.conn.Write([]byte("/pub " + topic + " " + string(e.Payload()) + " \n"))
	if err != nil {
		log.Println(err)
	}
	return
}

// a block method until receive a event from the eventbus
func (ep *endpoint) Receive() (data Event, err error) {
	msg, err := ep.rd.ReadString('\n')
	log.Print("receive", msg)
	args := strings.Split(msg, " ")
	// switch iface.(type) {
	// case *redis.Subscription:
	// 	// subscribe succeeded
	// case *redis.Message:
	// 	// received first message
	// case *redis.Pong:
	// 	// pong received
	// default:
	// 	// handle error
	// }
	data = Pong{pID: args[0], payload: []byte(args[1])}
	return data, nil
}

func (ep *endpoint) Detach() { //	close channel on EventBus
	ep.conn.Write([]byte("/exit " + ep.Descriptor()))
	ep.conn.Close()
}

func (ep *endpoint) Probe() bool {
	ep.conn.Write([]byte("/ping " + ep.Descriptor()))
	reader := bufio.NewReader(ep.conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		log.Println(err)
		return false
	}
	args := strings.Split(msg, " ")
	if args[0] == "/pong" {
		return true
	}
	return false
}

func (ep *endpoint) Descriptor() string { return ep.descriptor }
