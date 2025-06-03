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
		if client.Topics == nil {
			client.Topics = map[string]*SubscribedTopic{}
		}

		client.Topics[topic.Name] = &SubscribedTopic{
			Name: topic.Name,
			QoS:  topic.QoS,
		}
		res.TopicQoS = append(res.TopicQoS, topic.QoS)
	}

	rw.Write(mqtt.CONNACK, res)
}
