package monitor

import (
	"context"
	"mokapi/runtime/metrics"
)

var smtpKey = contextKey("smtp_monitor")

type Smtp struct {
	Mails    *metrics.CounterMap
	LastMail *metrics.GaugeMap
}

func NewSmtp() *Smtp {
	mails := metrics.NewCounterMap(
		metrics.WithFQName("smtp", "mails_total"),
		metrics.WithLabelNames("service"))
	lastMail := metrics.NewGaugeMap(
		metrics.WithFQName("smtp", "mail_timestamp"),
		metrics.WithLabelNames("service"))

	return &Smtp{Mails: mails, LastMail: lastMail}
}

func (s *Smtp) Metrics() []metrics.Metric {
	return []metrics.Metric{s.Mails, s.LastMail}
}

func (s *Smtp) Reset() {
	s.Mails.Reset()
	s.LastMail.Reset()
}

func NewSmtpContext(ctx context.Context, smtp *Smtp) context.Context {
	return context.WithValue(ctx, smtpKey, smtp)
}

func SmtpFromContext(ctx context.Context) (*Smtp, bool) {
	m, ok := ctx.Value(smtpKey).(*Smtp)
	return m, ok
}
