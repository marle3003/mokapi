package store

import "mokapi/mqtt"

func (s *Store) subscribe(rw mqtt.ResponseWriter, subscribe *mqtt.SubscribeRequest, ctx *mqtt.ClientContext) {
	client, ok := s.clients[ctx.ClientId]
	if !ok {
		panic("client not found")
	}

	res := &mqtt.SubscribeResponse{
		MessageId: subscribe.MessageId,
	}

	for _, topic := range subscribe.Topics {
		client.Subscribe(topic.Name, topic.QoS)
		res.TopicQoS = append(res.TopicQoS, topic.QoS)

		go func() {
			for _, msg := range s.getRetainedMessages(topic.Name) {
				client.publish(msg)
			}
		}()
	}

	rw.Write(mqtt.SUBACK, res)
}

func (s *Store) getRetainedMessages(name string) []*Message {
	var retained []*Message
	for _, topic := range s.Topics {
		if topic.Name == name {
			if topic.Retained != nil {
				{
					retained = append(retained, topic.Retained)
				}
			}
		}
	}
	return retained
}
