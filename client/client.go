package client

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/premiering/wubsub/message"
)

var outgoingWaitInterval = time.Millisecond * 5

type WubSubConnection struct {
	builder *WubSubBuilder

	mtx             sync.Mutex
	outgoing        []message.Message
	processMessages chan bool

	connected bool
}

func newConnection(builder *WubSubBuilder) WubSubConnection {
	mq := make([]message.Message, 0)
	queueSubRegMessages(builder, &mq)

	return WubSubConnection{
		builder,
		sync.Mutex{},
		mq,
		make(chan bool),
		false,
	}
}

func (w *WubSubConnection) Publish(channel string, data interface{}) {
	w.mtx.Lock()
	m := message.NewPublishMessage(channel, data)
	w.outgoing = append(w.outgoing, m)
	if w.connected {
		w.processMessages <- true
	}
	w.mtx.Unlock()
}

func (w *WubSubConnection) connect() {
	w.connectBlocking()
	w.builder.onInfo("disconnected from wubsub, waiting & retrying")
	time.Sleep(time.Duration(w.builder.reconnectTimeMs) * time.Millisecond)

	w.mtx.Lock()
	queueSubRegMessages(w.builder, &w.outgoing)
	w.mtx.Unlock()

	w.connect()
}

func (w *WubSubConnection) connectBlocking() {
	w.builder.onInfo("attempting to connect to wubsub")
	conn, _, err := websocket.DefaultDialer.Dial(w.builder.address.String(), nil)
	if err != nil {
		w.builder.onError(w, err)
		return
	}

	w.builder.onInfo("connected to wubsub")
	w.connected = true

	interrupt := make(chan os.Signal, 1)
	done := make(chan struct{})

	defer conn.Close()

	go func() {
		defer close(done)
		w.processMessages <- true
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				w.builder.onError(w, err)
				return
			}
			// handle the message we read in
			var message message.Message
			json.Unmarshal(data, &message)
			w.builder.handleMessage(w, message)
		}
	}()

	for {
		select {
		case <-done:
			w.mtx.Lock()
			w.connected = false
			w.mtx.Unlock()
			return
		case <-interrupt:
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		case <-w.processMessages:
			w.mtx.Lock()
			n := len(w.outgoing)
			if n == 0 {
				w.mtx.Unlock()
				continue
			}
			copied := make([]message.Message, n)
			copy(copied, w.outgoing)
			w.outgoing = make([]message.Message, 0)
			w.mtx.Unlock()

			for _, m := range copied {
				err := conn.WriteJSON(m)
				if err != nil {
					w.builder.onError(w, err)
				}
			}
			time.Sleep(outgoingWaitInterval)
		}
	}
}

func queueSubRegMessages(builder *WubSubBuilder, queue *[]message.Message) {
	for _, cha := range builder.subscribed {
		s := message.NewSubscribeMessage(cha)
		*queue = append(*queue, s)
	}
	for _, cha := range builder.registered {
		r := message.NewRegisterMessage(cha)
		*queue = append(*queue, r)
	}
}
