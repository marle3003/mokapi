package models

import (
	runtime2 "mokapi/providers/workflow/runtime"
	"time"
)

type RequestMetric struct {
	Id           string
	Service      string
	Method       string
	Url          string
	HttpStatus   int
	IsError      bool
	ResponseTime time.Duration
	Time         time.Time
	Parameters   []RequestParamter
	ContentType  string
	ResponseBody string
	Actions      []*runtime2.WorkflowSummary
}

type RequestParamter struct {
	Name  string
	Type  string
	Value string
	Raw   string
}

type ServiceMetric struct {
	Name        string    `json:"name"`
	LastRequest time.Time `json:"lastRequest"`
	Requests    int       `json:"requests"`
	Errors      int       `json:"errors"`
}

func NewRequestMetric(method string, url string) *RequestMetric {
	return &RequestMetric{
		Id:     newId(10),
		Method: method,
		Url:    url,
		Time:   time.Now(),
	}
}

func (m *Metrics) AddRequest(r *RequestMetric) {
	m.TotalRequests++
	if s, ok := m.OpenApi[r.Service]; ok {
		s.Requests++
	}

	if r.IsError {
		m.RequestsWithError++
		if len(m.LastErrorRequests) > 10 {
			m.LastErrorRequests = m.LastErrorRequests[1:]
		}
		m.LastErrorRequests = append(m.LastErrorRequests, r)

		if s, ok := m.OpenApi[r.Service]; ok {
			s.Errors++
		}
	}
	if len(m.LastRequests) > 10 {
		m.LastRequests = m.LastRequests[1:]
	}
	r.ResponseTime = time.Now().Sub(r.Time)
	if r.HttpStatus == 0 {
		r.HttpStatus = 200
	}
	m.LastRequests = append(m.LastRequests, r)

	if s, ok := m.OpenApi[r.Service]; ok {
		s.LastRequest = r.Time
	}
}
