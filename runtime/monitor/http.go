package monitor

import (
	"context"
	"mokapi/runtime/metrics"
)

type Http struct {
	RequestCounter      *metrics.CounterMap
	RequestErrorCounter *metrics.CounterMap
	LastRequest         *metrics.GaugeMap
}

func NewHttpContext(ctx context.Context, http *Http) context.Context {
	return context.WithValue(ctx, "monitor", http)
}

func HttpFromContext(ctx context.Context) (*Http, bool) {
	m, ok := ctx.Value("monitor").(*Http)
	return m, ok
}
