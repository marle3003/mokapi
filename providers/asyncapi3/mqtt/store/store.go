package store

import (
	engine "mokapi/engine/common"
	"mokapi/mqtt"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"mokapi/version"
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
	eh         events.Handler
	cfg        *asyncapi3.Config
}

func New(cfg *asyncapi3.Config, emitter engine.EventEmitter, eh events.Handler) *Store {
	s := &Store{
		RetryInterval: 10 * time.Second,
		Topics:        make(map[string]*Topic),
		close:         make(chan bool, 1),
		cfg:           cfg,
		eh:            eh,
	}

	s.Update(cfg)

	s.addSysTopic("$SYS/broker/version", "Mokapi "+version.BuildVersion)
	s.addSysTopic("$SYS/broker/uptime", time.Now().Format(time.RFC3339))

	return s
}

func (s *Store) Update(cfg *asyncapi3.Config) {
	for _, ch := range cfg.Channels {
		if ch == nil || ch.Value == nil {
			continue
		}
		if !ch.Value.IsChannelAvailable("mqtt") {
			continue
		}
		if s.Topics == nil {
			s.Topics = make(map[string]*Topic)
		}
		s.Topics[ch.Value.Name] = &Topic{Name: ch.Value.Name, cfg: ch.Value}
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
		s.publish(rw, msg, req.Header.QoS, req.Header.Retain, ctx)
	case *mqtt.UnsubscribeRequest:
		s.unsubscribe(rw, msg, ctx)
	case *mqtt.PingRequest:
		_ = rw.Write(&mqtt.Message{
			Header: &mqtt.Header{
				Type: mqtt.PINGRESP,
			},
			Payload: &mqtt.PingResponse{},
		})
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
