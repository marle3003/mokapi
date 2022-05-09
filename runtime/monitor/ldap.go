package monitor

import (
	"context"
	"mokapi/runtime/logs"
	"mokapi/runtime/metrics"
)

type Ldap struct {
	RequestCounter      *metrics.CounterMap
	RequestErrorCounter *metrics.CounterMap
	log                 []logs.LdapLog
}

func (m *Ldap) AppendLdap(log logs.LdapLog) {
	if len(m.log) == 10 {
		m.log = m.log[1:]
	}
	m.log = append(m.log, log)
}

func NewLdapContext(ctx context.Context, ldap *Ldap) context.Context {
	return context.WithValue(ctx, "monitor", ldap)
}

func LdapFromContext(ctx context.Context) (*Ldap, bool) {
	m, ok := ctx.Value("monitor").(*Ldap)
	return m, ok
}
