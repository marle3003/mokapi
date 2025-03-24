package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
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
			schema: schematest.New("integer"),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(42), i)
			},
		},
		{
			name:   "int64",
			s:      "42",
			schema: schematest.New("integer", schematest.WithFormat("int64")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(42), i)
			},
		},
		{
			name:   "int32",
			s:      "42",
			schema: schematest.New("integer", schematest.WithFormat("int32")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(42), i)
			},
		},
		{
			name:   "float",
			s:      "42.8",
			schema: schematest.New("number", schematest.WithFormat("float")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 42.8, i)
			},
		},
		{
			name:   "string format date",
			s:      "2021-09-21",
			schema: schematest.New("string", schematest.WithFormat("date")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2021-09-21", i)
			},
		},
		{
			name:   "string format date-time",
			s:      "2021-09-21T13:22:11.408Z",
			schema: schematest.New("string", schematest.WithFormat("date-time")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2021-09-21T13:22:11.408Z", i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i, err := schema.ParseString(tc.s, tc.schema)
			tc.test(t, i, err)
		})
	}
}
