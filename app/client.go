package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	app *WubSubApp

	writer   http.ResponseWriter
	req      *http.Request
	upgrader *websocket.Upgrader
	conn     *websocket.Conn

	// the channels that this client publishes to
	publishesTo map[string]*Channel
}

func CreateWSClient(app *WubSubApp, writer http.ResponseWriter, req *http.Request, upgrader *websocket.Upgrader) WSClient {
	return WSClient{app, writer, req, upgrader, nil, make(map[string]*Channel, 0)}
}

func (w *WSClient) UpgradeAndHandle() {
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

		var message Message
		err = json.Unmarshal(data, &message)
		if err != nil {
			fmt.Println(err)
			continue
		}

		err = w.handleMessage(&message)
		if err != nil {
			e_message := NewErrorMessage(err.Error())
			w.Send(&e_message)
		}
	}
	w.app.disconnects <- Disconnect{
		w,
	}
}

func (w *WSClient) Send(message *Message) {
	w.conn.WriteJSON(message)
}

func (w *WSClient) handleMessage(message *Message) error {
	switch message.Type {
	case "publish":
		w.app.publishes <- Publish{w, message.Channel, message.Data}
		return nil
	case "subscribe":
		w.app.subscribes <- Subscribe{w, message.Channel}
		return nil
	case "register":
		w.app.registers <- Register{w, message.Channel}
		return nil
	}
	return errors.New("Invalid message type!")
}
