package store

import (
	log "github.com/sirupsen/logrus"
	engine "mokapi/engine/common"
	"mokapi/mqtt"
	"mokapi/providers/asyncapi3"
	"sync"
)

type Store struct {
	Topics map[string]*Topic
	m      sync.RWMutex
}

func New(cfg *asyncapi3.Config, emitter engine.EventEmitter) *Store {
	s := &Store{}

	for _, ch := range cfg.Channels {
		if s.Topics == nil {
			s.Topics = make(map[string]*Topic)
		}

		if ch != nil && ch.Value != nil {
			s.Topics[ch.Value.Name] = &Topic{
				Name: ch.Value.Name,
			}
		}
	}

	return s
}

func (s *Store) Update(cfg *asyncapi3.Config) {
}

func (s *Store) ServeMessage(rw mqtt.ResponseWriter, req *mqtt.Request) {
	switch msg := req.Message.(type) {
	case *mqtt.ConnectRequest:
		if msg.Topic != "" {
			s.m.RLock()
			defer s.m.RUnlock()

			if t, ok := s.Topics[msg.Topic]; ok {
				m := &Message{
					Data:   msg.Message,
					QoS:    msg.Header.WillQoS,
					Retain: msg.Header.WillRetain,
				}
				t.AddMessage(m)
			} else {
				log.Infof("mqtt broker: invalid topic %v", msg.Topic)
				rw.Write(mqtt.CONNACK, &mqtt.ConnectResponse{
					SessionPresent: false,
					ReturnCode:     mqtt.ErrIdentifierRejected,
				})
				return
			}
		}

		rw.Write(mqtt.CONNACK, &mqtt.ConnectResponse{
			SessionPresent: !msg.Header.CleanSession,
			ReturnCode:     mqtt.Accepted,
		})
	}
}
