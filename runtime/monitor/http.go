package monitor

import (
	"context"
	"mokapi/runtime/metrics"
)

var httpKey = contextKey("http")

type Http struct {
	RequestCounter      *metrics.CounterMap
	RequestErrorCounter *metrics.CounterMap
	LastRequest         *metrics.GaugeMap
}

func NewHttp() *Http {
	httpRequestCounter := metrics.NewCounterMap(
		metrics.WithFQName("http", "requests_total"),
		metrics.WithLabelNames("service", "endpoint"))
	httpRequestErrorCounter := metrics.NewCounterMap(
		metrics.WithFQName("http", "requests_errors_total"),
		metrics.WithLabelNames("service", "endpoint"))
	httpLastRequest := metrics.NewGaugeMap(
		metrics.WithFQName("http", "request_timestamp"),
		metrics.WithLabelNames("service", "endpoint"))

	return &Http{
		RequestCounter:      httpRequestCounter,
		RequestErrorCounter: httpRequestErrorCounter,
		LastRequest:         httpLastRequest,
	}
}

func (h *Http) Metrics() []metrics.Metric {
	return []metrics.Metric{h.RequestCounter, h.RequestErrorCounter, h.LastRequest}
}

func NewHttpContext(ctx context.Context, http *Http) context.Context {
	return context.WithValue(ctx, httpKey, http)
}

func HttpFromContext(ctx context.Context) (*Http, bool) {
	m, ok := ctx.Value(httpKey).(*Http)
	return m, ok
}
