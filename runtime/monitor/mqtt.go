package monitor

import (
	"context"
	"mokapi/runtime/metrics"
)

var mqttKey = contextKey("mqtt")

type Mqtt struct {
	Messages    *metrics.CounterMap
	LastMessage *metrics.GaugeMap
	Lags        *metrics.GaugeMap
}

func NewMqtt() *Mqtt {
	messages := metrics.NewCounterMap(
		metrics.WithFQName("mqtt", "messages_total"),
		metrics.WithLabelNames("service", "topic"))
	lastMessage := metrics.NewGaugeMap(
		metrics.WithFQName("mqtt", "message_timestamp"),
		metrics.WithLabelNames("service", "topic"))
	lag := metrics.NewGaugeMap(
		metrics.WithFQName("mqtt", "consumer_group_lag"),
		metrics.WithLabelNames("service", "group", "topic", "partition"))

	return &Mqtt{
		Messages:    messages,
		LastMessage: lastMessage,
		Lags:        lag,
	}
}

func (k *Mqtt) Metrics() []metrics.Metric {
	return []metrics.Metric{k.Messages, k.LastMessage, k.Lags}
}

func (k *Mqtt) Reset() {
	k.Messages.Reset()
	k.LastMessage.Reset()
	k.Lags.Reset()
}

func NewMqttContext(ctx context.Context, mqtt *Mqtt) context.Context {
	return context.WithValue(ctx, mqttKey, mqtt)
}

func MqttFromContext(ctx context.Context) (*Mqtt, bool) {
	m, ok := ctx.Value(mqttKey).(*Mqtt)
	return m, ok
}
