package store

import "mokapi/mqtt"

func (s *Store) publish(rw mqtt.ResponseWriter, publish *mqtt.PublishRequest, qos byte, retain bool) {
	msg := &Message{
		Topic:  publish.Topic,
		Data:   publish.Data,
		QoS:    qos,
		Retain: retain,
	}

	if topic, ok := s.Topics[msg.Topic]; ok {
		if retain {
			topic.Retained = msg
		}
	}

	rw.Write(mqtt.PUBACK, &mqtt.PublishResponse{
		MessageId: publish.MessageId,
	})

	go func() {
		for _, client := range s.clients {
			client.publish(msg)
		}
	}()
}
