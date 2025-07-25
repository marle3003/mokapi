package monitor

import (
	"context"
	"mokapi/runtime/metrics"
)

var smtpKey = contextKey("mail_monitor")

type Mail struct {
	Mails    *metrics.CounterMap
	LastMail *metrics.GaugeMap
}

func NewMail() *Mail {
	mails := metrics.NewCounterMap(
		metrics.WithFQName("mail", "mails_total"),
		metrics.WithLabelNames("service", "sender"),
	)
	lastMail := metrics.NewGaugeMap(
		metrics.WithFQName("mail", "mail_timestamp"),
		metrics.WithLabelNames("service"))

	return &Mail{Mails: mails, LastMail: lastMail}
}

func (s *Mail) Metrics() []metrics.Metric {
	return []metrics.Metric{s.Mails, s.LastMail}
}

func (s *Mail) Reset() {
	s.Mails.Reset()
	s.LastMail.Reset()
}

func NewSmtpContext(ctx context.Context, smtp *Mail) context.Context {
	return context.WithValue(ctx, smtpKey, smtp)
}

func SmtpFromContext(ctx context.Context) (*Mail, bool) {
	m, ok := ctx.Value(smtpKey).(*Mail)
	return m, ok
}
