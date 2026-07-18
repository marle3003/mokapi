package engine

import (
	"cmp"
	"mokapi/engine/common"
	"slices"
	"sync"
)

type HttpEventHandler struct {
	Filter  common.HttpFilter
	Execute common.EventHandler
	Args    common.EventArgs
}

type HttpEventDispatcher struct {
	handlers map[string][]*HttpEventHandler
	mu       sync.RWMutex
}

func (e *HttpEventDispatcher) Clear(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.handlers, key)
}

func (e *HttpEventDispatcher) Has(key string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, ok := e.handlers[key]
	return ok
}

func (e *HttpEventDispatcher) Register(key string, handler *HttpEventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.handlers == nil {
		e.handlers = make(map[string][]*HttpEventHandler)
	}
	e.handlers[key] = append(e.handlers[key], handler)
}

func (sh *scriptHost) OnHttp(filter common.HttpFilter, do common.EventHandler, args common.EventArgs) {
	addDefaultTags(&args, sh)
	h := &HttpEventHandler{
		Filter:  filter,
		Execute: do,
		Args:    args,
	}
	sh.engine.HttpEventDispatcher.Register(sh.name, h)
}

func (e *HttpEventDispatcher) EmitHttp(request *common.HttpEventRequest, response *common.HttpEventResponse) []*common.Action {
	e.mu.RLock()
	var ehs []*HttpEventHandler
	for _, h := range e.handlers {
		ehs = append(ehs, h...)
	}
	e.mu.RUnlock()

	slices.SortStableFunc(ehs, func(a, b *HttpEventHandler) int { return -1 * cmp.Compare(a.Args.Priority, b.Args.Priority) })

	var result []*common.Action

	for _, eh := range ehs {
		if !eh.match(request, response) {
			continue
		}
		a := runEventHandler(eh.Execute, eh.Args, request, response)
		if a != nil {
			result = append(result, a)
		}
	}

	return result
}

func (h *HttpEventHandler) match(request *common.HttpEventRequest, response *common.HttpEventResponse) bool {
	return true
}
