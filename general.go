package main

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/premiering/wubsub/log"
	"github.com/premiering/wubsub/message"
)

type Channel struct {
	publisher   *WubSubSession
	id          string
	name        string
	description string
	ps_topics   map[string]*PubSubTopic
	rp_topics   map[string]*RestPieMethod
}

func (c *Channel) IsAlive() bool {
	return c.publisher != nil
}

type PubSubTopic struct {
	subscribers []*WubSubSession
}

func (p *PubSubTopic) RemoveSub(w *WubSubSession) {
	for i, v := range p.subscribers {
		if v == w {
			p.subscribers = append(p.subscribers[:i], p.subscribers[i+1:]...)
		}
	}
}

type RestPieMethod struct {
	ongoing map[uuid.UUID]*WubSubSession
}

type RegisterData struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	PubSub      RegisterPubSubData  `json:"pubsub"`
	RestPie     RegisterRestPieData `json:"restpie"`
}

type RegisterPubSubData struct {
	Topics []string `json:"topics"`
}

type RegisterRestPieData struct {
	Methods []string `json:"methods"`
}

func handleRegister(m *message.Message, w *WubSubSession) {
	log.DebugLog("got handle register")

	// slower and hacky
	b, err := json.Marshal(m.Data)
	if err != nil {
		return
	}
	var data RegisterData
	json.Unmarshal(b, &data)

	ch, exists := wubsub.channels[m.Channel]
	ps_topics := make(map[string]*PubSubTopic)
	for _, v := range data.PubSub.Topics {
		ps_topics[v] = &PubSubTopic{make([]*WubSubSession, 0)}
	}
	rp_topics := make(map[string]*RestPieMethod)
	for _, v := range data.RestPie.Methods {
		rp_topics[v] = &RestPieMethod{make(map[uuid.UUID]*WubSubSession)}
	}

	if !exists {
		ch = &Channel{
			w,
			m.Channel,
			data.Name,
			data.Description,
			ps_topics,
			rp_topics,
		}
		wubsub.channels[m.Channel] = ch
		log.DebugLog("added channel %v", *ch)
	} else {
		if ch.publisher != nil {
			// error
			return
		}
		ch.publisher = w
		ch.name = data.Name
		ch.description = data.Description
		// transfer subscribers from the unowned topic/channel to the new one
		for k, v := range ch.ps_topics {
			new, exists := ps_topics[k]
			if exists {
				new.subscribers = v.subscribers
			} else {
				for _, sub := range new.subscribers {
					sub.RemovePSSub(k)
				}
			}
		}
		ch.ps_topics = ps_topics
		ch.rp_topics = rp_topics
		log.DebugLog("added channel %v", *ch)
	}
	w.owned_channels = append(w.owned_channels, ch)
}

func handleDisconect(s *WubSubSession) {
	log.DebugLog("got handle disconnect")
	for _, v := range s.owned_channels {
		v.publisher = nil
	}
	for _, v := range s.ps_subscriptions {
		v.RemoveSub(s)
	}
}
