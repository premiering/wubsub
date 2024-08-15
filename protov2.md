## Protocol
C->S indicates Client to Server.

S->C indicates Server to Client.

### General commands
#### C->S Creating a channel
```json
{
    "type": "register",
    "channel": "sample_channel",
    "data": {
        "name": "sample_channel",
        "description": "Assistant chatting service",
        "pubsub": {
            "topics": [
                "responses"
            ]
        },
        "restpie": {
            "methods": [
                "status",
                "get-opts",
                "get-opts-schema"
            ]
        }
    }
}
```

#### S->C Global, error is received from server due to receive client's action
```json
{
    "type": "error",
    "data": "Tried to publish to channel you aren't a publisher for."
}
```

### Pub/Sub commands
#### C->S pubsub owner publishes data
```json
{
    "type": "ps-pub",
    "channel": "sample_channel",
    "topic": "responses",
    "data": {
        "text": "Why is the sky blue?"
    }
}
```

#### C->S general user subscribes to pubsub
```json
{
    "type": "ps-sub",
    "channel": "sample_channel",
    "topic": "responses"
}
```

#### S->C general user receives data from topic they subscribed to
```json
{
    "type": "ps-recv",
    "channel": "sample_channel",
    "topic": "responses",
    "data": {
        "text": "Why is the sky blue?"
    }
}
```

### restpie commands
The following restpie commands are shown with an example of the channel `sample_channel`, which has a restpie method/topic named `status`.

General user refers to the client that intiates the request.

Channel owner is the owner of the `sample_channel` channel.
#### C->S general user initiates request
```json
{
    "type": "pie-do",
    "channel": "sample_channel",
    "topic": "status"
}
```

#### S->C server confirms latest sent restpie request with ID with general user
This is done to ensure ordering and to know whom a packet should be sent to when receiving the response from the publisher of the restpie

```json
{
    "type": "pie-ack",
    "channel": "ollama",
    "topic": "status",
    "reqid": "3642af4c-1bde-4dc8-9833-e4d2104c33f2"
}
```

#### S->C channel owner receives request
```json
{
    "type": "pie-req",
    "channel": "sample_channel",
    "topic": "status",
    "reqid": "3642af4c-1bde-4dc8-9833-e4d2104c33f2"
}
```

#### C->S channel owner responds to the request
```json
{
    "type": "pie-resp",
    "channel": "sample_channel",
    "topic": "status",
    "reqid": "3642af4c-1bde-4dc8-9833-e4d2104c33f2",
    "data": {
        "text": "Ok!",
        "cpu_load": "5.2%"
    }
}
```

Server will then receive this request and send the data to the general user

#### S->C general user receives response from request
```json
{
    "type": "pie-recv",
    "channel": "sample_channel",
    "reqid": "3642af4c-1bde-4dc8-9833-e4d2104c33f2",
    "topic": "status",
    "data": {
        "text": "Ok!",
        "cpu_load": "5.2%"
    }
}
```

The request is then considered complete.