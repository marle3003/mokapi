package http

type Listener func(request, response interface{}) bool

type EventListener struct {
	Event    *Event
	Listener Listener
}

type Event struct {
	Method string `js:"method"`
	Url    string `js:"url"`
}

type Http struct {
	Listeners []*EventListener
}

func New() *Http {
	return &Http{Listeners: make([]*EventListener, 0)}
}

func (h *Http) On(event *Event, listener Listener) {
	h.Listeners = append(h.Listeners, &EventListener{event, listener})
}
