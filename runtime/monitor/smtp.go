package monitor

import (
	"context"
	"mokapi/runtime/metrics"
)

type Smtp struct {
	Mails *metrics.CounterMap
}

func NewSmtpContext(ctx context.Context, smtp *Smtp) context.Context {
	return context.WithValue(ctx, "monitor", smtp)
}

func SmtpFromContext(ctx context.Context) (*Smtp, bool) {
	m, ok := ctx.Value("monitor").(*Smtp)
	return m, ok
}
