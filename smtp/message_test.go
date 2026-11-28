package smtp_test

import (
	"mokapi/smtp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDecodeHeaderValue(t *testing.T) {
	testcases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty",
			in:   "",
			want: "",
		},
		{
			name: "utf-8",
			in:   "=?UTF-8?Q?=C2=A1Buenos_d=C3=ADas!?=",
			want: "¡Buenos días!",
		},
		{
			name: "base64",
			in:   "=?UTF-8?B?bW9rYXBp?=",
			want: "mokapi",
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			act, err := smtp.DecodeHeaderValue(tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.want, act)
		})
	}
}

func TestMessage_UpdateSize(t *testing.T) {
	msg := &smtp.Message{
		MessageId: "id@mokapi.io",
		Date:      time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		Subject:   "hi",
		From:      []smtp.Address{{Address: "a@foo.bar"}},
		To:        []smtp.Address{{Name: "Bob", Address: "b@foo.bar"}},
		Body:      "hello",
	}

	msg.UpdateSize()
	withoutAttachment := msg.Size
	require.Greater(t, withoutAttachment, len("hello"))

	msg.Attachments = []smtp.Attachment{{Data: []byte("12345678")}}
	msg.UpdateSize()
	require.Equal(t, withoutAttachment+8, msg.Size)
}
