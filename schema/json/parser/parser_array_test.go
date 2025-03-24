package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestParser_Array(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		data   interface{}
		test   func(t *testing.T, v interface{}, err error)
	}{
		{
			name:   "array error",
			schema: schematest.New("array"),
			data:   map[string]interface{}{"a": 1, "b": 2, "c": nil},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/type: invalid type, expected array but got object")
			},
		},
		{
			name:   "array",
			schema: schematest.New("array"),
			data:   []interface{}{1, 2, 3, 4, 5},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{1, 2, 3, 4, 5}, v)
			},
		},
		{
			name:   "array items error",
			schema: schematest.New("array", schematest.WithItems("integer")),
			data:   []interface{}{1, 2, 3, 4, "foo"},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/items/4/type: invalid type, expected integer but got string")
			},
		},
		{
			name:   "array items",
			schema: schematest.New("array", schematest.WithItems("integer")),
			data:   []interface{}{1, 2, 3, 4},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(1), int64(2), int64(3), int64(4)}, v)
			},
		},
		{
			name:   "empty array",
			schema: schematest.New("array", schematest.WithItems("integer")),
			data:   []interface{}{},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{}, v)
			},
		},
		{
			name: "prefixItems error",
			schema: schematest.New("array", schematest.WithPrefixItems(
				schematest.New("integer"),
				schematest.New("integer"),
			)),
			data: []interface{}{1, "foo"},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/items/1/type: invalid type, expected integer but got string")
			},
		},
		{
			name: "prefixItems with items error",
			schema: schematest.New("array", schematest.WithPrefixItems(
				schematest.New("integer"),
				schematest.New("integer"),
			), schematest.WithItems("integer")),
			data: []interface{}{1, 2, "foo"},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/items/2/type: invalid type, expected integer but got string")
			},
		},
		{
			name: "prefixItems",
			schema: schematest.New("array", schematest.WithPrefixItems(
				schematest.New("integer"),
				schematest.New("integer"),
			)),
			data: []interface{}{1, 2},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(1), int64(2)}, v)
			},
		},
		{
			name: "prefixItems but longer",
			schema: schematest.New("array", schematest.WithPrefixItems(
				schematest.New("integer"),
				schematest.New("integer"),
			)),
			data: []interface{}{1, 2, 4.5},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(1), int64(2), 4.5}, v)
			},
		},
		{
			name: "prefixItems with items=false",
			schema: schematest.New("array", schematest.WithPrefixItems(
				schematest.New("integer"),
				schematest.New("integer"),
			), schematest.WithItemsNew(schematest.NewBool(false))),
			data: []interface{}{1, 2, "foo"},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/items/2/valid: schema always fails validation")
			},
		},
		{
			name: "prefixItems with unevaluatedItems=false",
			schema: schematest.New("array", schematest.WithPrefixItems(
				schematest.New("integer"),
				schematest.New("integer"),
			), schematest.WithUnevaluatedItems(&schema.Schema{Boolean: toBoolP(false)})),
			data: []interface{}{1, 2, "foo"},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/unevaluatedItems: item at index 2 has not been successfully evaluated and the schema does not allow unevaluated items")
			},
		},
		{
			name: "prefixItems with unevaluatedItems=string",
			schema: schematest.New("array", schematest.WithPrefixItems(
				schematest.New("integer"),
				schematest.New("integer"),
			), schematest.WithUnevaluatedItems(schematest.New("string"))),
			data: []interface{}{1, 2, "foo"},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(1), int64(2), "foo"}, v)
			},
		},
		{
			name:   "contains error",
			schema: schematest.New("array", schematest.WithContains(schematest.New("integer"))),
			data:   []interface{}{"foo"},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/contains: no items match contains")
			},
		},
		{
			name:   "contains",
			schema: schematest.New("array", schematest.WithContains(schematest.New("integer"))),
			data:   []interface{}{"foo", 1},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", 1}, v)
			},
		},
		{
			name: "minContains error",
			schema: schematest.New("array",
				schematest.WithContains(schematest.New("integer")),
				schematest.WithMinContains(2),
			),
			data: []interface{}{1},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/minContains: contains match count 1 is less than minimum contains count of 2")
			},
		},
		{
			name: "minContains",
			schema: schematest.New("array",
				schematest.WithContains(schematest.New("integer")),
				schematest.WithMinContains(2),
			),
			data: []interface{}{"foo", 1, 2},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", 1, 2}, v)
			},
		},
		{
			name: "maxContains error",
			schema: schematest.New("array",
				schematest.WithContains(schematest.New("integer")),
				schematest.WithMaxContains(1),
			),
			data: []interface{}{1, 2},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/maxContains: contains match count 2 exceeds maximum contains count of 1")
			},
		},
		{
			name: "maxContains",
			schema: schematest.New("array",
				schematest.WithContains(schematest.New("integer")),
				schematest.WithMaxContains(2),
			),
			data: []interface{}{"foo", 1, 2},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", 1, 2}, v)
			},
		},
		{
			name: "minItems error",
			schema: schematest.New("array",
				schematest.WithMinItems(2),
			),
			data: []interface{}{1},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/minItems: item count 1 is less than minimum count of 2")
			},
		},
		{
			name: "minItems",
			schema: schematest.New("array",
				schematest.WithMinItems(2),
			),
			data: []interface{}{"foo", 1, 2},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", 1, 2}, v)
			},
		},
		{
			name: "maxItems error",
			schema: schematest.New("array",
				schematest.WithMaxItems(1),
			),
			data: []interface{}{1, 2},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/maxItems: item count 2 exceeds maximum count of 1")
			},
		},
		{
			name: "maxContains",
			schema: schematest.New("array",
				schematest.WithMaxContains(2),
			),
			data: []interface{}{"foo", 1, 2},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", 1, 2}, v)
			},
		},
		{
			name: "uniqueItems error",
			schema: schematest.New("array",
				schematest.WithUniqueItems(),
			),
			data: []interface{}{1, 2, 2},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/uniqueItems: non-unique array item at index 2")
			},
		},
		{
			name: "uniqueItems",
			schema: schematest.New("array",
				schematest.WithUniqueItems(),
			),
			data: []interface{}{"foo", 1, 2},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", 1, 2}, v)
			},
		},
		{
			name: "const error",
			schema: schematest.New("array",
				schematest.WithConst([]string{"a", "b", "c"}),
			),
			data: []string{"a", "b"},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/const: value '[a, b]' does not match const '[a, b, c]'")
			},
		},
		{
			name: "uniqueItems",
			schema: schematest.New("array",
				schematest.WithConst([]string{"a", "b", "c"}),
			),
			data: []string{"a", "b", "c"},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"a", "b", "c"}, v)
			},
		},
		{
			name: "numErrors",
			schema: schematest.New("array",
				schematest.WithItems("string"),
			),
			data: []interface{}{"a", 1, 2},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 2:\n\t- #/items/1/type: invalid type, expected string but got integer\n\t- #/items/2/type: invalid type, expected string but got integer")
			},
		},
		{
			name: "numErrors nested",
			schema: schematest.New("array",
				schematest.WithItems("object",
					schematest.WithProperty("foo", schematest.New("string", schematest.WithMinLength(3))),
					schematest.WithProperty("bar", schematest.New("integer", schematest.WithMinimum(5))),
				),
			),
			data: []interface{}{
				map[string]interface{}{"foo": "foo"},
				map[string]interface{}{"foo": "foo", "bar": "bar"},
				map[string]interface{}{"foo": "a", "bar": "bar"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 3:\n\t- #/items/1/bar/type: invalid type, expected integer but got string\n\t- #/items/2/foo/minLength: string 'a' is less than minimum of 3\n\t- #/items/2/bar/type: invalid type, expected integer but got string")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &parser.Parser{Schema: tc.schema}
			v, err := p.Parse(tc.data)
			tc.test(t, v, err)
		})
	}
}
