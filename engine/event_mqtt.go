package engine

import (
	"cmp"
	"mokapi/engine/common"
	"slices"
	"sync"
)

type MqttEventHandler struct {
	Filter  common.MqttFilter
	Execute common.EventHandler
	Args    common.EventArgs
}

type MqttEventDispatcher struct {
	handlers map[string][]*MqttEventHandler
	mu       sync.RWMutex
}

func (e *MqttEventDispatcher) Clear(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.handlers, key)
}

func (e *MqttEventDispatcher) Has(key string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, ok := e.handlers[key]
	return ok
}

func (e *MqttEventDispatcher) Register(key string, handler *MqttEventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.handlers == nil {
		e.handlers = make(map[string][]*MqttEventHandler)
	}
	e.handlers[key] = append(e.handlers[key], handler)
}

func (sh *scriptHost) OnMqtt(filter common.MqttFilter, do common.EventHandler, args common.EventArgs) {
	addDefaultTags(&args, sh)
	h := &MqttEventHandler{
		Filter:  filter,
		Execute: do,
		Args:    args,
	}
	sh.engine.MqttEventDispatcher.Register(sh.name, h)
}

func (e *MqttEventDispatcher) EmitMqtt(message *common.MqttMessageEvent) []*common.Action {
	e.mu.RLock()
	var ehs []*MqttEventHandler
	for _, h := range e.handlers {
		ehs = append(ehs, h...)
	}
	e.mu.RUnlock()

	slices.SortStableFunc(ehs, func(a, b *MqttEventHandler) int { return -1 * cmp.Compare(a.Args.Priority, b.Args.Priority) })

	var result []*common.Action

	for _, eh := range ehs {
		if !eh.match(message) {
			continue
		}
		a := runEventHandler(eh.Execute, eh.Args, message)
		if a != nil {
			result = append(result, a)
		}
	}

	return result
}

func (h *MqttEventHandler) match(message *common.MqttMessageEvent) bool {
	return true
}
