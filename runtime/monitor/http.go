package monitor

import (
	"context"
	"mokapi/runtime/logs"
	"mokapi/runtime/metrics"
)

type Http struct {
	*metrics.HttpMetrics
	log []logs.HttpLog
}

func (m *Http) AppendHttp(log logs.HttpLog) {
	if len(m.log) == 10 {
		m.log = m.log[1:]
	}
	m.log = append(m.log, log)
}

func NewHttpContext(ctx context.Context, http *Http) context.Context {
	return context.WithValue(ctx, "monitor", http)
}

func HttpFromContext(ctx context.Context) (*Http, bool) {
	m, ok := ctx.Value("monitor").(*Http)
	return m, ok
}
