package parser

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
	"testing"
)

func TestParser_ParseAll(t *testing.T) {
	testcases := []struct {
		name   string
		data   interface{}
		schema *schema.Schema
		test   func(t *testing.T, v interface{}, err error)
	}{
		{
			name:   "AllOf empty",
			data:   12,
			schema: schematest.NewAllOf(),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), v)
			},
		},
		{
			name:   "AllOf with one matching type",
			data:   12,
			schema: schematest.NewAllOf(schematest.New("integer")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), v)
			},
		},
		{
			name: "AllOf with two matching type",
			data: 12,
			schema: schematest.NewAllOf(
				schematest.New("integer"),
				schematest.New("integer", schematest.WithMaximum(12)),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), v)
			},
		},
		{
			name: "AllOf with two types one is empty",
			data: 12,
			schema: schematest.NewAllOf(
				schematest.New("integer"),
				schematest.NewTypes(nil),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), v)
			},
		},
		{
			name: "AllOf with two matching type but not valid",
			data: 12,
			schema: schematest.NewAllOf(
				schematest.New("integer"),
				schematest.New("integer", schematest.WithMaximum(11)),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "12 is greater as the required maximum 11, expected schema type=integer maximum=11")
			},
		},
		{
			name: "AllOf with two NOT matching type",
			data: 12,
			schema: schematest.NewAllOf(
				schematest.New("integer"),
				schematest.New("string"),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "allOf contains different types: all of schema type=integer, schema type=string")
			},
		},
		{
			name: "AllOf with two objects",
			data: map[string]interface{}{
				"name": "carol",
				"age":  28,
			},
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("name", schematest.New("string"))),
				schematest.New("object", schematest.WithProperty("age", schematest.New("integer"))),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"age": int64(28), "name": "carol"}, v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &Parser{}
			v, err := p.Parse(tc.data, &schema.Ref{Value: tc.schema})
			tc.test(t, v, err)
		})
	}
}
