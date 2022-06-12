package monitor

import (
	"context"
	"mokapi/runtime/metrics"
)

type Ldap struct {
	Bind       *metrics.CounterMap
	Search     *metrics.CounterMap
	Errors     *metrics.CounterMap
	LastSearch *metrics.GaugeMap
}

var ldapKey = contextKey("ldap_monitor")

func NewLdapContext(ctx context.Context, ldap *Ldap) context.Context {
	return context.WithValue(ctx, ldapKey, ldap)
}

func LdapFromContext(ctx context.Context) (*Ldap, bool) {
	m, ok := ctx.Value(ldapKey).(*Ldap)
	return m, ok
}
