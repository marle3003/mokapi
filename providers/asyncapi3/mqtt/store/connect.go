package store

import (
	log "github.com/sirupsen/logrus"
	"mokapi/mqtt"
)

func (s *Store) connect(rw mqtt.MessageWriter, connect *mqtt.ConnectRequest, ctx *mqtt.ClientContext) {

	if ctx != nil {
		ctx.ClientId = connect.ClientId
	}

	if len(connect.ClientId) == 0 || len(connect.ClientId) > 23 {
		rw.Write(&mqtt.Message{
			Header: &mqtt.Header{
				Type: mqtt.CONNACK,
			},
			Payload: &mqtt.ConnectResponse{
				SessionPresent: false,
				ReturnCode:     mqtt.ErrIdentifierRejected,
			},
		})
		return
	}

	sessionPresent := false
	if connect.CleanSession {
		delete(s.clients, connect.ClientId)
	}
	if c, ok := s.clients[connect.ClientId]; ok {
		sessionPresent = true
		c.ctx = ctx
		go c.ResendInflight(0)
	} else {
		if s.clients == nil {
			s.clients = map[string]*Client{}
		}
		s.clients[connect.ClientId] = &Client{Id: connect.ClientId, ctx: ctx}
	}

	if connect.Topic != "" {
		s.m.Lock()

		if t, ok := s.Topics[connect.Topic]; ok {
			m := &Message{
				Data: connect.Message,
				QoS:  connect.WillQoS,
			}
			for _, c := range s.clients {
				c.publish(m)
			}
			if connect.WillRetain {
				t.Retained = m
			}
			s.m.Unlock()
		} else {
			log.Infof("mqtt broker: invalid topic %v", connect.Topic)
			rw.Write(&mqtt.Message{
				Header: &mqtt.Header{
					Type: mqtt.CONNACK,
				},
				Payload: &mqtt.ConnectResponse{
					SessionPresent: sessionPresent,
					ReturnCode:     mqtt.ErrUnspecifiedError,
				},
			})
			s.m.Unlock()
			return
		}
	}

	rw.Write(&mqtt.Message{
		Header: &mqtt.Header{
			Type: mqtt.CONNACK,
		},
		Payload: &mqtt.ConnectResponse{
			SessionPresent: sessionPresent,
			ReturnCode:     mqtt.Accepted,
		},
	})

	s.startQoS()
}
