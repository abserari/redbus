# Redbus

a eventbus which attempt to use redis pub/sub to communicate.

## Usage
```go
import "github.com/yhyddr/redbus"

func add(e *Event) {
   a := strconv.Atoi(e.Get("a"))
   b := strconv.Atoi(e.Get("b"))
   fmt.Println(a+b)
}

func main() {
    redbus.HandleFunc("/add", add)
    go redbus.Serve()
    msg := make(http.Header)
    msg.Add("a", "1")
    msg.Add("b", "2")
    redbus.Publish("/add",msg)
}
```
