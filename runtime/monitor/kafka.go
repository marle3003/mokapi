package monitor

import (
	"context"
	"mokapi/runtime/metrics"
)

var kafkaKey = contextKey("kafka")

type Kafka struct {
	Messages    *metrics.CounterMap
	LastMessage *metrics.GaugeMap
	Lags        *metrics.GaugeMap
}

func NewKafka() *Kafka {
	messages := metrics.NewCounterMap(
		metrics.WithFQName("kafka", "messages_total"),
		metrics.WithLabelNames("service", "topic"))
	lastMessage := metrics.NewGaugeMap(
		metrics.WithFQName("kafka", "message_timestamp"),
		metrics.WithLabelNames("service", "topic"))
	lag := metrics.NewGaugeMap(
		metrics.WithFQName("kafka", "consumer_group_lag"),
		metrics.WithLabelNames("service", "group", "topic", "partition"))

	return &Kafka{
		Messages:    messages,
		LastMessage: lastMessage,
		Lags:        lag,
	}
}

func (k *Kafka) Metrics() []metrics.Metric {
	return []metrics.Metric{k.Messages, k.LastMessage, k.Lags}
}

func NewKafkaContext(ctx context.Context, kafka *Kafka) context.Context {
	return context.WithValue(ctx, kafkaKey, kafka)
}

func KafkaFromContext(ctx context.Context) (*Kafka, bool) {
	m, ok := ctx.Value(kafkaKey).(*Kafka)
	return m, ok
}
