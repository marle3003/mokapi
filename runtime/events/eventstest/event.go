package eventstest

import "mokapi/runtime/events"

type Event struct {
	Name string
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
}

func (h *Handler) Push(data events.EventData, traits events.Traits) error {
	h.Events = append(h.Events, events.Event{Data: data, Traits: traits})
	return nil
}

func (h *Handler) GetEvents(traits events.Traits) []events.Event {
	var result []events.Event
	for _, e := range h.Events {
		if traits.Match(e.Traits) {
			result = append(result, e)
		}
	}
	return result
}
