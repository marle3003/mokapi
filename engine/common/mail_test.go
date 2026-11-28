package common_test

import (
	"mokapi/engine/common"
	"mokapi/smtp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func aSmtpMessage() *smtp.Message {
	return &smtp.Message{
		Server:                  "mail.foo.bar",
		Sender:                  &smtp.Address{Address: "sender@foo.bar"},
		From:                    []smtp.Address{{Address: "from@foo.bar"}},
		To:                      []smtp.Address{{Address: "to@foo.bar"}},
		Cc:                      []smtp.Address{{Address: "cc@foo.bar"}},
		MessageId:               "id@foo.bar",
		Date:                    time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		Subject:                 "hello",
		ContentType:             "text/plain",
		ContentTransferEncoding: "7bit",
		Body:                    "body",
		Attachments:             []smtp.Attachment{{Name: "a.txt"}},
	}
}

func TestMailEventMessage_CopiesContent(t *testing.T) {
	m := aSmtpMessage()

	e := common.NewMailEventMessage("Mail A", m)

	require.Equal(t, "Mail A", e.Api)
	require.Equal(t, "mail.foo.bar", e.Server)
	require.Equal(t, "sender@foo.bar", e.Sender.Address)
	require.Equal(t, m.Date, e.Time)
	require.Equal(t, "7bit", e.Encoding)
	require.Equal(t, "a.txt", e.Attachments[0].Name)
}

func TestMailEventMessage_ChangesStayInCopyUntilApplied(t *testing.T) {
	m := aSmtpMessage()
	e := common.NewMailEventMessage("Mail A", m)

	e.Subject = "changed"
	e.To[0].Address = "other@foo.bar"
	e.Sender.Address = "other-sender@foo.bar"
	require.Equal(t, "hello", m.Subject)
	require.Equal(t, "to@foo.bar", m.To[0].Address)
	require.Equal(t, "sender@foo.bar", m.Sender.Address)

	e.ApplyTo(m)
	require.Equal(t, "changed", m.Subject)
	require.Equal(t, "other@foo.bar", m.To[0].Address)
	require.Equal(t, "other-sender@foo.bar", m.Sender.Address)
}

func TestMailEventMessage_ApplyToKeepsServerAndWithoutSender(t *testing.T) {
	m := aSmtpMessage()
	m.Sender = nil
	e := common.NewMailEventMessage("Mail A", m)
	e.Server = "changed"

	e.ApplyTo(m)

	require.Equal(t, "mail.foo.bar", m.Server)
	require.Nil(t, m.Sender)
}
