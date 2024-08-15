package main

import (
	"github.com/premiering/wubsub/log"
	"github.com/premiering/wubsub/message"
)

func handlePSPublish(m *message.Message, w *WubSubSession) {
	ch := wubsub.channels[m.Channel]
	if ch == nil || ch.publisher != w {
		return
	}
	topic, exists := ch.ps_topics[m.Topic]
	if !exists {
		return
	}
	msg := message.NewPSReceiveMsg(m.Channel, m.Topic, m.Data)
	for _, sub := range topic.subscribers {
		sub.Send(&msg)
	}
	log.DebugLog("ps-pub, channel %s, topic %s", m.Channel, m.Topic)
}

func handlePSSubscribe(m *message.Message, w *WubSubSession) {
	ch, exists := wubsub.channels[m.Channel]
	if !exists {
		ch = &Channel{nil, m.Channel, "", "", make(map[string]*PubSubTopic), make(map[string]*RestPieMethod)}
		wubsub.channels[m.Channel] = ch
	}
	topic, exists := ch.ps_topics[m.Topic]
	if !exists {
		topic = &PubSubTopic{make([]*WubSubSession, 0)}
		ch.ps_topics[m.Topic] = topic
	}
	topic.subscribers = append(topic.subscribers, w)
	log.DebugLog("ps-sub, channel %s, topic %s", m.Channel, m.Topic)
}
