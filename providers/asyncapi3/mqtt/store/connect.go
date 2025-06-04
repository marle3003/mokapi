package store

import (
	log "github.com/sirupsen/logrus"
	"mokapi/mqtt"
)

func (s *Store) connect(rw mqtt.ResponseWriter, connect *mqtt.ConnectRequest, ctx *mqtt.ClientContext) {

	if len(connect.ClientId) == 0 || len(connect.ClientId) > 23 {
		rw.Write(mqtt.CONNACK, &mqtt.ConnectResponse{
			SessionPresent: false,
			ReturnCode:     mqtt.ErrIdentifierRejected,
		})
		return
	}

	sessionPresent := false
	if connect.CleanSession {
		delete(s.clients, connect.ClientId)
	}
	if _, ok := s.clients[connect.ClientId]; ok {
		sessionPresent = true
	} else {
		if s.clients == nil {
			s.clients = map[string]*Client{}
		}
		s.clients[connect.ClientId] = &Client{Id: connect.ClientId, ctx: ctx}
	}

	if connect.Topic != "" {
		s.m.RLock()
		defer s.m.RUnlock()

		if t, ok := s.Topics[connect.Topic]; ok {
			m := &Message{
				Data: connect.Message,
				QoS:  connect.WillQoS,
			}
			_ = m
			_ = t
		} else {
			log.Infof("mqtt broker: invalid topic %v", connect.Topic)
			rw.Write(mqtt.CONNACK, &mqtt.ConnectResponse{
				SessionPresent: sessionPresent,
				ReturnCode:     mqtt.ErrUnspecifiedError,
			})
			return
		}
	}

	rw.Write(mqtt.CONNACK, &mqtt.ConnectResponse{
		SessionPresent: sessionPresent,
		ReturnCode:     mqtt.Accepted,
	})
}
