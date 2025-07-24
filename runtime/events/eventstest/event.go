package eventstest

import "mokapi/runtime/events"

type Event struct {
	Name string
	Api  string `json:"api"`
}

func (e *Event) Title() string {
	return e.Name
}

type Handler struct {
	Events []events.Event
}

func (h *Handler) Push(data events.EventData, traits events.Traits) error {
	h.Events = append(h.Events, events.Event{Data: data, Traits: traits})
	return nil
}

func (h *Handler) GetEvents(traits events.Traits) []events.Event {
	var result []events.Event
	for _, e := range h.Events {
		if e.Traits.Match(traits) {
			result = append(result, e)
		}
	}
	return result
}
