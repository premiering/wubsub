package app

// This is the message format for client -> server and server -> client messages
type Message struct {
	Type    string      `json:"type"`
	Channel string      `json:"channel"`
	Data    interface{} `json:"data"`
}

func NewErrorMessage(text string) Message {
	data := make(map[string]interface{})
	data["message"] = text
	return Message{"error", "", data}
}

func NewReceiveMessage(channel string, data interface{}) Message {
	return Message{
		"receive",
		channel,
		data,
	}
}
