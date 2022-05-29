package events

type store struct {
	size   int
	events []Event
	traits Traits
}

func (s *store) Push(e Event) {
	if len(s.events) == 10 {
		s.events = s.events[1:]
	}
	// prepend
	s.events = append([]Event{e}, s.events...)
}

func (s *store) Events(traits Traits) []Event {
	var events []Event
	for _, e := range s.events {
		if e.Traits.Includes(traits) {
			events = append(events, e)
		}
	}
	return events
}
