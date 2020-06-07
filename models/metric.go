package models

import "time"

type Metrics struct {
	Start             time.Time
	TotalRequests     int64
	RequestsWithError int64
	LastRequests      []*RequestMetric
}

type RequestMetric struct {
	Method       string
	Url          string
	HttpStatus   int
	Error        string
	ResponseTime time.Duration
	Time         time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{LastRequests: make([]*RequestMetric, 0), Start: time.Now()}
}

func (m *Metrics) AddRequest(r *RequestMetric) {
	m.TotalRequests++
	if len(r.Error) > 0 {
		m.RequestsWithError++
	}
	if len(m.LastRequests) > 5 {
		m.LastRequests = m.LastRequests[1:]
	}
	r.Time = time.Now()
	m.LastRequests = append(m.LastRequests, r)
}
