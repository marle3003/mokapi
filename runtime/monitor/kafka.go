package monitor

import (
	"context"
	"mokapi/runtime/logs"
	"mokapi/runtime/metrics"
)

type Kafka struct {
	*metrics.Kafka
	log []logs.KafkaMessage
}

func (m *Http) AppendKafka(log logs.HttpLog) {
	if len(m.log) == 10 {
		m.log = m.log[1:]
	}
	m.log = append(m.log, log)
}

func NewKafkaContext(ctx context.Context, kafka *Kafka) context.Context {
	return context.WithValue(ctx, "monitor", kafka)
}

func KafkaFromContext(ctx context.Context) (*Kafka, bool) {
	m, ok := ctx.Value("monitor").(*Kafka)
	return m, ok
}
