package message

// This is the message format for client -> server and server -> client messages
type Message struct {
	Type    string      `json:"type"`
	Channel string      `json:"channel"`
	Data    interface{} `json:"data"`
}

const Register string = "register"
const Publish string = "publish"
const Subscribe string = "subscribe"
const Receive string = "receive"
const Error string = "error"

// Server-side
func NewErrorMessage(text string) Message {
	return Message{Error, "", text}
}

func NewReceiveMessage(channel string, data interface{}) Message {
	return Message{Receive, channel, data}
}

// Client-side
func NewPublishMessage(channel string, data interface{}) Message {
	return Message{Publish, channel, data}
}

func NewRegisterMessage(channel string) Message {
	return Message{Register, channel, nil}
}

func NewSubscribeMessage(channel string) Message {
	return Message{Subscribe, channel, nil}
}
