package events

import "sync"

type store struct {
	size   int
	events []Event
	traits Traits
	m      sync.Mutex
}

func (s *store) Push(e Event) {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.events) == s.size {
		s.events = s.events[1:]
	}
	// prepend
	s.events = append([]Event{e}, s.events...)
}

func (s *store) Events(traits Traits) []Event {
	var events []Event
	for _, e := range s.events {
		if e.Traits.Contains(traits) {
			events = append(events, e)
		}
	}
	return events
}
