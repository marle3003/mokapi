package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
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
				require.EqualError(t, err, "error count 1:\n- #/oneOf: valid against more than one schema: valid schema indexes: 0, 1")
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
				require.EqualError(t, err, "error count 1:\n- #/oneOf: valid against more than one schema: valid schema indexes: 0, 1")
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
				require.EqualError(t, err, "error count 1:\n- #/oneOf: valid against more than one schema: valid schema indexes: 0, 1")
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
				require.EqualError(t, err, "error count 1:\n- #/oneOf: valid against no schemas")
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
				require.EqualError(t, err, "error count 1:\n- #/oneOf: valid against more than one schema: valid schema indexes: 0, 1")
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
		{
			name: "unevaluatedProperty",
			s: schematest.NewOneOf(
				schematest.New("object", schematest.WithUnevaluatedProperties(&schema.Schema{Boolean: toBoolP(false)})),
			),
			data: map[string]interface{}{"foo": "bar"},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/oneOf: valid against no schemas\n\t- #/oneOf/0/unevaluatedProperties: property foo not successfully evaluated and schema does not allow unevaluated properties")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parser.Parser{Schema: tc.s, ValidateAdditionalProperties: true}
			v, err := p.Parse(tc.data)
			tc.test(t, v, err)
		})
	}
}
