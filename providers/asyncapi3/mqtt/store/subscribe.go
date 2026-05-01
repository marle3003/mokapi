package store

import (
	"mokapi/mqtt"
	"strings"
)

func (s *Store) subscribe(rw mqtt.MessageWriter, subscribe *mqtt.SubscribeRequest, ctx *mqtt.ClientContext) {
	client, ok := s.clients[ctx.ClientId]
	if !ok {
		panic("client not found")
	}

	res := &mqtt.SubscribeResponse{
		MessageId: subscribe.MessageId,
	}

	for _, topic := range subscribe.Topics {
		client.Subscribe(topic.Name, topic.QoS)
		res.ReasonCodes = append(res.ReasonCodes, mqtt.SubscriptionReason(topic.QoS))

		go func() {
			for _, msg := range s.getRetainedMessages(topic.Name) {
				client.publish(msg)
			}
		}()
	}

	rw.Write(&mqtt.Message{
		Header: &mqtt.Header{
			Type: mqtt.SUBACK,
		},
		Payload: res,
	})
}

func (s *Store) unsubscribe(rw mqtt.MessageWriter, req *mqtt.UnsubscribeRequest, ctx *mqtt.ClientContext) {
	client, ok := s.clients[ctx.ClientId]
	if !ok {
		panic("client not found")
	}

	res := &mqtt.UnsubscribeResponse{
		MessageId: req.MessageId,
	}

	for _, topic := range req.Topics {
		if client.Subscription == nil {
			res.ReasonCodes = append(res.ReasonCodes, mqtt.NoSubscriptionExisted)
		} else {
			_, ok = client.Subscription[topic]
			if !ok {
				res.ReasonCodes = append(res.ReasonCodes, mqtt.NoSubscriptionExisted)
			} else {
				delete(client.Subscription, topic)
				res.ReasonCodes = append(res.ReasonCodes, mqtt.UnsubscribeSuccess)
			}
		}
	}

	rw.Write(&mqtt.Message{
		Header: &mqtt.Header{
			Type: mqtt.UNSUBACK,
		},
		Payload: res,
	})
}

func (s *Store) getRetainedMessages(name string) []*Message {
	var retained []*Message
	for _, topic := range s.Topics {
		if topic.Retained == nil || !topicMatches(name, topic.Name) {
			continue
		}

		retained = append(retained, topic.Retained)
	}
	return retained
}

func topicMatches(filter string, topic string) bool {
	// special case for system topics
	if len(topic) > 0 && topic[0] == '$' {
		if len(filter) > 0 && filter[0] != '$' {
			return false
		}
	}

	fLevels := strings.Split(filter, "/")
	tLevels := strings.Split(topic, "/")

	for i := 0; i < len(fLevels); i++ {
		f := fLevels[i]

		// multi level matches all
		if f == "#" {
			return true
		}

		if i >= len(tLevels) {
			return false
		}

		// single level wildcard or exact match
		if f == "+" || f == tLevels[i] {
			continue
		}

		return false
	}

	return len(fLevels) == len(tLevels)
}
