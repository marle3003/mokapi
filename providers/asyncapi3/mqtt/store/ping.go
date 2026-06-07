package store

import (
	"mokapi/mqtt"
)

func (s *Store) ping(rw mqtt.MessageWriter, ping *mqtt.PingRequest, ctx *mqtt.ClientContext) {
	client, ok := s.clients[ctx.ClientId]
	if !ok {
		panic("client not found")
	}
	client.Alive()

	_ = rw.Write(&mqtt.Message{
		Header: &mqtt.Header{
			Type: mqtt.PINGRESP,
		},
		Payload: &mqtt.PingResponse{},
	})
}
