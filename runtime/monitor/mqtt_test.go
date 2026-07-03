package monitor_test

import (
	"context"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMqtt_Metrics_Messages(t *testing.T) {
	m := monitor.NewMqtt()
	m.Messages.WithLabel("service_a", "topic_a").Add(1)
	require.Equal(t, float64(1), m.Messages.Sum(metrics.NewQuery()))
}

func TestMqtt_LastMessage(t *testing.T) {
	m := monitor.NewMqtt()
	m.LastMessage.WithLabel("service_a", "topic_a").Set(10)
	require.Equal(t, float64(10), m.LastMessage.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestMqttContext(t *testing.T) {
	ctx := context.Background()
	h := monitor.New()
	ctx = monitor.NewMqttContext(ctx, h.Mqtt)
	result, ok := monitor.MqttFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, h.Mqtt, result)
}
