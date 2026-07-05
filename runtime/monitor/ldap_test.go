package monitor_test

import (
	"context"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLdapLabels(t *testing.T) {
	l := monitor.NewLdap()
	require.Equal(t, "ldap_search_errors_total", l.Errors.Info().String())
	require.Equal(t, "ldap_requests_total", l.RequestCounter.Info().String())
	require.Equal(t, "ldap_request_timestamp", l.LastRequest.Info().String())
}

func TestLdap_Metrics_Bind(t *testing.T) {
	l := monitor.NewLdap()
	l.RequestCounter.WithLabel("service_a", "bind").Add(1)
	require.Equal(t, float64(1), l.RequestCounter.Sum(metrics.NewQuery()))
}

func TestLdap_Search(t *testing.T) {
	l := monitor.NewLdap()
	l.RequestCounter.WithLabel("service_a", "search").Add(10)
	require.Equal(t, float64(10), l.RequestCounter.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestLdap_Metrics_Errors(t *testing.T) {
	l := monitor.NewLdap()
	l.Errors.WithLabel("service_a").Add(10)
	require.Equal(t, float64(10), l.Errors.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestLdap_LastSearch(t *testing.T) {
	l := monitor.NewLdap()
	l.LastRequest.WithLabel("service_a").Set(10)
	require.Equal(t, float64(10), l.LastRequest.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestLdapContext(t *testing.T) {
	ctx := context.Background()
	h := monitor.New()
	ctx = monitor.NewLdapContext(ctx, h.Ldap)
	result, ok := monitor.LdapFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, h.Ldap, result)
}
