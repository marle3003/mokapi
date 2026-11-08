package store

import (
	"mokapi/mqtt"

	log "github.com/sirupsen/logrus"
)

// disconnect holds s.m only to resolve the client and to log the request. The
// will is published afterwards without the lock, see ServeMessage.
func (s *Store) disconnect(disconnect *mqtt.DisconnectRequest, ctx *mqtt.ClientContext) {
	s.m.Lock()
	client, ok := s.clients[ctx.ClientId]
	if !ok {
		s.m.Unlock()
		panic("client not found")
	}
	s.logRequest(&DisconnectRequest{Reason: disconnect.Reason}, nil, ctx)
	s.m.Unlock()

	if disconnect.Reason != mqtt.DisconnectWithWillMessage {
		return
	}
	will := client.Will()
	if will == nil {
		return
	}
	// The will goes through Publish like a regular message, so it is validated,
	// handled by scripts, logged and retained only if requested. The client is
	// already gone, so a failure can only be logged.
	_, err := s.Publish(
		&mqtt.PublishRequest{Topic: will.Topic, Data: will.Data},
		PublishArgs{QoS: will.QoS, Retain: will.Retain, ClientId: ctx.ClientId},
	)
	if err != nil {
		log.Errorf("mqtt: failed to publish will message of client '%s': %s", ctx.ClientId, err)
	}
}
