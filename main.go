package main

import (
	"encoding/json"
	"flag"
	golog "log"
	"net/http"
	"strconv"

	"github.com/olahol/melody"
	"github.com/premiering/wubsub/log"
	"github.com/premiering/wubsub/message"
)

type WubSub struct {
	mel *melody.Melody

	process_state ProcessState

	channels map[string]*Channel
}

type TupleMessageSession struct {
	message *message.Message
	session *WubSubSession
}

type ProcessState struct {
	incoming     chan TupleMessageSession
	disconnected chan *WubSubSession
}

type WubSubSession struct {
	session *melody.Session

	owned_channels []*Channel

	ps_subscriptions map[string]*PubSubTopic
}

func (s *WubSubSession) Send(message *message.Message) {
	b, err := json.Marshal(message)
	if err != nil {
		return
	}
	s.session.Write(b)
}

func (s *WubSubSession) SendError(err string) {
	m := message.NewErrorMsg(err)
	s.Send(&m)
}

func (s *WubSubSession) RemovePSSub(topic string) {
	s.ps_subscriptions[topic] = nil
}

type MessageHandler func(*message.Message, *WubSubSession)

var wubsub *WubSub = &WubSub{
	melody.New(),
	ProcessState{make(chan TupleMessageSession), make(chan *WubSubSession)},
	make(map[string]*Channel),
}

func main() {
	port := flag.Int("port", 9190, "The port the server will listen on")
	debugMode := flag.Bool("debug", false, "When enabled, pub/sub events are logged to output")
	tls := flag.Bool("tls", false, "Enables TLS, (requires key and cert file)")
	keyfile := flag.String("keyfile", "", "Path to key file relative to working dir")
	certfile := flag.String("certfile", "", "Path to cert file relative to working dir")
	flag.Parse()

	log.SetApp("wubsub", *debugMode)
	mel := wubsub.mel // make melody shorthand

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mel.HandleRequest(w, r)
	})

	mel.HandleConnect(func(s *melody.Session) {
		wubsession := &WubSubSession{s, make([]*Channel, 0), make(map[string]*PubSubTopic, 0)}
		s.Set("wubsession", wubsession)
		log.DebugLog("connection received")
	})

	mel.HandleMessage(func(s *melody.Session, b []byte) {
		anysession, exists := s.Get("wubsession")
		if !exists {
			return
		}
		wubsession := anysession.(*WubSubSession)

		var message message.Message
		err := json.Unmarshal(b, &message)
		if err != nil {
			return
		}
		_, exists = typeToHandler(message.Type)
		if exists {
			log.DebugLog("valid message received")
			wubsub.process_state.incoming <- TupleMessageSession{&message, wubsession}
		}
	})

	mel.HandleDisconnect(func(s *melody.Session) {
		anysession, exists := s.Get("wubsession")
		if !exists {
			return
		}
		wubsub.process_state.disconnected <- anysession.(*WubSubSession)
		log.DebugLog("disconnection")
	})

	go func() {
		for {
			process()
		}
	}()

	address := ":" + strconv.Itoa(*port)
	log.InfoLog("Listening on port " + address)
	if *tls {
		golog.Fatal(http.ListenAndServeTLS(address, *certfile, *keyfile, nil))
	} else {
		golog.Fatal(http.ListenAndServe(address, nil))
	}
}

func process() {
	select {
	case session := <-wubsub.process_state.disconnected:
		handleDisconect(session)
	case tup := <-wubsub.process_state.incoming:
		handler, _ := typeToHandler(tup.message.Type)
		handler(tup.message, tup.session)
	}
}

func typeToHandler(t string) (MessageHandler, bool) {
	switch t {
	case message.Register:
		return handleRegister, true
	case message.PSPublish:
		return handlePSPublish, true
	case message.PSSubscribe:
		return handlePSSubscribe, true
	case message.RPRequestInitiated:
		return handleRPDoRequest, true
	case message.RPResponding:
		return handleRPRespond, true
	}
	return nil, false
}
