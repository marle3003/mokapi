package monitor_test

import (
	"context"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKafka_Metrics_Messages(t *testing.T) {
	k := monitor.NewKafka()
	k.Messages.WithLabel("service_a", "topic_a").Add(1)
	require.Equal(t, float64(1), k.Messages.Sum(metrics.NewQuery()))
}

func TestKafka_LastMessage(t *testing.T) {
	k := monitor.NewKafka()
	k.LastMessage.WithLabel("service_a", "topic_a").Set(10)
	require.Equal(t, float64(10), k.LastMessage.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestKafka_Metrics_Lags(t *testing.T) {
	k := monitor.NewKafka()
	k.Lags.WithLabel("service_a", "group_a", "topic_a", "10").Set(10)
	require.Equal(t, float64(10), k.Lags.Value(metrics.NewQuery(metrics.ByLabel("service", "service_a"))))
}

func TestKafkaContext(t *testing.T) {
	ctx := context.Background()
	h := monitor.New()
	ctx = monitor.NewKafkaContext(ctx, h.Kafka)
	result, ok := monitor.KafkaFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, h.Kafka, result)
}
