package websocket

import (
	engine "mokapi/engine/common"
	"mokapi/providers/asyncapi3"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"sync"
)

type Channel struct {
	Name    string
	api     string
	clients map[string]*Client
	m       sync.RWMutex
	cfg     *asyncapi3.Channel
	emitter engine.EventEmitter
	log     func(log *Log, traits events.Traits)
	monitor *monitor.Websocket
}

func (s *Store) Channel(name string) (*Channel, bool) {
	s.m.RLock()
	if t, ok := s.Channels[name]; ok {
		s.m.RUnlock()
		return t, ok
	}
	s.m.RUnlock()

	s.m.Lock()
	defer s.m.Unlock()

	for _, ref := range s.cfg.Channels {
		if ref == nil || ref.Value == nil {
			continue
		}
		ch := ref.Value
		if !ch.IsChannelAvailable("ws") {
			continue
		}
		if len(ch.Parameters) == 0 {
			continue
		}

		err := ch.IsNameValid(name)
		if err != nil {
			continue
		}

		if s.Channels == nil {
			s.Channels = make(map[string]*Channel)
		}

		c := &Channel{
			Name:    name,
			api:     s.cfg.Info.Name,
			emitter: s.emitter,
			cfg:     ch,
			log:     s.log,
			monitor: s.monitor,
		}
		s.Channels[name] = c
		return c, true
	}

	return nil, false
}
