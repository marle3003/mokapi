package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
	"testing"
)

func TestParser_ParseOne(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Schema
		data interface{}
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "oneOf integer valid first",
			s: schematest.NewOneOf(
				schematest.New("integer", schematest.WithMultipleOf(5)),
				schematest.New("integer", schematest.WithMultipleOf(3)),
			),
			data: 10,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), v)
			},
		},
		{
			name: "oneOf integer valid second",
			s: schematest.NewOneOf(
				schematest.New("integer", schematest.WithMultipleOf(5)),
				schematest.New("integer", schematest.WithMultipleOf(3)),
			),
			data: 9,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(9), v)
			},
		},
		{
			name: "oneOf integer valid both",
			s: schematest.NewOneOf(
				schematest.New("integer", schematest.WithMultipleOf(5)),
				schematest.New("integer", schematest.WithMultipleOf(3)),
			),
			data: 15,
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "parse 15 failed: valid against more than one schema from 'oneOf': one of schema type=integer multipleOf=5, schema type=integer multipleOf=3")
			},
		},
		{
			name: "oneOf is null",
			s: schematest.NewOneOf(
				schematest.New("integer"),
				nil,
			),
			data: 15,
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "parse 15 failed: valid against more than one schema from 'oneOf': one of schema type=integer, empty schema")
			},
		},
		{
			name: "oneOf is empty",
			s: schematest.NewOneOf(
				schematest.New("integer"),
				&schema.Schema{},
			),
			data: 15,
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "parse 15 failed: valid against more than one schema from 'oneOf': one of schema type=integer, empty schema")
			},
		},
		{
			name: "not valid against one schema",
			s: schematest.NewOneOf(
				schematest.New("integer"),
				schematest.New("boolean"),
			),
			data: "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "parse foo failed: valid against no schemas from 'oneOf': one of schema type=integer, schema type=boolean")
			},
		},
		{
			name: "empty object match both",
			s: schematest.NewOneOf(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("string"))),
			),
			data: map[string]interface{}{},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "parse {} failed: valid against more than one schema from 'oneOf': one of schema type=object properties=[foo], schema type=object properties=[bar]")
			},
		},
		{
			name: "example from swagger.io but cat does not allow additional properties",
			s: schematest.NewOneOf(
				schematest.New("object",
					schematest.WithProperty("bark", schematest.New("boolean")),
					schematest.WithProperty("breed", schematest.New("string",
						schematest.WithEnum([]interface{}{"Dingo", "Husky", "Retriever", "Shepherd"})),
					),
				),
				schematest.New("object",
					schematest.WithProperty("hunts", schematest.New("boolean")),
					schematest.WithProperty("age", schematest.New("integer")),
					schematest.WithFreeForm(false),
				),
			),
			data: map[string]interface{}{"bark": true, "breed": "Dingo"},
			test: func(t *testing.T, result interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"bark": true, "breed": "Dingo"}, result)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parser.Parser{ValidateAdditionalProperties: true}
			v, err := p.ParseOne(tc.s, tc.data)
			tc.test(t, v, err)
		})
	}
}
