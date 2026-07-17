package engine

import (
	"cmp"
	"mokapi/engine/common"
	"slices"
	"sync"
)

type WebsocketEventHandler struct {
	Filter  common.WebsocketFilter
	Execute common.EventHandler
	Args    common.EventArgs
}

type WebsocketEventDispatcher struct {
	handlers map[string][]*WebsocketEventHandler
	mu       sync.RWMutex
}

func (e *WebsocketEventDispatcher) Clear(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.handlers, key)
}

func (e *WebsocketEventDispatcher) Has(key string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, ok := e.handlers[key]
	return ok
}

func (e *WebsocketEventDispatcher) EmitWebsocketConnect(event *common.WebsocketConnectEvent) []*common.Action {
	return e.run(event, func(h *WebsocketEventHandler) bool {
		if h.Filter.Type != "" && h.Filter.Type != common.WebsocketConnectEventType {
			return false
		}
		if !h.match(event.WebsocketEvent) {
			return false
		}
		return true
	})
}

func (e *WebsocketEventDispatcher) EmitWebsocketClose(event *common.WebsocketCloseEvent) []*common.Action {
	return e.run(event, func(h *WebsocketEventHandler) bool {
		if h.Filter.Type != "" && h.Filter.Type != common.WebsocketCloseEventType {
			return false
		}
		if !h.match(event.WebsocketEvent) {
			return false
		}
		return true
	})
}

func (e *WebsocketEventDispatcher) EmitWebsocketMessage(event *common.WebsocketMessageEvent) []*common.Action {
	return e.run(event, func(h *WebsocketEventHandler) bool {
		if h.Filter.Type != "" && h.Filter.Type != common.WebsocketMessageEventType {
			return false
		}
		if !h.match(event.WebsocketEvent) {
			return false
		}
		return true
	})
}

func (e *WebsocketEventDispatcher) EmitWebsocket(event any) []*common.Action {
	return e.run(event, func(handler *WebsocketEventHandler) bool {
		return true
	})
}

func (e *WebsocketEventDispatcher) Register(key string, handler *WebsocketEventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.handlers == nil {
		e.handlers = make(map[string][]*WebsocketEventHandler)
	}
	e.handlers[key] = append(e.handlers[key], handler)
}

func (sh *scriptHost) OnWebsocket(filter common.WebsocketFilter, do common.EventHandler, args common.EventArgs) {
	addDefaultTags(&args, sh)
	h := &WebsocketEventHandler{
		Filter:  filter,
		Execute: do,
		Args:    args,
	}
	sh.engine.WebsocketEventDispatcher.Register(sh.name, h)
}

func (h *WebsocketEventHandler) match(record common.WebsocketEvent) bool {
	return true
}

func (e *WebsocketEventDispatcher) run(event any, condition func(*WebsocketEventHandler) bool) []*common.Action {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var ehs []*WebsocketEventHandler
	for _, list := range e.handlers {
		for _, h := range list {
			if condition(h) {
				ehs = append(ehs, h)
			}
		}
	}
	slices.SortStableFunc(ehs, func(a, b *WebsocketEventHandler) int { return -1 * cmp.Compare(a.Args.Priority, b.Args.Priority) })

	var result []*common.Action
	for _, eh := range ehs {
		a := runEventHandler(eh.Execute, eh.Args, event)
		if a != nil {
			result = append(result, a)
		}
	}
	return result
}
