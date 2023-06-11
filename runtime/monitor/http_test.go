package monitor

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/runtime/metrics"
	"testing"
)

func TestHttp_Metrics_Request_Total(t *testing.T) {
	h := NewHttp()
	h.RequestCounter.WithLabel("service_a", "endpoint_a").Add(1)
	require.Equal(t, float64(1), h.RequestCounter.Sum())
}

func TestHttp_Metrics_Request_Errors_Total(t *testing.T) {
	h := NewHttp()
	h.RequestErrorCounter.WithLabel("service_a", "endpoint_a").Add(1)
	require.Equal(t, float64(1), h.RequestErrorCounter.Sum())
}

func TestHttp_Metrics_LastRequest(t *testing.T) {
	h := NewHttp()
	h.LastRequest.WithLabel("service_a", "endpoint_a").Set(10)
	require.Equal(t, float64(10), h.LastRequest.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestHttpContext(t *testing.T) {
	ctx := context.Background()
	h := NewHttp()
	ctx = NewHttpContext(ctx, h)
	result, ok := HttpFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, h, result)
}
