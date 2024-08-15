package message

// This is the message format for client -> server and server -> client messages
type Message struct {
	Type      string      `json:"type"`
	Channel   string      `json:"channel,omitempty"`
	Topic     string      `json:"topic,omitempty"`
	RequestId string      `json:"reqid,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

// general
const Register string = "register"
const Error string = "error"

// pubsub
const PSPublish string = "ps-pub"
const PSSubscribe string = "ps-sub"
const PSReceive string = "ps-recv"

// restpie
const RPRequestInitiate string = "pie-do"
const RPAcknowledge string = "pie-ack"
const RPRequestReceived string = "pie-req"
const RPResponding string = "pie-resp"
const RPReceiveResponse string = "pie-recv"

func NewErrorMsg(text string) Message {
	return Message{Error, "", "", "", text}
}

func NewPSReceiveMsg(channel string, topic string, data interface{}) Message {
	return Message{PSReceive, channel, topic, "", data}
}

func NewRPAckMsg(channel string, topic string, good bool, reqid string) Message {
	return Message{RPAcknowledge, channel, topic, reqid, good}
}

func NewRPReqMsg(channel string, topic string, reqid string, data interface{}) Message {
	return Message{RPRequestReceived, channel, topic, reqid, data}
}

func NewRPRecvMsg(channel string, topic string, reqid string, data interface{}) Message {
	return Message{RPReceiveResponse, channel, topic, reqid, data}
}
