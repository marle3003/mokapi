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
				require.EqualError(t, err, "parse 12 failed: does not match all of schema type=integer, schema type=integer maximum=11: 12 is greater as the required maximum 11, expected schema type=integer maximum=11")
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
				require.EqualError(t, err, "parse 12 failed: does not match all of schema type=integer, schema type=string: parse 12 failed, expected schema type=string")
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
		{
			name: "AllOf with two objects, one defines a property without type",
			data: map[string]interface{}{
				"name": "carol",
				"age":  28,
			},
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("name", schematest.New("string")), schematest.WithProperty("age", schematest.New(""))),
				schematest.New("object", schematest.WithProperty("age", schematest.New("integer"))),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"age": int64(28), "name": "carol"}, v)
			},
		},
		{
			name: "AllOf with two objects free-form false and errors - error message should contain free-form=false",
			data: map[string]interface{}{
				"name": 12,
				"age":  "28",
			},
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("name", schematest.New("string")), schematest.WithFreeForm(false)),
				schematest.New("object", schematest.WithProperty("age", schematest.New("integer")), schematest.WithFreeForm(false)),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "parse {age: 28, name: 12} failed: does not match all of schema type=object properties=[name] free-form=false, schema type=object properties=[age] free-form=false:\nparse property 'name' failed: parse 12 failed, expected schema type=string\nparse property 'age' failed: parse '28' failed, expected schema type=integer")
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
