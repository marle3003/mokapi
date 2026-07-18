package engine

import (
	"cmp"
	"mokapi/engine/common"
	"slices"
	"sync"
)

type KafkaEventHandler struct {
	Filter  common.KafkaFilter
	Execute common.EventHandler
	Args    common.EventArgs
}

type KafkaEventDispatcher struct {
	handlers map[string][]*KafkaEventHandler
	mu       sync.RWMutex
}

func (e *KafkaEventDispatcher) Clear(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.handlers, key)
}

func (e *KafkaEventDispatcher) Has(key string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, ok := e.handlers[key]
	return ok
}

func (e *KafkaEventDispatcher) Register(key string, handler *KafkaEventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.handlers == nil {
		e.handlers = make(map[string][]*KafkaEventHandler)
	}
	e.handlers[key] = append(e.handlers[key], handler)
}

func (sh *scriptHost) OnKafka(filter common.KafkaFilter, do common.EventHandler, args common.EventArgs) {
	addDefaultTags(&args, sh)
	h := &KafkaEventHandler{
		Filter:  filter,
		Execute: do,
		Args:    args,
	}
	sh.engine.KafkaEventDispatcher.Register(sh.name, h)
}

func (e *KafkaEventDispatcher) EmitKafka(record *common.KafkaEventRecord) []*common.Action {
	e.mu.RLock()
	var ehs []*KafkaEventHandler
	for _, h := range e.handlers {
		ehs = append(ehs, h...)
	}
	e.mu.RUnlock()

	slices.SortStableFunc(ehs, func(a, b *KafkaEventHandler) int { return -1 * cmp.Compare(a.Args.Priority, b.Args.Priority) })

	var result []*common.Action

	for _, eh := range ehs {
		if !eh.match(record) {
			continue
		}
		a := runEventHandler(eh.Execute, eh.Args, record)
		if a != nil {
			result = append(result, a)
		}
	}

	return result
}

func (h *KafkaEventHandler) match(record *common.KafkaEventRecord) bool {
	return true
}
