package smtp_test

import (
	"mokapi/smtp"
	"testing"

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
