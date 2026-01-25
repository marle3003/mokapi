package monitor

import (
	"context"
	"mokapi/runtime/metrics"
)

var kafkaKey = contextKey("kafka")

type Kafka struct {
	Messages        *metrics.CounterMap
	LastMessage     *metrics.GaugeMap
	Lags            *metrics.GaugeMap
	Commits         *metrics.GaugeMap
	LastRebalancing *metrics.GaugeMap
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

	commits := metrics.NewGaugeMap(
		metrics.WithFQName("kafka", "consumer_group_commit"),
		metrics.WithLabelNames("service", "group", "topic", "partition"))
	lastRebalancing :=
		metrics.NewGaugeMap(
			metrics.WithFQName("kafka", "rebalance_timestamp"),
			metrics.WithLabelNames("service", "group"),
		)

	return &Kafka{
		Messages:        messages,
		LastMessage:     lastMessage,
		Lags:            lag,
		LastRebalancing: lastRebalancing,
		Commits:         commits,
	}
}

func (k *Kafka) Metrics() []metrics.Metric {
	return []metrics.Metric{k.Messages, k.LastMessage, k.Lags, k.Commits, k.LastRebalancing}
}

func (k *Kafka) Reset() {
	k.Messages.Reset()
	k.LastMessage.Reset()
	k.Lags.Reset()
	k.Commits.Reset()
	k.LastRebalancing.Reset()
}

func NewKafkaContext(ctx context.Context, kafka *Kafka) context.Context {
	return context.WithValue(ctx, kafkaKey, kafka)
}

func KafkaFromContext(ctx context.Context) (*Kafka, bool) {
	m, ok := ctx.Value(kafkaKey).(*Kafka)
	return m, ok
}
