package monitor

import (
	"context"
	"mokapi/runtime/logs"
	"mokapi/runtime/metrics"
)

type Kafka struct {
	Messages    *metrics.CounterMap
	LastMessage *metrics.GaugeMap
	Lags        *metrics.GaugeMap
	Log         []*logs.KafkaMessage
}

func (m *Kafka) AppendKafka(log *logs.KafkaMessage) {
	if len(m.Log) == 10 {
		m.Log = m.Log[1:]
	}
	m.Log = append(m.Log, log)
}

func NewKafkaContext(ctx context.Context, kafka *Kafka) context.Context {
	return context.WithValue(ctx, "monitor", kafka)
}

func KafkaFromContext(ctx context.Context) (*Kafka, bool) {
	m, ok := ctx.Value("monitor").(*Kafka)
	return m, ok
}
