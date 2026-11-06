package store

import "mokapi/mqtt"

func (s *Store) disconnect(_ mqtt.MessageWriter, disconnect *mqtt.DisconnectRequest, ctx *mqtt.ClientContext) {
	client, ok := s.clients[ctx.ClientId]
	if !ok {
		panic("client not found")
	}

	if disconnect.Reason == mqtt.DisconnectWithWillMessage {
		will := client.Will()
		t := s.Topics[will.Topic]
		t.Retained = will
		for _, c := range s.clients {
			c.publish(will)
		}
	}
	s.logRequest(&DisconnectRequest{Reason: disconnect.Reason}, nil, ctx)
}
