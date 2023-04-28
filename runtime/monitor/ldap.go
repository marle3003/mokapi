package monitor

import (
	"context"
	"mokapi/runtime/metrics"
)

var ldapKey = contextKey("ldap")

type Ldap struct {
	Bind       *metrics.CounterMap
	Search     *metrics.CounterMap
	Errors     *metrics.CounterMap
	LastSearch *metrics.GaugeMap
}

func NewLdap() *Ldap {
	bind := metrics.NewCounterMap(
		metrics.WithFQName("ldap", "bind_total"),
		metrics.WithLabelNames("service"))
	search := metrics.NewCounterMap(
		metrics.WithFQName("ldap", "search_total"),
		metrics.WithLabelNames("service"))
	errors := metrics.NewCounterMap(
		metrics.WithFQName("ldap", "search_total"),
		metrics.WithLabelNames("service"))
	lastSearch := metrics.NewGaugeMap(
		metrics.WithFQName("ldap", "search_timestamp"),
		metrics.WithLabelNames("service"))

	return &Ldap{
		Bind:       bind,
		Search:     search,
		Errors:     errors,
		LastSearch: lastSearch,
	}
}

func (l *Ldap) Metrics() []metrics.Metric {
	return []metrics.Metric{l.Bind, l.Search, l.Errors, l.LastSearch}
}

func NewLdapContext(ctx context.Context, ldap *Ldap) context.Context {
	return context.WithValue(ctx, ldapKey, ldap)
}

func LdapFromContext(ctx context.Context) (*Ldap, bool) {
	m, ok := ctx.Value(ldapKey).(*Ldap)
	return m, ok
}
