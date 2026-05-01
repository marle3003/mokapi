package store

import (
	"mokapi/mqtt"

	log "github.com/sirupsen/logrus"
)

func (s *Store) connect(rw mqtt.MessageWriter, connect *mqtt.ConnectRequest, ctx *mqtt.ClientContext) {

	if ctx != nil {
		ctx.ClientId = connect.ClientId
	}

	if len(connect.ClientId) == 0 || len(connect.ClientId) > 23 {
		err := rw.Write(&mqtt.Message{
			Header: &mqtt.Header{
				Type: mqtt.CONNACK,
			},
			Payload: &mqtt.ConnectResponse{
				SessionPresent: false,
				ReasonCode:     mqtt.ErrIdentifierRejected,
			},
		})
		if err != nil {
			log.Errorf("mqtt: failed to write connect response: %v", err)
		}
		return
	}

	sessionPresent := false
	if connect.CleanSession {
		delete(s.clients, connect.ClientId)
	}
	c, ok := s.clients[connect.ClientId]
	if ok {
		sessionPresent = true
		c.ctx = ctx
		go c.ResendInflight(0)
	} else {
		if s.clients == nil {
			s.clients = map[string]*Client{}
		}
		c = &Client{
			Id:                    connect.ClientId,
			ctx:                   ctx,
			SessionExpiryInterval: connect.Properties.SessionExpiryInterval(),
		}
		s.clients[connect.ClientId] = c
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
			err := rw.Write(&mqtt.Message{
				Header: &mqtt.Header{
					Type: mqtt.CONNACK,
				},
				Payload: &mqtt.ConnectResponse{
					SessionPresent: sessionPresent,
					ReasonCode:     mqtt.ErrUnspecifiedError,
				},
			})
			if err != nil {
				log.Errorf("mqtt: failed to write connect response: %v", err)
			}
			s.m.Unlock()
			return
		}
	}

	err := rw.Write(&mqtt.Message{
		Header: &mqtt.Header{
			Type: mqtt.CONNACK,
		},
		Payload: &mqtt.ConnectResponse{
			SessionPresent: sessionPresent,
			ReasonCode:     mqtt.Success,
			Properties: mqtt.Properties{
				mqtt.SessionExpiryInterval: c.SessionExpiryInterval,
			},
		},
	})
	if err != nil {
		log.Errorf("mqtt: failed to write connect response: %v", err)
	}

	s.startQoS()
}
