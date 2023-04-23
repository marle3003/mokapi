package kafka

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestErrorCode_String(t *testing.T) {
	testcases := []struct {
		name     string
		error    ErrorCode
		expected string
	}{
		{
			name:     "known error code",
			error:    UnknownServerError,
			expected: "UNKNOWN_SERVER_ERROR (-1)",
		},
		{
			name:     "unknown error code",
			error:    3200,
			expected: "unknown kafka error code: 3200",
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, tc.error.String())
		})
	}
}

func TestError_Error(t *testing.T) {
	testcases := []struct {
		name     string
		error    Error
		expected string
	}{
		{
			name:     "empty message",
			error:    Error{Header: &Header{}, Code: UnknownServerError},
			expected: "kafka: error code UNKNOWN_SERVER_ERROR (-1)",
		},
		{
			name:     "with message",
			error:    Error{Header: &Header{}, Code: UnsupportedVersion, Message: fmt.Sprintf("unsupported api version")},
			expected: "kafka: error code UNSUPPORTED_VERSION (35): unsupported api version",
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, tc.error.Error())
		})
	}
}
