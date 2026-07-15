package runtime

import (
	"mokapi/engine/common"
)

type EventHandlers struct {
	handlers map[string][]HTTPHandler
}

type HTTPHandler struct {
	Filter  common.HTTPFilter
	Execute common.HttpEventFunc
}

func (e *EventHandlers) AddHttpEventHandlers(key string, filter common.HTTPFilter, execute common.HttpEventFunc) {
	if e.handlers == nil {
		e.handlers = make(map[string][]HTTPHandler)
	}
	e.handlers[key] = append(e.handlers[key], HTTPHandler{Filter: filter, Execute: execute})
}

func (e *EventHandlers) Remove(key string) {
	delete(e.handlers, key)
}

func (h *HTTPHandler) match(request common.HttpEventRequest, response common.HttpEventResponse) bool {
	return true
}
