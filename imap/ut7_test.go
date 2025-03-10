package imap

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDecodeUT7(t *testing.T) {
	testcases := []struct {
		input    string
		expected string
	}{
		{
			input:    "",
			expected: "",
		},
		{
			input:    "Entw&APw-rfe",
			expected: "Entwürfe",
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			v, err := DecodeUTF7(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, v)
		})
	}
}
