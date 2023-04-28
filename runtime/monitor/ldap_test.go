package monitor

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/runtime/metrics"
	"testing"
)

func TestLdap_Metrics_Bind(t *testing.T) {
	l := NewLdap()
	l.Bind.WithLabel("service_a").Add(1)
	require.Equal(t, float64(1), l.Bind.Sum())
}

func TestLdap_Search(t *testing.T) {
	l := NewLdap()
	l.Search.WithLabel("service_a").Add(10)
	require.Equal(t, float64(10), l.Search.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestLdap_Metrics_Errors(t *testing.T) {
	l := NewLdap()
	l.Errors.WithLabel("service_a").Add(10)
	require.Equal(t, float64(10), l.Errors.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestLdap_LastSearch(t *testing.T) {
	l := NewLdap()
	l.LastSearch.WithLabel("service_a").Set(10)
	require.Equal(t, float64(10), l.LastSearch.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestLdapContext(t *testing.T) {
	ctx := context.Background()
	h := New()
	ctx = NewLdapContext(ctx, h.Ldap)
	result, ok := LdapFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, h.Ldap, result)
}
