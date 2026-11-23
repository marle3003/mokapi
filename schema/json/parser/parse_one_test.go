package parser_test

import (
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"mokapi/sortedmap"
	"testing"

	"github.com/stretchr/testify/require"
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
				require.EqualError(t, err, "Validation error count 1:\n\t- #/oneOf: valid against more than one schema: valid schema indexes: 0, 1")
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
				require.EqualError(t, err, "Validation error count 1:\n\t- #/oneOf: valid against more than one schema: valid schema indexes: 0, 1")
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
				require.EqualError(t, err, "Validation error count 1:\n\t- #/oneOf: valid against more than one schema: valid schema indexes: 0, 1")
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
				require.EqualError(t, err, "Validation error count 1:\n\t- #/oneOf: valid against no schemas")
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
				require.EqualError(t, err, "Validation error count 1:\n\t- #/oneOf: valid against more than one schema: valid schema indexes: 0, 1")
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
				require.EqualError(t, err, "Validation error count 1:\n\t- #/oneOf/0/unevaluatedProperties: property foo not successfully evaluated and schema does not allow unevaluated properties")
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

func TestParser_ParseOne_PartialObjects(t *testing.T) {
	dogOrCat := schematest.NewOneOf(
		schematest.New("object", schematest.WithProperty("bark", schematest.New("string")), schematest.WithRequired("bark"), schematest.WithFreeForm(false)),
		schematest.New("object", schematest.WithProperty("meow", schematest.New("string")), schematest.WithRequired("meow"), schematest.WithFreeForm(false)),
	)
	multipleOf := schematest.NewOneOf(
		schematest.New("integer", schematest.WithMultipleOf(5)),
		schematest.New("integer", schematest.WithMultipleOf(3)),
	)

	testcases := []struct {
		name    string
		partial bool
		s       *schema.Schema
		data    any
		err     string
		result  any
	}{
		{
			name:    "several object matches, first wins",
			partial: true,
			s:       dogOrCat,
			data:    map[string]any{},
			result:  map[string]any{},
		},
		{
			name: "strict mode: object matches none because required is missing",
			s:    dogOrCat,
			data: map[string]any{},
			err:  "required properties are missing: meow",
		},
		{
			name:    "no object match is still an error",
			partial: true,
			s:       dogOrCat,
			data:    map[string]any{"bark": 1},
			err:     "#/oneOf/1/additionalProperties",
		},
		{
			name:    "null subschema and object subschema, first wins",
			partial: true,
			s:       schematest.NewOneOf(nil, schematest.New("object")),
			data:    map[string]any{"a": 1},
			result:  map[string]any{"a": 1},
		},
		{
			name:    "object subschema before null subschema",
			partial: true,
			s:       schematest.NewOneOf(schematest.New("object"), nil),
			data:    map[string]any{"a": 1},
			result:  map[string]any{"a": 1},
		},
		{
			name:    "several primitive matches, first wins",
			partial: true,
			s:       multipleOf,
			data:    15,
			result:  int64(15),
		},
		{
			name: "strict mode: several primitive matches",
			s:    multipleOf,
			data: 15,
			err:  "valid against more than one schema: valid schema indexes: 0, 1",
		},
		{
			name:    "no primitive match is still an error",
			partial: true,
			s:       multipleOf,
			data:    7,
			err:     "valid against no schemas",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.Parser{Schema: tc.s, ValidateAdditionalProperties: true, PartialObjects: tc.partial}
			v, err := p.Parse(tc.data)
			if tc.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.result, toMap(v))
		})
	}
}

func toMap(v any) any {
	if m, ok := v.(*sortedmap.LinkedHashMap[string, any]); ok {
		return m.ToMap()
	}
	return v
}
