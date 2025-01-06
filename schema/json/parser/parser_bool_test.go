package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
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
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected boolean but got string\nschema path #/type")
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
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected boolean but got integer\nschema path #/type")
			},
		},
		{
			// Values that evaluate to true or false are still not accepted by the schema:
			name:   "1",
			data:   1,
			schema: schematest.New("boolean"),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected boolean but got integer\nschema path #/type")
			},
		},
		{
			name:   "not bool",
			data:   []bool{true},
			schema: schematest.New("boolean"),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected boolean but got array\nschema path #/type")
			},
		},
		{
			name:   "const error",
			data:   false,
			schema: schematest.New("boolean", schematest.WithConst(true)),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\nvalue 'false' does not match const 'true'\nschema path #/const")
			},
		},
		{
			name:   "const",
			data:   true,
			schema: schematest.New("boolean", schematest.WithConst(true)),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, true, v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &parser.Parser{Schema: &schema.Ref{Value: tc.schema}, ConvertStringToBoolean: tc.convertStringToBoolean}
			v, err := p.Parse(tc.data)
			tc.test(t, v, err)
		})
	}
}
