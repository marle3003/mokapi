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
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected array but got object\nschema path #/type")
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
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected integer but got string\nschema path #/items/type")
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
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected integer but got string\nschema path #/prefixItems/1/type")
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
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected integer but got string\nschema path #/items/type")
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
				require.EqualError(t, err, "found 1 error:\nschema always fails validation\nschema path #/items/valid")
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
				require.EqualError(t, err, "found 1 error:\nitem at index 2 has not been successfully evaluated and the schema does not allow unevaluated items.\nschema path #/unevaluatedItems")
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
				require.EqualError(t, err, "found 1 error:\nno items match contains\nschema path #/contains")
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
				require.EqualError(t, err, "found 1 error:\ncontains match count 1 is less than minimum contains count of 2\nschema path #/minContains")
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
				require.EqualError(t, err, "found 1 error:\ncontains match count 2 exceeds maximum contains count of 1\nschema path #/maxContains")
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
				require.EqualError(t, err, "found 1 error:\nitem count 1 is less than minimum count of 2\nschema path #/minItems")
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
				require.EqualError(t, err, "found 1 error:\nitem count 2 exceeds maximum count of 1\nschema path #/maxItems")
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
				require.EqualError(t, err, "found 1 error:\nnon-unique array item at index 2\nschema path #/uniqueItems")
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
				require.EqualError(t, err, "found 1 error:\nvalue '[a, b]' does not match const '[a, b, c]'\nschema path #/const")
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
