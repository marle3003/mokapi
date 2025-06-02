package monitor

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/runtime/metrics"
	"testing"
)

func TestMqtt_Metrics_Messages(t *testing.T) {
	m := NewMqtt()
	m.Messages.WithLabel("service_a", "topic_a").Add(1)
	require.Equal(t, float64(1), m.Messages.Sum())
}

func TestMqtt_LastMessage(t *testing.T) {
	m := NewMqtt()
	m.LastMessage.WithLabel("service_a", "topic_a").Set(10)
	require.Equal(t, float64(10), m.LastMessage.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestMqtt_Metrics_Lags(t *testing.T) {
	m := NewMqtt()
	m.Lags.WithLabel("service_a", "group_a", "topic_a", "10").Set(10)
	require.Equal(t, float64(10), m.Lags.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestMqttContext(t *testing.T) {
	ctx := context.Background()
	h := New()
	ctx = NewMqttContext(ctx, h.Mqtt)
	result, ok := MqttFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, h.Mqtt, result)
}
