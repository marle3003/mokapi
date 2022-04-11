package monitor

import (
	"context"
	"mokapi/runtime/metrics"
	"mokapi/safe"
	"runtime"
	"time"
)

type Monitor struct {
	StartTime   time.Time `json:"startTime"`
	MemoryUsage uint64    `json:"memoryUsage"`

	Http  *Http  `json:"http"`
	Kafka *Kafka `json:"kafka"`
}

func New() *Monitor {
	return &Monitor{
		StartTime: time.Now(),
		Http: &Http{
			HttpMetrics: metrics.NewHttp(),
		},
		Kafka: &Kafka{
			Kafka: metrics.NewKafka(),
			log:   nil,
		},
	}
}

func (m *Monitor) Start(pool *safe.Pool) {
	pool.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(5)):
				m.update()
			}
		}
	})
}

func (m *Monitor) update() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	m.MemoryUsage = stats.Alloc
}
