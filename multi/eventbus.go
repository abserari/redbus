package eventbus

import (
	"bufio"
	"errors"
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

// Endpoint use for receive and send events
type Endpoint interface {
	Descriptor() string
	Receive() (Event, error)
	Send(topic string, payload Event) error
	Detach()
}

// Bus use to get endpoint and send events to other endpoint
type Bus interface {
	Attach(topic string) Endpoint
	Send(topic string, e Event)
}

// Event is a interface which realize yourself event to dispatch
type Event interface {
	ID() string
	Type() string
	Payload() []byte
}

// Pong -
type Pong struct {
	uid     string
	pID     string
	payload []byte
}

// NewPong -
func NewPong() Pong { return Pong{uid: "asdf", pID: "pong", payload: []byte("pong")} }

// ID -
func (p Pong) ID() string { return p.uid }

// Type -
func (p Pong) Type() string { return p.pID }

// Payload -
func (p Pong) Payload() []byte { return p.payload }

// EventBus realize bus to dispatch events.
type EventBus struct {
	epcount int
	eps     map[string]map[int]net.Conn
	rm      sync.RWMutex

	// use absolute path like /etc/redbus/bus.sock
	filepath string
}

// New -
func New() *EventBus {
	return &EventBus{
		epcount:  0,
		eps:      make(map[string]map[int]net.Conn),
		rm:       sync.RWMutex{},
		filepath: "test.sock",
	}
}

// Serve - multiEventbus serve at once to communicate events.
func (eb *EventBus) Serve() {
start:
	lis, err := net.Listen("unix", eb.filepath)
	if err != nil {
		log.Println("UNIX Domain Socket failed to listen:", err)
		err = os.Remove(eb.filepath)
		if err != nil {
			log.Fatalln("Faild to remove unix domain socket file:", err)
		}
		goto start
	} else {
		fmt.Println("Listening on", eb.filepath)
	}
	defer lis.Close()

	for {
		// listen conn
		conn, err := lis.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		// eb get conn.
		go eb.handle(conn)
	}
}

func (eb *EventBus) handle(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err == io.EOF { // read over
			fmt.Println("client exit with io.EOF")
			//remove eb.eps.conn
			break
		} else if err != nil { // other err
			log.Println(err)
			break
		}
		log.Print("receive: " + msg)

		args := strings.Split(msg, " ")
		switch args[0] {
		case "/attach":
			{
				//attach the conn, return number which to desc endpoint self
				conn.Write([]byte("/number " + strconv.Itoa(eb.epcount) + " \n"))

				// ensure have not nil map
				if _, found := eb.eps[args[1]]; !found {
					eb.eps[args[1]] = make(map[int]net.Conn)
				}
				eb.eps[args[1]][eb.epcount] = conn
				// flag to distinguish endpoint to one topic
				eb.epcount++
			}
		case "/pub":
			{
				//pub msg to target topic
				eb.Send(args[1], Pong{uid: "asdf", pID: "data", payload: []byte(args[2])})
			}
		case "/exit":
			{
				i, _ := strconv.Atoi(args[1])
				delete(eb.eps[args[1]], i)
				// should find other way to distinguish ep like a map[int]bool
				// eb.epcount--
				log.Println("delete", i, eb.epcount)
			}
		default:
		}

	}
}

// Attach get a endpoint to receive related topic's events.
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

// Send is eb's method to send message to target topic with data.
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

// endpoint -
type endpoint struct {
	conn       net.Conn
	rd         *bufio.Reader
	descriptor string
	ch         chan Event
}

// Newendpoint return a ep. use by eb attach.
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
	if len(args) < 2 {
		return nil, errors.New("error event received")
	}
	// todo: according to message type to return
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

	data = Pong{uid: "asdf", pID: args[0], payload: []byte(args[1])}
	return data, nil
}

func (ep *endpoint) Detach() { //	close channel on EventBus
	ep.conn.Write([]byte("/exit " + ep.Descriptor() + " \n"))
	ep.conn.Close()
}

func (ep *endpoint) Descriptor() string { return ep.descriptor }
