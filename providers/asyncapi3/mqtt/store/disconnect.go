package store

import "mokapi/mqtt"

func (s *Store) disconnect(_ mqtt.MessageWriter, disconnect *mqtt.DisconnectRequest, ctx *mqtt.ClientContext) {
	client, ok := s.clients[ctx.ClientId]
	if !ok {
		panic("client not found")
	}

	if disconnect.Reason == mqtt.DisconnectWithWillMessage {
		t := s.Topics[client.WillMessage.Topic]
		t.Retained = client.WillMessage
		for _, c := range s.clients {
			c.publish(client.WillMessage)
		}
	}
	s.logRequest(&DisconnectRequest{Reason: disconnect.Reason}, nil, ctx)
}
