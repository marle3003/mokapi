package store

import (
	"maps"
	engine "mokapi/engine/common"
	"mokapi/mqtt"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/version"
	"slices"
	"sync"
	"time"
)

type Store struct {
	RetryInterval time.Duration

	clients      map[string]*Client
	topics       map[string]*Topic
	startedQoS   bool
	m            sync.RWMutex
	close        chan bool
	eh           events.Handler
	cfg          *asyncapi3.Config
	monitor      *monitor.Mqtt
	eventEmitter engine.EventEmitter

	stopClientCleaner chan bool
}

func New(cfg *asyncapi3.Config, emitter engine.EventEmitter, eh events.Handler, m *monitor.Mqtt) *Store {
	s := &Store{
		RetryInterval: 10 * time.Second,
		topics:        make(map[string]*Topic),
		close:         make(chan bool, 1),
		cfg:           cfg,
		eh:            eh,
		monitor:       m,
		eventEmitter:  emitter,
	}

	s.Update(cfg)

	s.addSysTopic("$SYS/broker/version", "Mokapi "+version.BuildVersion)
	s.addSysTopic("$SYS/broker/uptime", time.Now().Format(time.RFC3339))

	s.startClientSessionCleaner()

	return s
}

// Topics returns a snapshot of the topics, including the $SYS topics.
func (s *Store) Topics() []*Topic {
	s.m.RLock()
	defer s.m.RUnlock()
	return slices.Collect(maps.Values(s.topics))
}

func (s *Store) Update(cfg *asyncapi3.Config) {
	s.m.Lock()
	defer s.m.Unlock()

	s.cfg = cfg

	for address, ch := range cfg.Channels {
		if ch == nil || ch.Value == nil {
			continue
		}
		if !ch.Value.IsChannelAvailable("mqtt") {
			continue
		}
		if len(ch.Value.Parameters) > 0 {
			continue
		}

		if s.topics == nil {
			s.topics = make(map[string]*Topic)
		}

		if ch.Value.Address != "" {
			address = ch.Value.Address
		}

		s.topics[address] = &Topic{Name: address, cfg: ch.Value}

	}
}

func (s *Store) ServeMessage(rw mqtt.MessageWriter, req *mqtt.Message) {
	ctx := mqtt.ClientFromContext(req.Context)

	// A publish runs JavaScript handlers, which may publish themselves
	// (mqtt.publish) and so re-enter the store. It must not hold s.m.
	// The same applies to a disconnect, which publishes the client's will.
	switch msg := req.Payload.(type) {
	case *mqtt.PublishRequest:
		s.publish(rw, msg, req.Header.QoS, req.Header.Retain, ctx)
		return
	case *mqtt.DisconnectRequest:
		s.disconnect(msg, ctx)
		return
	}

	s.m.Lock()
	defer s.m.Unlock()

	switch msg := req.Payload.(type) {
	case *mqtt.ConnectRequest:
		s.connect(rw, msg, ctx)
	case *mqtt.SubscribeRequest:
		s.subscribe(rw, msg, ctx)
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
				for _, c := range s.Clients() {
					c.ResendInflight(s.RetryInterval)
				}
			}
		}
	}()

	s.startedQoS = true
}

func (s *Store) Clients() []*Client {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.clientList()
}

// clientList returns a snapshot, the caller must hold s.m.
func (s *Store) clientList() []*Client {
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
		client.checkSession(now)
	}
}
