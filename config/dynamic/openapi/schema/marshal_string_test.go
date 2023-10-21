package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi/schema"
	"testing"
)

func TestParseString(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name:   "int",
			s:      "42",
			schema: &schema.Schema{Type: "integer"},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(42), i)
			},
		},
		{
			name:   "int64",
			s:      "42",
			schema: &schema.Schema{Type: "integer", Format: "int64"},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(42), i)
			},
		},
		{
			name:   "int32",
			s:      "42",
			schema: &schema.Schema{Type: "integer", Format: "int32"},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(42), i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i, err := schema.ParseString("42", &schema.Ref{Value: tc.schema})
			tc.test(t, i, err)
		})
	}
}
