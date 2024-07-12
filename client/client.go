package client

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/premiering/wubsub/message"
)

type ReceiveHandler func(w *WubSubClient, data interface{})

type ErrorHandler func(w *WubSubClient, err error)

type AddOnReceive struct {
	channel string
	handler ReceiveHandler
}

type WubSubClient struct {
	address url.URL

	conn *websocket.Conn

	registeredChans   []string
	addRegisteredChan chan string

	onReceive    map[string]ReceiveHandler
	addOnReceive chan AddOnReceive

	onError    []ErrorHandler
	addOnError chan ErrorHandler

	incoming chan message.Message
	outgoing chan message.Message

	closed chan struct{}
}

func NewWubSubClient(address url.URL) WubSubClient {
	return WubSubClient{
		address,
		nil,
		make([]string, 0),
		make(chan string),
		make(map[string]ReceiveHandler),
		make(chan AddOnReceive),
		make([]ErrorHandler, 0),
		make(chan ErrorHandler),
		make(chan message.Message),
		make(chan message.Message),
		make(chan struct{}),
	}
}

// Registers a channel in the wubsub to be able to publish to it
func (w *WubSubClient) Register(channel string) {
	// TODO: already registered checking
	w.addRegisteredChan <- channel
	r := message.NewRegisterMessage(channel)
	w.outgoing <- r
}

// Publishes data to the specified channel
func (w *WubSubClient) Publish(channel string, data interface{}) {
	p := message.NewPublishMessage(channel, data)
	w.outgoing <- p
}

func (w *WubSubClient) Subscribe(channel string) {
	s := message.NewSubscribeMessage(channel)
	w.outgoing <- s
}

// Registers a function to handle publishes from a channel
func (w *WubSubClient) Receive(channel string, handler ReceiveHandler) {
	w.addOnReceive <- AddOnReceive{channel, handler}
}

// Adds a function that is called when an error is received from the wubsub
func (w *WubSubClient) Error(handler ErrorHandler) {
	w.addOnError <- handler
}

// Connects blockingly to the wubsub server
func (w *WubSubClient) Connect() error {
	c, _, err := websocket.DefaultDialer.Dial(w.address.String(), nil)
	if err != nil {
		return err
	}
	w.conn = c

	go w.handleConnection()
	go w.processChannels()

	return nil
}

func (w *WubSubClient) handleConnection() {
	defer close(w.closed)
	for {
		_, data, err := w.conn.ReadMessage()
		if err != nil {
			return
		}
		var message message.Message
		json.Unmarshal(data, &message)
		w.incoming <- message
	}
}

func (w *WubSubClient) processChannels() {
	defer w.conn.Close()

	for {
		select {
		case <-w.closed:
			return
		case c := <-w.addRegisteredChan:
			w.registeredChans = append(w.registeredChans, c)
		case a := <-w.addOnReceive:
			w.onReceive[a.channel] = a.handler
		case e := <-w.addOnError:
			w.onError = append(w.onError, e)
		case i := <-w.incoming:
			w.handleMessage(i)
		case o := <-w.outgoing:
			w.conn.WriteJSON(o)
		}
	}
}

func (w *WubSubClient) handleMessage(m message.Message) {
	switch m.Type {
	case message.Receive:
		handler := w.onReceive[m.Channel]
		if handler != nil {
			handler(w, m.Data)
		}
		break
	case message.Error:
		for _, handler := range w.onError {
			handler(w, errors.New(m.Data.(string)))
		}
	}
}
