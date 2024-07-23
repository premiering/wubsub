package client

import (
	"errors"
	"net/url"

	"github.com/premiering/wubsub/log"
	"github.com/premiering/wubsub/message"
)

type InfoHandler func(info string)
type ReceiveHandler func(w *WubSubConnection, data interface{})
type ErrorHandler func(w *WubSubConnection, err error)

// Used to handle storing and functionality of wubsub client options
type WubSubBuilder struct {
	address url.URL

	subscribed []string
	registered []string

	onInfo    InfoHandler
	onError   ErrorHandler
	onReceive map[string][]ReceiveHandler

	reconnectTimeMs int
}

// Creates the builder for a new wubsub client
func NewBuilder(address url.URL) *WubSubBuilder {
	log.SetAppName("wubsub-client")
	return &WubSubBuilder{
		address,
		make([]string, 0),
		make([]string, 0),
		func(s string) {
			log.InfoLog(s)
		},
		func(w *WubSubConnection, err error) {
			log.ErrorLog("%s", err.Error())
		},
		make(map[string][]ReceiveHandler),
		2500,
	}
}

func (w *WubSubBuilder) Register(channel string) *WubSubBuilder {
	w.registered = append(w.registered, channel)
	return w
}

func (w *WubSubBuilder) Subscribe(channel string) *WubSubBuilder {
	w.subscribed = append(w.subscribed, channel)
	return w
}

func (w *WubSubBuilder) Connect() *WubSubConnection {
	con := newConnection(w)
	go con.connect()
	return &con
}

func (w *WubSubBuilder) OnReceive(channel string, r ReceiveHandler) *WubSubBuilder {
	arr, exists := w.onReceive[channel]
	if !exists {
		arr = make([]ReceiveHandler, 1)
		arr[0] = r
	} else {
		arr = append(arr, r)
	}
	w.onReceive[channel] = arr
	return w
}

func (w *WubSubBuilder) SetLogHandler(l InfoHandler) *WubSubBuilder {
	w.onInfo = l
	return w
}

func (w *WubSubBuilder) SetErrorHandler(e ErrorHandler) *WubSubBuilder {
	w.onError = e
	return w
}

func (w *WubSubBuilder) SetReconnectTime(ms int) *WubSubBuilder {
	w.reconnectTimeMs = ms
	return w
}

func (w *WubSubBuilder) handleMessage(wc *WubSubConnection, m message.Message) {
	switch m.Type {
	case message.Receive:
		handlers, exists := w.onReceive[m.Channel]
		if !exists {
			return
		}
		for _, handler := range handlers {
			handler(wc, m.Data)
		}
	case message.Error:
		w.onError(wc, errors.New(m.Data.(string)))
	}
}
