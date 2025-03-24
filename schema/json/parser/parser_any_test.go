package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestParser_ParseAny(t *testing.T) {
	testcases := []struct {
		name   string
		data   interface{}
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name: "any",
			data: 12,
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
			data: 12.6,
			schema: schematest.NewAny(
				schematest.New("string"),
				schematest.New("integer")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/anyOf: does not match any schemas")
			},
		},
		{
			name: "nil but not nullable",
			data: nil,
			schema: schematest.NewAny(
				schematest.New("string"),
				schematest.New("integer")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/anyOf: does not match any schemas")
			},
		},
		{
			name: "nil and nullable",
			data: nil,
			schema: schematest.NewAny(
				schematest.NewTypes([]string{"string", "null"}),
				schematest.New("integer")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, i)
			},
		},
		{
			name: "any object",
			data: map[string]interface{}{"foo": "bar"},
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
			data: map[string]interface{}{"name": "bar"},
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
			data: map[string]interface{}{"name": "bar", "age": 12},
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
			data: map[string]interface{}{"test": 12, "test2": true},
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
			data: "hello world",
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
			data: map[string]interface{}{"foo": 12, "bar": 12},
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithFreeForm(true)),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("integer"))),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": 12, "bar": int64(12)}, i)
			},
		},
		{
			name: "free-form overriding value",
			data: map[string]interface{}{"foo": 12},
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
			data: map[string]interface{}{"foo": 12},
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithFreeForm(true)),
				schematest.New("object"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": 12}, i)
			},
		},
		{
			name: "first is not free-form",
			data: map[string]interface{}{"foo": "12", "bar": 12},
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("integer")), schematest.WithFreeForm(false)),
				schematest.New("object"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "12", "bar": 12}, i)
			},
		},
		{
			name: "one with error",
			data: map[string]interface{}{"foo": 12},
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
				schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": int64(12)}, i)
			},
		},
		{
			name: "unevaluatedProperties error",
			data: map[string]interface{}{"foo": "bar", "yuh": "abc"},
			schema: schematest.NewAny(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("string")),
					schematest.WithUnevaluatedProperties(schematest.New("boolean")),
				),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/anyOf/0/unevaluatedProperties/type: does not match any schemas: invalid type, expected boolean but got string")
			},
		},
		{
			name: "unevaluatedProperties should only considered on valid ones",
			data: map[string]interface{}{"foo": 12, "bar": 123},
			schema: schematest.NewTypes(nil,
				schematest.Any(
					schematest.New("object",
						schematest.WithProperty("foo", schematest.New("string")),
					),
					schematest.New("object",
						schematest.WithProperty("foo", schematest.New("integer")),
						schematest.WithProperty("bar", schematest.New("integer")),
					),
				),
				schematest.WithUnevaluatedProperties(schematest.New("boolean")),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": int64(12), "bar": int64(123)}, i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &parser.Parser{Schema: tc.schema}
			r, err := p.Parse(tc.data)

			tc.test(t, r, err)
		})
	}
}
