package monitor

import (
	"context"
	"mokapi/runtime/metrics"
)

var ldapKey = contextKey("ldap")

type Ldap struct {
	Errors         *metrics.CounterMap
	RequestCounter *metrics.CounterMap
	LastRequest    *metrics.GaugeMap
}

func NewLdap() *Ldap {
	requests := metrics.NewCounterMap(
		metrics.WithFQName("ldap", "request_total"),
		metrics.WithLabelNames("service", "operation"))
	errors := metrics.NewCounterMap(
		metrics.WithFQName("ldap", "search_errors_total"),
		metrics.WithLabelNames("service"))
	lastRequest := metrics.NewGaugeMap(
		metrics.WithFQName("ldap", "request_timestamp"),
		metrics.WithLabelNames("service"))

	return &Ldap{
		RequestCounter: requests,
		Errors:         errors,
		LastRequest:    lastRequest,
	}
}

func (l *Ldap) Metrics() []metrics.Metric {
	return []metrics.Metric{l.RequestCounter, l.Errors, l.LastRequest}
}

func (l *Ldap) Reset() {
	l.RequestCounter.Reset()
	l.Errors.Reset()
	l.LastRequest.Reset()
}

func NewLdapContext(ctx context.Context, ldap *Ldap) context.Context {
	return context.WithValue(ctx, ldapKey, ldap)
}

func LdapFromContext(ctx context.Context) (*Ldap, bool) {
	m, ok := ctx.Value(ldapKey).(*Ldap)
	return m, ok
}
