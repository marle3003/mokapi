package openapi_test

import (
	"mokapi/providers/openapi"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncoding_IsExplode(t *testing.T) {
	testcases := []struct {
		name     string
		encoding openapi.Encoding
		expected bool
	}{
		{name: "explicit true", encoding: openapi.Encoding{Style: "pipeDelimited", Explode: new(true)}, expected: true},
		{name: "explicit false with style form", encoding: openapi.Encoding{Style: "form", Explode: new(false)}, expected: false},
		{name: "not set with style form", encoding: openapi.Encoding{Style: "form"}, expected: true},
		{name: "not set without style", encoding: openapi.Encoding{}, expected: true},
		{name: "not set with other style", encoding: openapi.Encoding{Style: "spaceDelimited"}, expected: false},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.encoding.IsExplode())
		})
	}
}

func TestEncoding_IsAllowReserved(t *testing.T) {
	require.False(t, (&openapi.Encoding{}).IsAllowReserved())
	require.False(t, (&openapi.Encoding{AllowReserved: new(false)}).IsAllowReserved())
	require.True(t, (&openapi.Encoding{AllowReserved: new(true)}).IsAllowReserved())
}
