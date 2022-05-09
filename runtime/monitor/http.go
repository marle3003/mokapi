package monitor

import (
	"context"
	"mokapi/runtime/logs"
	"mokapi/runtime/metrics"
)

type Http struct {
	RequestCounter      *metrics.CounterMap
	RequestErrorCounter *metrics.CounterMap
	LastRequest         *metrics.GaugeMap
	Log                 []*logs.HttpLog `json:"log"`
}

func (m *Http) AppendHttp(log *logs.HttpLog) {
	if len(m.Log) == 10 {
		m.Log = m.Log[1:]
	}
	// prepend
	m.Log = append([]*logs.HttpLog{log}, m.Log...)
}

func NewHttpContext(ctx context.Context, http *Http) context.Context {
	return context.WithValue(ctx, "monitor", http)
}

func HttpFromContext(ctx context.Context) (*Http, bool) {
	m, ok := ctx.Value("monitor").(*Http)
	return m, ok
}
