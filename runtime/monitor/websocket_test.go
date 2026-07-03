package monitor_test

import (
	"context"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWebsocket_Metrics_Messages(t *testing.T) {
	m := monitor.NewWebsocket()
	m.Messages.WithLabel("service_a", "channel_a").Add(1)
	require.Equal(t, float64(1), m.Messages.Sum(metrics.NewQuery()))
}

func TestWebsocket_LastMessage(t *testing.T) {
	m := monitor.NewWebsocket()
	m.LastMessage.WithLabel("service_a", "channel_a").Set(10)
	require.Equal(t, float64(10), m.LastMessage.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestWebsocketContext(t *testing.T) {
	ctx := context.Background()
	h := monitor.New()
	ctx = monitor.NewWebsocketContext(ctx, h.Websocket)
	result, ok := monitor.WebsocketFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, h.Websocket, result)
}
