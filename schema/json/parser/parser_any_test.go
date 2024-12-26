package parser_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
	"testing"
)

func TestParser_ParseAny(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name: "any",
			s:    "12",
			schema: schematest.NewAny(
				schematest.New("string"),
				schematest.New("integer")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), i)
			},
		},
		{
			name: "not match any",
			s:    "12.6",
			schema: schematest.NewAny(
				schematest.New("string"),
				schematest.New("integer")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ndoes not match any schemas of 'anyOf'\nschema path #/anyOf")
			},
		},
		{
			name: "any object",
			s:    `{"foo": "bar"}`,
			schema: schematest.NewAny(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("string")))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, i)
			},
		},
		{
			name: "missing required property should not error",
			s:    `{"name": "bar"}`,
			schema: schematest.NewAny(
				schematest.New("object",
					schematest.WithProperty("name", schematest.New("string"))),
				schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("age", schematest.New("integer")),
					schematest.WithRequired("age"),
				)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "bar"}, i)
			},
		},
		{
			name: "merge",
			s:    `{"name": "bar", "age": 12}`,
			schema: schematest.NewAny(
				schematest.New("object",
					schematest.WithProperty("name", schematest.New("string"))),
				schematest.New("object",
					schematest.WithProperty("age", schematest.New("integer")),
					schematest.WithRequired("age"),
				)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "bar", "age": int64(12)}, i)
			},
		},
		{
			name: "anyOf: object containing both properties",
			s:    `{"test": 12, "test2": true}`,
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("test", schematest.New("integer"))),
				schematest.New("object", schematest.WithProperty("test2", schematest.New("boolean"))),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"test": int64(12), "test2": true}, i)
			},
		},
		{
			name: "anyOf",
			s:    `"hello world"`,
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("test", schematest.New("integer"))),
				schematest.New("string"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "hello world", i)
			},
		},
		{
			name: "free-form but not override",
			s:    `{"foo": 12, "bar": 12}`,
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithFreeForm(true)),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("integer"))),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": float64(12), "bar": int64(12)}, i)
			},
		},
		{
			name: "free-form overriding value",
			s:    `{"foo": 12}`,
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithFreeForm(true)),
				schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": int64(12)}, i)
			},
		},
		{
			name: "free-form and second object defines no property",
			s:    `{"foo": 12}`,
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithFreeForm(true)),
				schematest.New("object"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": float64(12)}, i)
			},
		},
		{
			name: "first is not free-form",
			s:    `{"foo": 12, "bar": 12}`,
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("integer")), schematest.WithFreeForm(false)),
				schematest.New("object"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": int64(12), "bar": float64(12)}, i)
			},
		},
		{
			name: "one with error",
			s:    `{"foo": 12}`,
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
				schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": int64(12)}, i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var v interface{}
			err := json.Unmarshal([]byte(tc.s), &v)
			require.NoError(t, err)

			p := &parser.Parser{}
			r, err := p.Parse(v, &schema.Ref{Value: tc.schema})

			tc.test(t, r, err)
		})
	}
}
