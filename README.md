# wubsub
small, <400 line WebSocket pubsub client/server

one publisher, multi subscriber

probably best for lightweight usages like ipc rpc or such

## running the server
```
go run .
````

need tls?
```bash
go run . --tls --keyfile server.key --certfile server.cert
```

want to some debug info too?
```
go run . --tls --debug --keyfile server.key --certfile server.cert
```

docker is coming soon

## built-in go client
wubsub has a client implementation in the client package

it is not on Go Packages, so you'll have to clone this repo to use it

example of setting up the package locally:
```bash
# In a parent folder containing a folder to your Go project
git clone https://github.com/premiering/wubsub.git

# In your Go project's root folder
go mod edit -replace github.com/premiering/wubsub=../wubsub
```

example of publishing with the client:
```go
import (
    "fmt"
    "net/url"
    "time"

    "github.com/premiering/wubsub/client"
)

func main() {
    builder := client.NewBuilder(url.URL{Scheme: "ws", Host: "localhost:9190", Path: "/"})
    builder.Register("cool_channel")
    client := builder.Connect()
    client.Publish("cool_channel", "Isn't this great!?")
    time.Sleep(time.Second * 10)
}
```

example of subscribing with the client:
```go
import (
    "fmt"
    "net/url"
    "time"

    "github.com/premiering/wubsub/client"
)

func main() {
    builder := client.NewBuilder(url.URL{Scheme: "ws", Host: "localhost:9190", Path: "/"})
    builder.Subscribe("cool_channel")
    builder.OnReceive("cool_channel", func(w *client.WubSubConnection, data interface{}) {
        m := data.(string)
        fmt.Println(m)
        // Output: "Isn't this great!?"
    })
    builder.Connect()
    time.Sleep(time.Second * 10)
}
```

## client specifications
all communication is in pure JSON for convenience

to subscribe to a channel, send this JSON packet:
```json
{
    "type": "subscribe",
    "channel": "cool_channel_name"
}
```

when you receive a message, it'll be like this
```json
{
    "type": "receive",
    "channel": "cool_channel_name",
    "data": {
        "some_key": "a great value"
    }
}
```

if you want to start publishing to channels, you first have to register yourself

to register, send this packet:
```json
{
    "type": "register",
    "channel": "my_cool_channel"
}
```

then to publish data:

```json
{
    "type": "publish",
    "channel": "my_cool_channel",
    "data": {
        "some_key": "a great value"
    }
}
```

if you or wubsub does something wrong, you'll receive an error message back like this:
```json
{
    "type": "error",
    "channel": "",
    "data": "You aren't the publisher of this channel!"
}
```

that simple
