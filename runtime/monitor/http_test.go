package monitor_test

import (
	"context"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHttp_Metrics_Request_Total(t *testing.T) {
	h := monitor.NewHttp()
	h.RequestCounter.WithLabel("service_a", "endpoint_a", "post").Add(1)
	require.Equal(t, float64(1), h.RequestCounter.Sum(metrics.NewQuery()))
}

func TestHttp_Metrics_Request_Errors_Total(t *testing.T) {
	h := monitor.NewHttp()
	h.RequestErrorCounter.WithLabel("service_a", "endpoint_a", "put").Add(1)
	require.Equal(t, float64(1), h.RequestErrorCounter.Sum(metrics.NewQuery()))
}

func TestHttp_Metrics_LastRequest(t *testing.T) {
	h := monitor.NewHttp()
	h.LastRequest.WithLabel("service_a", "endpoint_a", "delete").Set(10)
	require.Equal(t, float64(10), h.LastRequest.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestHttpContext(t *testing.T) {
	ctx := context.Background()
	h := monitor.NewHttp()
	ctx = monitor.NewHttpContext(ctx, h)
	result, ok := monitor.HttpFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, h, result)
}
