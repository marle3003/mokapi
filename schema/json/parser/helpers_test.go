package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
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
			name:     "nil",
			value:    nil,
			expected: "null",
		},
		{
			name:     "string",
			value:    "hello",
			expected: "hello",
		},
		{
			name:     "integer",
			value:    12,
			expected: "12",
		},
		{
			name:     "number",
			value:    12.4,
			expected: "12.4",
		},
		{
			name:     "object as map",
			value:    map[string]interface{}{"foo": "bar"},
			expected: "{foo: bar}",
		},
		{
			name:     "object as struct",
			value:    struct{ Foo string }{"bar"},
			expected: "{foo: bar}",
		},
		{
			name:     "object as pointer",
			value:    &struct{ Foo string }{"bar"},
			expected: "{foo: bar}",
		},
		{
			name:     "string array",
			value:    []string{"foo", "bar"},
			expected: "[foo, bar]",
		},
		{
			name:     "interface array",
			value:    []interface{}{"foo", "bar"},
			expected: "[foo, bar]",
		},
		{
			name:     "boolean",
			value:    true,
			expected: "true",
		},
		{
			name:     "time.Time",
			value:    time.Date(2024, 12, 17, 20, 34, 58, 0, time.UTC),
			expected: "2024-12-17 20:34:58 +0000 UTC",
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := parser.ToString(tc.value)
			require.Equal(t, tc.expected, s)
		})
	}
}
