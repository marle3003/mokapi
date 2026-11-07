package engine

import (
	"mokapi/engine/common"
	"mokapi/ldap"
)

type LdapEventHandler = eventHandler[common.LdapFilter]

type LdapEventDispatcher struct {
	dispatcher[common.LdapFilter]
}

func (sh *scriptHost) OnLdap(filter common.LdapFilter, do common.EventHandlerFunc, args common.EventArgs) {
	sh.engine.LdapEventDispatcher.Register(sh.name, newEventHandler(sh, "ldap", filter, do, args))
}

func (e *LdapEventDispatcher) EmitLdap(request *ldap.SearchRequest, response *ldap.SearchResponse) []*common.Action {
	return e.emit(emitOptions[common.LdapFilter]{}, request, response)
}
