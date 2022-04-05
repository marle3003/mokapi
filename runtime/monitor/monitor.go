package monitor

import "mokapi/runtime/metrics"

type Monitor struct {
	Http  *Http
	Kafka *Kafka
}

func New() *Monitor {
	return &Monitor{
		Http: &Http{
			HttpMetrics: metrics.NewHttp(),
		},
		Kafka: &Kafka{
			Kafka: metrics.NewKafka(),
			log:   nil,
		},
	}
}
