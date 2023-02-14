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
	Ldap  *Ldap  `json:"ldap"`
	Smtp  *Smtp  `json:"smtp"`

	metrics []metrics.Metric
}

type contextKey string

func New() *Monitor {
	startTime := metrics.NewGauge(metrics.WithFQName("app", "start_timestamp"))
	memoryUsage := metrics.NewGauge(metrics.WithFQName("app", "memory_usage_bytes"))

	kafkaMessage := metrics.NewCounterMap(
		metrics.WithFQName("kafka", "messages_total"),
		metrics.WithLabelNames("service", "topic"))
	kafkaLastMessage := metrics.NewGaugeMap(
		metrics.WithFQName("kafka", "message_timestamp"),
		metrics.WithLabelNames("service", "topic"))
	kafkaLag := metrics.NewGaugeMap(
		metrics.WithFQName("kafka", "consumer_group_lag"),
		metrics.WithLabelNames("service", "group", "topic", "partition"))

	ldapBind := metrics.NewCounterMap(
		metrics.WithFQName("ldap", "bind_total"),
		metrics.WithLabelNames("service"))
	ldapSearch := metrics.NewCounterMap(
		metrics.WithFQName("ldap", "search_total"),
		metrics.WithLabelNames("service"))
	ldapErrors := metrics.NewCounterMap(
		metrics.WithFQName("ldap", "search_total"),
		metrics.WithLabelNames("service"))
	ldapLastSearch := metrics.NewGaugeMap(
		metrics.WithFQName("ldap", "search_timestamp"),
		metrics.WithLabelNames("service"))

	smtpMails := metrics.NewCounterMap(
		metrics.WithFQName("smtp", "mails_total"),
		metrics.WithLabelNames("service"))

	h := NewHttp()
	m := &Monitor{
		StartTime:   startTime,
		MemoryUsage: memoryUsage,
		Http:        h,
		Kafka: &Kafka{
			Messages:    kafkaMessage,
			LastMessage: kafkaLastMessage,
			Lags:        kafkaLag,
		},
		Ldap: &Ldap{
			Bind:       ldapBind,
			Search:     ldapSearch,
			Errors:     ldapErrors,
			LastSearch: ldapLastSearch,
		},
		Smtp: &Smtp{
			Mails: smtpMails,
		},
		metrics: []metrics.Metric{
			startTime,
			memoryUsage,
			h.RequestCounter,
			h.RequestErrorCounter,
			h.LastRequest,
			kafkaMessage,
			kafkaLastMessage,
			kafkaLag,
			ldapBind,
			ldapSearch,
			ldapLastSearch,
			ldapErrors,
			smtpMails,
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

func (m *Monitor) FindAll(query ...metrics.QueryFunc) []metrics.Metric {
	q := &metrics.Query{}
	for _, qe := range query {
		qe(q)
	}

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
