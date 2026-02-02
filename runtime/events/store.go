package events

import "sync"

const defaultSize = 20

type store struct {
	size   int
	events []Event
	traits Traits
	m      sync.RWMutex
}

func (s *store) Push(e Event) *Event {
	s.m.Lock()
	defer s.m.Unlock()

	size := s.size
	if size == 0 {
		size = defaultSize
	}

	var removed *Event
	if len(s.events) == size {
		removed = &s.events[len(s.events)-1]
		s.events = s.events[0 : len(s.events)-1]
	}
	// prepend
	s.events = append([]Event{e}, s.events...)

	return removed
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

func (s *store) some(traits Traits) bool {
	if len(traits) == 0 {
		return true
	}
	for key, value := range traits {
		if s.traits.Has(key, value) {
			return true
		}
	}
	return true
}
