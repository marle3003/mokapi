package runtime

import (
	"context"
	"mokapi/config/dynamic/directory"
	"mokapi/ldap"
	"mokapi/runtime/monitor"
	"time"
)

type LdapInfo struct {
	*directory.Config
}

type LdapHandler struct {
	ldap *monitor.Ldap
	next ldap.Handler
}

func NewLdapHandler(ldap *monitor.Ldap, next ldap.Handler) *LdapHandler {
	return &LdapHandler{ldap: ldap, next: next}
}

func (h *LdapHandler) ServeLDAP(rw ldap.ResponseWriter, r *ldap.Request) {
	r.Context = monitor.NewLdapContext(r.Context, h.ldap)
	r.Context = context.WithValue(r.Context, "time", time.Now())

	h.next.ServeLDAP(rw, r)

}
