package models

import (
	"runtime"
	"time"
)

type Metrics struct {
	Start             time.Time
	TotalRequests     int64
	RequestsWithError int64
	LastRequests      []*RequestMetric
	LastErrorRequests []*RequestMetric
	Memory            int64
	Kafka             map[string]*KafkaMetrics
}

type RequestMetric struct {
	Method       string
	Url          string
	HttpStatus   int
	Error        string
	ResponseTime time.Duration
	Time         time.Time
}

type KafkaMetrics struct {
	Topics     int
	Partitions int
	Segments   int
	Messages   int64
	TopicSizes map[string]int64
}

func NewRequestMetric(method string, url string) *RequestMetric {
	return &RequestMetric{
		Method: method,
		Url:    url,
		Time:   time.Now(),
	}
}

func newMetrics() *Metrics {
	return &Metrics{LastRequests: make([]*RequestMetric, 0), Start: time.Now(), Kafka: make(map[string]*KafkaMetrics)}
}

func (m *Metrics) AddRequest(r *RequestMetric) {
	m.TotalRequests++
	if len(r.Error) > 0 {
		m.RequestsWithError++
		if len(m.LastErrorRequests) > 10 {
			m.LastErrorRequests = m.LastErrorRequests[1:]
		}
		m.LastErrorRequests = append(m.LastErrorRequests, r)
	}
	if len(m.LastRequests) > 10 {
		m.LastRequests = m.LastRequests[1:]
	}
	r.ResponseTime = time.Now().Sub(r.Time)
	if r.HttpStatus == 0 {
		r.HttpStatus = 200
	}
	m.LastRequests = append(m.LastRequests, r)
}

func (m *Metrics) Update() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	m.Memory = int64(stats.Alloc)
}
