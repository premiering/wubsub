package client

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/premiering/wubsub/message"
)

type WubSubConnection struct {
	builder *WubSubBuilder

	mtx      sync.Mutex
	outgoing []message.Message
}

func newConnection(builder *WubSubBuilder) WubSubConnection {
	return WubSubConnection{
		builder,
		sync.Mutex{},
		make([]message.Message, 0),
	}
}

func (w *WubSubConnection) Publish(channel string, data interface{}) {
	w.mtx.Lock()
	m := message.NewPublishMessage(channel, data)
	w.outgoing = append(w.outgoing, m)
	w.mtx.Unlock()
}

func (w *WubSubConnection) connect() {
	w.connectBlocking()
	w.builder.onInfo("disconnected from wubsub, waiting & retrying")
	time.Sleep(time.Duration(w.builder.reconnectTimeMs) * time.Millisecond)
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

	interrupt := make(chan os.Signal, 1)
	done := make(chan struct{})

	defer conn.Close()

	go func() {
		defer close(done)
		// queue subscribes / registers to our channels
		// sent manually since we're not in the read loop yet
		w.mtx.Lock()
		for _, cha := range w.builder.subscribed {
			s := message.NewSubscribeMessage(cha)
			w.outgoing = append(w.outgoing, s)
		}
		for _, cha := range w.builder.registered {
			r := message.NewRegisterMessage(cha)
			w.outgoing = append(w.outgoing, r)
		}
		w.mtx.Unlock()
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
		default:
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
		}
	}
}
