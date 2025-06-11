package store

import (
	engine "mokapi/engine/common"
	"mokapi/mqtt"
	"mokapi/providers/asyncapi3"
	"sync"
	"time"
)

type Store struct {
	RetryInterval time.Duration

	clients    map[string]*Client
	Topics     map[string]*Topic
	startedQoS bool
	m          sync.RWMutex
	close      chan bool
}

func New(cfg *asyncapi3.Config, emitter engine.EventEmitter) *Store {
	s := &Store{
		RetryInterval: 10 * time.Second,
		close:         make(chan bool, 1),
	}

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

func (s *Store) ServeMessage(rw mqtt.MessageWriter, req *mqtt.Message) {
	ctx := mqtt.ClientFromContext(req.Context)

	switch msg := req.Payload.(type) {
	case *mqtt.ConnectRequest:
		s.connect(rw, msg, ctx)
	case *mqtt.SubscribeRequest:
		s.subscribe(rw, msg, ctx)
	case *mqtt.PublishRequest:
		s.publish(rw, msg, req.Header.QoS, req.Header.Retain)
	}
}

func (s *Store) Close() {
	s.close <- true
}

func (s *Store) startQoS() {
	if s.startedQoS {
		return
	}

	s.m.Lock()
	defer s.m.Unlock()

	if s.startedQoS {
		return
	}

	ticker := time.NewTicker(s.RetryInterval)

	go func() {
		for {
			select {
			case <-s.close:
				ticker.Stop()
				return
			case <-ticker.C:
				for _, c := range s.clients {
					c.ResendInflight(s.RetryInterval)
				}
			}
		}
	}()

	s.startedQoS = true
}
