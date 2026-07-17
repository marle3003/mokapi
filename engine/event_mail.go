package engine

import (
	"cmp"
	"mokapi/engine/common"
	"mokapi/smtp"
	"slices"
	"sync"
)

type MailEventHandler struct {
	Filter  common.MailFilter
	Execute common.EventHandler
	Args    common.EventArgs
}

type MailEventDispatcher struct {
	handlers map[string][]*MailEventHandler
	mu       sync.RWMutex
}

func (e *MailEventDispatcher) Clear(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.handlers, key)
}

func (e *MailEventDispatcher) Has(key string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, ok := e.handlers[key]
	return ok
}

func (e *MailEventDispatcher) Register(key string, handler *MailEventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.handlers == nil {
		e.handlers = make(map[string][]*MailEventHandler)
	}
	e.handlers[key] = append(e.handlers[key], handler)
}

func (sh *scriptHost) OnMail(filter common.MailFilter, do common.EventHandler, args common.EventArgs) {
	addDefaultTags(&args, sh)
	h := &MailEventHandler{
		Filter:  filter,
		Execute: do,
		Args:    args,
	}
	sh.engine.MailEventDispatcher.Register(sh.name, h)
}

func (e *MailEventDispatcher) EmitSmtp(message *smtp.Message, status *smtp.Status) []*common.Action {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var ehs []*MailEventHandler
	for _, h := range e.handlers {
		ehs = append(ehs, h...)
	}
	slices.SortStableFunc(ehs, func(a, b *MailEventHandler) int { return -1 * cmp.Compare(a.Args.Priority, b.Args.Priority) })

	var result []*common.Action

	for _, eh := range ehs {
		if !eh.match(message) {
			continue
		}
		a := runEventHandler(eh.Execute, eh.Args, message, status)
		if a != nil {
			result = append(result, a)
		}
	}

	return result
}

func (h *MailEventHandler) match(message *smtp.Message) bool {
	return true
}
