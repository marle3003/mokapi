package parser

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
	"testing"
)

func TestParser_ParseBoolean(t *testing.T) {
	testcases := []struct {
		name                   string
		data                   interface{}
		schema                 *schema.Schema
		convertStringToBoolean bool
		test                   func(t *testing.T, v interface{}, err error)
	}{
		{
			name:   "true",
			data:   true,
			schema: schematest.New("boolean"),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.True(t, v.(bool))
			},
		},
		{
			name:   "false",
			data:   false,
			schema: schematest.New("boolean"),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.False(t, v.(bool))
			},
		},
		{
			name:   "true as string",
			data:   "true",
			schema: schematest.New("boolean"),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "parse 'true' (string) failed, expected schema type=boolean")
			},
		},
		{
			name:                   "true as string with convert",
			data:                   "true",
			schema:                 schematest.New("boolean"),
			convertStringToBoolean: true,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.True(t, v.(bool))
			},
		},
		{
			name:                   "FAlse as string with convert",
			data:                   "FAlse",
			schema:                 schematest.New("boolean"),
			convertStringToBoolean: true,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.False(t, v.(bool))
			},
		},
		{
			// Values that evaluate to true or false are still not accepted by the schema:
			name:   "0",
			data:   0,
			schema: schematest.New("boolean"),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "parse '0' (int) failed, expected schema type=boolean: invalid type")
			},
		},
		{
			// Values that evaluate to true or false are still not accepted by the schema:
			name:   "1",
			data:   1,
			schema: schematest.New("boolean"),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "parse '1' (int) failed, expected schema type=boolean: invalid type")
			},
		},
		{
			name:   "not bool",
			data:   []bool{true},
			schema: schematest.New("boolean"),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "parse [true] failed, expected schema type=boolean")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &Parser{ConvertStringToBoolean: tc.convertStringToBoolean}
			v, err := p.ParseBoolean(tc.data, tc.schema)
			tc.test(t, v, err)
		})
	}
}
