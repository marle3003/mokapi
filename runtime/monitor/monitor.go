package monitor

import (
	"context"
	"mokapi/runtime/metrics"
	"mokapi/safe"
	"runtime"
	"time"
)

type Monitor struct {
	StartTime   *metrics.Gauge `json:"start_time"`
	MemoryUsage *metrics.Gauge `json:"memstats_alloc_bytes"`

	Http  *Http  `json:"http"`
	Kafka *Kafka `json:"kafka"`
}

func New() *Monitor {
	m := &Monitor{
		StartTime:   metrics.NewGauge("start_time"),
		MemoryUsage: metrics.NewGauge("memstats_alloc_bytes"),
		Http: &Http{
			Http: metrics.NewHttp(),
		},
		Kafka: &Kafka{
			Kafka: metrics.NewKafka(),
			log:   nil,
		},
	}
	m.StartTime.Set(float64(time.Now().Unix()))

	return m
}

func (m *Monitor) Start(pool *safe.Pool) {
	pool.Go(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(5) * time.Second):
				m.update()
			}
		}
	})
}

func (m *Monitor) update() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	m.MemoryUsage.Set(float64(stats.Alloc))
}
