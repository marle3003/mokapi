package engine

import (
	"cmp"
	"mokapi/engine/common"
	"mokapi/ldap"
	"slices"
	"sync"
)

type LdapEventHandler struct {
	Filter  common.LdapFilter
	Execute common.EventHandler
	Args    common.EventArgs
}

type LdapEventDispatcher struct {
	handlers map[string][]*LdapEventHandler
	mu       sync.RWMutex
}

func (e *LdapEventDispatcher) Clear(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.handlers, key)
}

func (e *LdapEventDispatcher) Has(key string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, ok := e.handlers[key]
	return ok
}

func (e *LdapEventDispatcher) Register(key string, handler *LdapEventHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.handlers == nil {
		e.handlers = make(map[string][]*LdapEventHandler)
	}
	e.handlers[key] = append(e.handlers[key], handler)
}

func (sh *scriptHost) OnLdap(filter common.LdapFilter, do common.EventHandler, args common.EventArgs) {
	addDefaultTags(&args, sh)
	h := &LdapEventHandler{
		Filter:  filter,
		Execute: do,
		Args:    args,
	}
	sh.engine.LdapEventDispatcher.Register(sh.name, h)
}

func (e *LdapEventDispatcher) EmitLdap(request *ldap.SearchRequest, response *ldap.SearchResponse) []*common.Action {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var ehs []*LdapEventHandler
	for _, h := range e.handlers {
		ehs = append(ehs, h...)
	}
	slices.SortStableFunc(ehs, func(a, b *LdapEventHandler) int { return -1 * cmp.Compare(a.Args.Priority, b.Args.Priority) })

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

func (h *LdapEventHandler) match(request *ldap.SearchRequest, response *ldap.SearchResponse) bool {
	return true
}
