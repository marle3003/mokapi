package parser

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestToString(t *testing.T) {
	testcases := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "time.Time",
			value:    time.Date(2024, 12, 17, 20, 34, 58, 0, time.UTC),
			expected: `2024-12-17 20:34:58 +0000 UTC`,
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := ToString(tc.value)
			require.Equal(t, tc.expected, s)
		})
	}
}
