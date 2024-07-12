package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/premiering/wubsub/message"
)

type WSConnection struct {
	app *WubSubApp

	writer   http.ResponseWriter
	req      *http.Request
	upgrader *websocket.Upgrader
	conn     *websocket.Conn

	// the channels that this client publishes to
	publishesTo map[string]*Channel
}

func CreateWSClient(app *WubSubApp, writer http.ResponseWriter, req *http.Request, upgrader *websocket.Upgrader) WSConnection {
	return WSConnection{app, writer, req, upgrader, nil, make(map[string]*Channel, 0)}
}

func (w *WSConnection) UpgradeAndHandle() {
	c, err := w.upgrader.Upgrade(w.writer, w.req, nil)
	if err != nil {
		return
	}
	w.conn = c
	defer c.Close()
	for {
		_, data, err := c.ReadMessage()

		if err != nil {
			break
		}

		var m message.Message
		err = json.Unmarshal(data, &m)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = w.handleMessage(&m)
		if err != nil {
			e_message := message.NewErrorMessage(err.Error())
			w.Send(&e_message)
		}
	}
	w.app.disconnects <- Disconnect{
		w,
	}
}

func (w *WSConnection) Send(m *message.Message) {
	w.conn.WriteJSON(m)
}

func (w *WSConnection) handleMessage(m *message.Message) error {
	switch m.Type {
	case message.Publish:
		w.app.publishes <- Publish{w, m.Channel, m.Data}
		return nil
	case message.Subscribe:
		w.app.subscribes <- Subscribe{w, m.Channel}
		return nil
	case message.Register:
		w.app.registers <- Register{w, m.Channel}
		return nil
	}
	return errors.New("Invalid message type!")
}
