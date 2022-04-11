package monitor

import (
	"context"
	"mokapi/runtime/logs"
	"mokapi/runtime/metrics"
)

type Http struct {
	*metrics.HttpMetrics
	Log []*logs.HttpLog `json:"log"`
}

func (m *Http) AppendHttp(log *logs.HttpLog) {
	if len(m.Log) == 10 {
		m.Log = m.Log[1:]
	}
	m.Log = append(m.Log, log)
}

func NewHttpContext(ctx context.Context, http *Http) context.Context {
	return context.WithValue(ctx, "monitor", http)
}

func HttpFromContext(ctx context.Context) (*Http, bool) {
	m, ok := ctx.Value("monitor").(*Http)
	return m, ok
}
