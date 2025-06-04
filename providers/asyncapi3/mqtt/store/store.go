package store

import (
	engine "mokapi/engine/common"
	"mokapi/mqtt"
	"mokapi/providers/asyncapi3"
	"sync"
)

type Store struct {
	clients map[string]*Client
	Topics  map[string]*Topic
	m       sync.RWMutex
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
	for _, ch := range cfg.Channels {
		if s.Topics == nil {
			s.Topics = make(map[string]*Topic)
		}
		s.Topics[ch.Value.Name] = &Topic{Name: ch.Value.Name}
	}
}

func (s *Store) ServeMessage(rw mqtt.ResponseWriter, req *mqtt.Request) {
	ctx := mqtt.ClientFromContext(req.Context)

	switch msg := req.Message.(type) {
	case *mqtt.ConnectRequest:
		s.connect(rw, msg)
	case *mqtt.SubscribeRequest:
		s.subscribe(rw, msg, ctx)
	case *mqtt.PublishRequest:
		s.publish(rw, msg, req.Header.QoS, req.Header.Retain)
	}
}
