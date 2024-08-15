package main

import (
	"github.com/google/uuid"
	"github.com/premiering/wubsub/log"
	"github.com/premiering/wubsub/message"
)

// A request is initiated
func handleRPDoRequest(m *message.Message, w *WubSubSession) {
	ch := wubsub.channels[m.Channel]
	if ch == nil || ch.publisher == nil {
		msg := message.NewRPAckMsg(m.Channel, m.Topic, false, "")
		w.Send(&msg)
		return
	}
	var rp *RestPieMethod
	exists := false
	for k, v := range ch.rp_topics {
		if k == m.Topic {
			rp = v
			exists = true
			break
		}
	}
	if !exists {
		msg := message.NewRPAckMsg(m.Channel, m.Topic, false, "")
		w.Send(&msg)
		return
	}
	// we're good now
	// acknowledge the request as good with the id
	id := uuid.New()
	msg := message.NewRPAckMsg(m.Channel, m.Topic, true, id.String())
	w.Send(&msg)

	rp.ongoing[id] = w
	req := message.NewRPReqMsg(m.Channel, m.Topic, id.String(), m.Data)
	ch.publisher.Send(&req)

	log.DebugLog("rp-do request, channel %s, topic %s, id %s", m.Channel, m.Topic, id.String())
}

// Responding to a request
func handleRPRespond(m *message.Message, w *WubSubSession) {
	ch := wubsub.channels[m.Channel]
	if ch == nil || ch.publisher != w {
		return
	}
	topic := ch.rp_topics[m.Topic]
	if topic == nil {
		return
	}
	u, err := uuid.Parse(m.RequestId)
	if err != nil {
		return
	}
	recipient := topic.ongoing[u]
	if recipient == nil {
		return
	}
	response := message.NewRPRecvMsg(m.Channel, m.Topic, u.String(), m.Data)
	recipient.Send(&response)
	delete(topic.ongoing, u)
	log.DebugLog("rp-recv response, channel %s, topic %s, id %s", m.Channel, m.Topic, u.String())
}
