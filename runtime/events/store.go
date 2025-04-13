package events

import "sync"

const defaultSize = 20

type store struct {
	size   int
	events []Event
	traits Traits
	m      sync.RWMutex
}

func (s *store) Push(e Event) {
	s.m.Lock()
	defer s.m.Unlock()

	size := s.size
	if size == 0 {
		size = defaultSize
	}

	if len(s.events) == size {
		s.events = s.events[0 : len(s.events)-1]
	}
	// prepend
	s.events = append([]Event{e}, s.events...)
}

func (s *store) Events(traits Traits) []Event {
	s.m.RLock()
	defer s.m.RUnlock()

	var events []Event
	for _, e := range s.events {
		if e.Traits.Contains(traits) {
			events = append(events, e)
		}
	}
	return events
}
