package store

import (
	"mokapi/mqtt"

	log "github.com/sirupsen/logrus"
)

func (s *Store) connect(rw mqtt.MessageWriter, connect *mqtt.ConnectRequest, ctx *mqtt.ClientContext) {

	reqLog := &ConnectRequest{
		Version:      connect.Version,
		CleanSession: connect.CleanSession,
		KeepAlive:    connect.KeepAlive,
		Message:      nil,
		Username:     connect.Username,
		Password:     connect.Password,
	}
	if connect.Topic != "" {
		reqLog.Message = &PublishMessage{
			QoS:     connect.WillQoS,
			Retain:  connect.WillRetain,
			Topic:   connect.Topic,
			Message: string(connect.Message),
		}
	}

	if ctx != nil {
		ctx.ClientId = connect.ClientId
	}

	if len(connect.ClientId) == 0 || len(connect.ClientId) > 23 {
		err := rw.Write(&mqtt.Message{
			Header: &mqtt.Header{
				Type: mqtt.CONNACK,
			},
			Payload: &mqtt.ConnectResponse{
				ReasonCode: mqtt.ErrIdentifierRejected,
			},
		})
		if err != nil {
			log.Errorf("mqtt: failed to write connect response: %v", err)
		}
		s.logRequest(reqLog, ConnectResponse{
			ReasonCode: mqtt.ErrIdentifierRejected,
		}, ctx)
		return
	}

	sessionPresent := false
	if connect.CleanSession {
		delete(s.clients, connect.ClientId)
	}
	c, ok := s.clients[connect.ClientId]
	if ok {
		sessionPresent = true
		c.attach(ctx, connect.KeepAlive)
		go c.ResendInflight(0)
	} else {
		if s.clients == nil {
			s.clients = map[string]*Client{}
		}
		c = &Client{
			Id:                    connect.ClientId,
			SessionExpiryInterval: connect.Properties.SessionExpiryInterval(),
			state:                 ClientConnected,
		}
		c.attach(ctx, connect.KeepAlive)
		s.clients[connect.ClientId] = c
	}

	if connect.Topic != "" {
		if _, ok := s.topic(connect.Topic); ok {

			if connect.WillFlag {
				c.SetWill(&Message{
					Topic:  connect.Topic,
					Data:   connect.Message,
					QoS:    connect.WillQoS,
					Retain: connect.WillRetain,
				})
			}
		} else {
			log.Infof("mqtt broker: invalid topic %v", connect.Topic)
			err := rw.Write(&mqtt.Message{
				Header: &mqtt.Header{
					Type: mqtt.CONNACK,
				},
				Payload: &mqtt.ConnectResponse{
					SessionPresent: sessionPresent,
					ReasonCode:     mqtt.ErrTopicNameInvalid,
				},
			})
			if err != nil {
				log.Errorf("mqtt: failed to write connect response: %v", err)
			}
			s.logRequest(reqLog, ConnectResponse{
				ReasonCode: mqtt.ErrTopicNameInvalid,
			}, ctx)
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

	s.logRequest(reqLog, ConnectResponse{
		SessionPresent: sessionPresent,
		ReasonCode:     mqtt.Success,
	}, ctx)

	s.startQoS()
}
