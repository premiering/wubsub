# wubsub
small, <250 line WebSocket pubsub server

one publisher, multi subscriber

probably best for lightweight usages like ipc rpc or such

## running the server
`go run main.go`

need tls?
`go run main.go --tls --keyfile server.key --certfile server.cert`

want to some debug info too?
`go run main.go --tls --debug --keyfile server.key --certfile server.cert`

docker is coming soon

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
    "data": {
        "message": "You aren't the publisher of this channel!"
    }
}
```

that simple