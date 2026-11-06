package eventstest

import (
	"mokapi/runtime/events"
	"sync"
)

type Event struct {
	Name string `json:"name"`
	Api  string `json:"api"`
}

func (e *Event) Title() string {
	return e.Name
}

func (e *Event) IndexFields() map[string]any {
	return map[string]any{
		"name": e.Name,
	}
}

type Handler struct {
	Events []events.Event
	m      sync.Mutex
}

func (h *Handler) Push(data events.EventData, traits events.Traits) error {
	h.m.Lock()
	defer h.m.Unlock()
	h.Events = append(h.Events, events.Event{Data: data, Traits: traits})
	return nil
}

func (h *Handler) GetEvents(traits events.Traits) []events.Event {
	h.m.Lock()
	defer h.m.Unlock()
	var result []events.Event
	for _, e := range h.Events {
		if traits.Match(e.Traits) {
			result = append(result, e)
		}
	}
	return result
}
