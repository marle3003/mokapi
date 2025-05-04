package monitor

import (
	"context"
	"mokapi/runtime/metrics"
	"mokapi/safe"
	"runtime"
	"time"
)

type Monitor struct {
	StartTime   *metrics.Gauge   `json:"start_time"`
	MemoryUsage *metrics.Gauge   `json:"memstats_alloc_bytes"`
	JobCounter  *metrics.Counter `json:"job_counter"`

	Http  *Http  `json:"http"`
	Kafka *Kafka `json:"kafka"`
	Ldap  *Ldap  `json:"ldap"`
	Smtp  *Smtp  `json:"smtp"`

	RefreshRateSeconds int

	metrics []metrics.Metric
}

type contextKey string

func New() *Monitor {
	startTime := metrics.NewGauge(metrics.WithFQName("app", "start_timestamp"))
	memoryUsage := metrics.NewGauge(metrics.WithFQName("app", "memory_usage_bytes"))
	jobCounter := metrics.NewCounter(metrics.WithFQName("app", "job_run_total"))

	h := NewHttp()
	k := NewKafka()
	l := NewLdap()
	s := NewSmtp()

	collection := []metrics.Metric{
		startTime,
		memoryUsage,
		jobCounter,
	}
	collection = append(collection, h.Metrics()...)
	collection = append(collection, k.Metrics()...)
	collection = append(collection, l.Metrics()...)
	collection = append(collection, s.Metrics()...)

	m := &Monitor{
		RefreshRateSeconds: 5,
		StartTime:          startTime,
		MemoryUsage:        memoryUsage,
		JobCounter:         jobCounter,
		Http:               h,
		Kafka:              k,
		Ldap:               l,
		Smtp:               s,
		metrics:            collection,
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
			case <-time.After(time.Duration(m.RefreshRateSeconds) * time.Second):
				m.update()
			}
		}
	})
}

func (m *Monitor) Reset() {
	m.Http.Reset()
	m.Kafka.Reset()
	m.Smtp.Reset()
	m.Ldap.Reset()
}

func (m *Monitor) update() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	m.MemoryUsage.Set(float64(stats.Alloc))
}

func (m *Monitor) FindAll(query ...metrics.QueryOptions) []metrics.Metric {
	q := metrics.NewQuery(query...)

	ch := make(chan metrics.Metric)
	result := make([]metrics.Metric, 0)

	go func() {
		for _, metric := range m.metrics {
			metric.Collect(ch)
		}
		close(ch)
	}()

	for metric := range ch {
		if metric.Info().Match(q) {
			result = append(result, metric)
		}
	}

	return result
}

func (c contextKey) String() string {
	return "monitor context key " + string(c)
}
