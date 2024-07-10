package app

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	wlog "github.com/premiering/wubsub/log"
)

type Register struct {
	client  *WSClient
	channel string
}

type Publish struct {
	publisher *WSClient
	channel   string
	data      interface{}
}

type Subscribe struct {
	subscriber *WSClient
	channel    string
}

type Disconnect struct {
	client *WSClient
}

type Channel struct {
	publisher   *WSClient
	subscribers []*WSClient
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
			for _, channel := range disconnect.client.publishesTo {
				// leave the channel in the map so that we still know
				// its subscribers but allow a new publisher to take over
				channel.publisher = nil
			}
		case register := <-app.registers:
			channel := app.channels[register.channel]
			// does channel already exist?
			if channel != nil {
				if channel.publisher != nil {
					message := NewErrorMessage("Channel is already owned by another publisher.")
					register.client.Send(&message)
				} else {
					channel.publisher = register.client
				}
				continue
			}
			// lets register it
			channel = &Channel{register.client, make([]*WSClient, 0)}
			app.channels[register.channel] = channel
			if app.debug {
				wlog.DebugLog("A client registered to " + register.channel)
			}
			break
		case publish := <-app.publishes:
			channel := app.channels[publish.channel]
			// does this channel even exist?
			if channel == nil {
				message := NewErrorMessage("This channel doesn't exist!")
				publish.publisher.Send(&message)
				continue
			}
			// is this wrong publisher?
			if channel.publisher != publish.publisher {
				message := NewErrorMessage("You aren't the publisher of this channel!")
				publish.publisher.Send(&message)
				continue
			}
			// publish data
			message := NewReceiveMessage(publish.channel, publish.data)
			for _, subscriber := range channel.subscribers {
				subscriber.Send(&message)
			}
			if app.debug {
				wlog.DebugLog("A client published to " + publish.channel)
			}
			break
		case subscribe := <-app.subscribes:
			channel := app.channels[subscribe.channel]
			if channel == nil {
				// let's create the channel with no publisher
				channel = &Channel{nil, make([]*WSClient, 0)}
				app.channels[subscribe.channel] = channel
			}
			channel.subscribers = append(channel.subscribers, subscribe.subscriber)
			if app.debug {
				wlog.DebugLog("A client subscribed to " + subscribe.channel)
			}
			break
		}
	}
}

func (app *WubSubApp) handleConnection(rwriter http.ResponseWriter, req *http.Request) {
	client := CreateWSClient(app, rwriter, req, &app.upgrader)
	client.UpgradeAndHandle()
}
