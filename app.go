package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	wlog "github.com/premiering/wubsub/log"
	"github.com/premiering/wubsub/message"
)

type Register struct {
	client  *WSConnection
	channel string
}

type Publish struct {
	publisher *WSConnection
	channel   string
	data      interface{}
}

type Subscribe struct {
	subscriber *WSConnection
	channel    string
}

type Disconnect struct {
	client *WSConnection
}

type Channel struct {
	publisher   *WSConnection
	subscribers []*WSConnection
}

type WubSubApp struct {
	channels map[string]*Channel

	registers   chan Register
	publishes   chan Publish
	subscribes  chan Subscribe
	disconnects chan Disconnect

	upgrader websocket.Upgrader

	port  int
	debug bool

	tls bool
	// these are the paths
	keyfile  string
	certfile string
}

func CreateApp(port int, debug bool, tls bool, keyfile string, certfile string) WubSubApp {
	var upgrader = websocket.Upgrader{}
	return WubSubApp{
		make(map[string]*Channel),
		make(chan Register),
		make(chan Publish),
		make(chan Subscribe),
		make(chan Disconnect),
		upgrader,
		port,
		debug,
		tls,
		keyfile,
		certfile,
	}
}

func (app *WubSubApp) Start() {
	go app.processChannels()
	wlog.InfoLog("Started event processing channels")

	http.HandleFunc("/", app.handleConnection)
	port := strconv.Itoa(app.port)
	wlog.InfoLog("Listening on port :" + port)
	if app.tls {
		log.Fatal(http.ListenAndServeTLS(":"+port, app.certfile, app.keyfile, nil))
	} else {
		log.Fatal(http.ListenAndServe(":"+port, nil))
	}
}

func (app *WubSubApp) processChannels() {
	for {
		select {
		case disconnect := <-app.disconnects:
			app.handleDisconnect(&disconnect)
		case register := <-app.registers:
			app.handleRegister(&register)
		case publish := <-app.publishes:
			app.handlePublish(&publish)
		case subscribe := <-app.subscribes:
			app.handleSubscribe(&subscribe)
		}
	}
}

func (w *WubSubApp) handleDisconnect(d *Disconnect) {
	for _, channel := range d.client.publishesTo {
		// leave the channel in the map so that we still know
		// its subscribers but allow a new publisher to take over
		channel.publisher = nil
		if w.debug {
			wlog.DebugLog("Disconnect unregister")
		}
	}
}

func (w *WubSubApp) handleRegister(r *Register) {
	channel := w.channels[r.channel]
	// does channel already exist?
	if channel != nil {
		if channel.publisher != nil {
			var m message.Message
			if channel.publisher == r.client {
				m = message.NewErrorMessage("You are already registered to this channel!")
			} else {
				m = message.NewErrorMessage("Channel is already owned by another publisher.")
			}
			r.client.Send(&m)
			return
		} else {
			channel.publisher = r.client
		}
	} else {
		// lets register it
		channel = &Channel{r.client, make([]*WSConnection, 0)}
	}
	r.client.publishesTo[r.channel] = channel
	w.channels[r.channel] = channel
	if w.debug {
		wlog.DebugLog("Registered: '%s'", r.channel)
	}
}

func (w *WubSubApp) handlePublish(p *Publish) {
	channel := w.channels[p.channel]
	// does this channel even exist?
	if channel == nil {
		message := message.NewErrorMessage("This channel doesn't exist!")
		p.publisher.Send(&message)
		return
	}
	// is this wrong publisher?
	if channel.publisher != p.publisher {
		message := message.NewErrorMessage("You aren't the publisher of this channel!")
		p.publisher.Send(&message)
		return
	}
	// publish data
	message := message.NewReceiveMessage(p.channel, p.data)
	for _, subscriber := range channel.subscribers {
		subscriber.Send(&message)
	}
	if w.debug {
		wlog.DebugLog("'%s' published: %v", p.channel, p.data)
	}
}

func (w *WubSubApp) handleSubscribe(s *Subscribe) {
	channel := w.channels[s.channel]
	if channel == nil {
		// let's create the channel with no publisher
		channel = &Channel{nil, make([]*WSConnection, 0)}
		w.channels[s.channel] = channel
	}
	channel.subscribers = append(channel.subscribers, s.subscriber)
	if w.debug {
		wlog.DebugLog("Subscription: '%s'", s.channel)
	}
}

func (app *WubSubApp) handleConnection(rwriter http.ResponseWriter, req *http.Request) {
	client := CreateWSClient(app, rwriter, req, &app.upgrader)
	client.UpgradeAndHandle()
}
