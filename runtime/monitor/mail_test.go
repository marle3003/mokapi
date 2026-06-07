package monitor

import (
	"context"
	"mokapi/runtime/metrics"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSmtp_Metrics_Mails(t *testing.T) {
	s := NewMail()
	s.Mails.WithLabel("service_a", "sender_a").Add(1)
	require.Equal(t, float64(1), s.Mails.Sum(metrics.NewQuery()))
}

func TestSmtpContext(t *testing.T) {
	ctx := context.Background()
	s := NewMail()
	ctx = NewSmtpContext(ctx, s)
	result, ok := SmtpFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, s, result)
}
