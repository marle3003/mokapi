package monitor

import (
	"context"
	"mokapi/runtime/metrics"
)

var mqttKey = contextKey("mqtt")

type Mqtt struct {
	Messages    *metrics.CounterMap
	LastMessage *metrics.GaugeMap
}

func NewMqtt() *Mqtt {
	messages := metrics.NewCounterMap(
		metrics.WithFQName("mqtt", "messages_total"),
		metrics.WithLabelNames("service", "topic"))
	lastMessage := metrics.NewGaugeMap(
		metrics.WithFQName("mqtt", "message_timestamp"),
		metrics.WithLabelNames("service", "topic"))

	return &Mqtt{
		Messages:    messages,
		LastMessage: lastMessage,
	}
}

func (k *Mqtt) Metrics() []metrics.Metric {
	return []metrics.Metric{k.Messages, k.LastMessage}
}

func (k *Mqtt) Reset() {
	k.Messages.Reset()
	k.LastMessage.Reset()
}

func NewMqttContext(ctx context.Context, mqtt *Mqtt) context.Context {
	return context.WithValue(ctx, mqttKey, mqtt)
}

func MqttFromContext(ctx context.Context) (*Mqtt, bool) {
	m, ok := ctx.Value(mqttKey).(*Mqtt)
	return m, ok
}
