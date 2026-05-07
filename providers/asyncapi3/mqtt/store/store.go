package store

import (
	engine "mokapi/engine/common"
	"mokapi/mqtt"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
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
	monitor    *monitor.Mqtt

	stopClientCleaner chan bool
}

func New(cfg *asyncapi3.Config, emitter engine.EventEmitter, eh events.Handler, m *monitor.Mqtt) *Store {
	s := &Store{
		RetryInterval: 10 * time.Second,
		Topics:        make(map[string]*Topic),
		close:         make(chan bool, 1),
		cfg:           cfg,
		eh:            eh,
		monitor:       m,
	}

	s.Update(cfg)

	s.addSysTopic("$SYS/broker/version", "Mokapi "+version.BuildVersion)
	s.addSysTopic("$SYS/broker/uptime", time.Now().Format(time.RFC3339))

	s.startClientSessionCleaner()

	return s
}

func (s *Store) Update(cfg *asyncapi3.Config) {
	s.cfg = cfg

	for address, ch := range cfg.Channels {
		if ch == nil || ch.Value == nil {
			continue
		}
		if !ch.Value.IsChannelAvailable("mqtt") {
			continue
		}
		if s.Topics == nil {
			s.Topics = make(map[string]*Topic)
		}

		if ch.Value.Address != "" {
			address = ch.Value.Address
		}

		if len(ch.Value.Parameters) == 0 {
			s.Topics[address] = &Topic{Name: address, cfg: ch.Value}
		}
	}
}

func (s *Store) ServeMessage(rw mqtt.MessageWriter, req *mqtt.Message) {
	s.m.Lock()
	defer s.m.Unlock()

	ctx := mqtt.ClientFromContext(req.Context)

	switch msg := req.Payload.(type) {
	case *mqtt.ConnectRequest:
		s.connect(rw, msg, ctx)
	case *mqtt.DisconnectRequest:
		s.disconnect(rw, msg, ctx)
	case *mqtt.SubscribeRequest:
		s.subscribe(rw, msg, ctx)
	case *mqtt.PublishRequest:
		s.publish(rw, msg, req.Header.QoS, req.Header.Retain, ctx)
	case *mqtt.UnsubscribeRequest:
		s.unsubscribe(rw, msg, ctx)
	case *mqtt.PingRequest:
		s.ping(rw, msg, ctx)
	}
}

func (s *Store) Close() {
	s.close <- true
	close(s.stopClientCleaner)
}

func (s *Store) startQoS() {
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

func (s *Store) Clients() []*Client {
	var result []*Client
	for _, c := range s.clients {
		result = append(result, c)
	}
	return result
}

func (s *Store) startClientSessionCleaner() {
	ticker := time.NewTicker(30 * time.Second)
	s.stopClientCleaner = make(chan bool)

	go func() {
		for {
			select {
			case <-s.stopClientCleaner:
				ticker.Stop()
				return
			case <-ticker.C:
				s.checkClientSession()
			}
		}
	}()
}

func (s *Store) checkClientSession() {
	s.m.Lock()
	defer s.m.Unlock()

	now := time.Now()
	for _, client := range s.clients {
		if now.After(client.LastSeen.Add(time.Duration(client.KeepAlive) * time.Second)) {
			client.State = ClientConnected
		}
	}
}
