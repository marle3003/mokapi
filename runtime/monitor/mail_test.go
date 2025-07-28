package monitor

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSmtp_Metrics_Mails(t *testing.T) {
	s := NewMail()
	s.Mails.WithLabel("service_a", "sender_a").Add(1)
	require.Equal(t, float64(1), s.Mails.Sum())
}

func TestSmtpContext(t *testing.T) {
	ctx := context.Background()
	s := NewMail()
	ctx = NewSmtpContext(ctx, s)
	result, ok := SmtpFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, s, result)
}
