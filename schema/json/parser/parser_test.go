package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func toBoolP(b bool) *bool { return &b }

func TestParser_NoType(t *testing.T) {
	// JSON schema does not require a type
	// Some validation keywords only apply to one or more primitive types. When the primitive
	// type of the instance cannot be validated by a given keyword, validation for this keyword
	// and instance SHOULD succeed.

	testcases := []struct {
		name   string
		data   interface{}
		schema *schema.Schema
		test   func(t *testing.T, v interface{}, err error)
	}{
		{
			name:   "no type",
			data:   map[string]interface{}{"foo": "bar"},
			schema: schematest.NewTypes(nil),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, v)
			},
		},
		{
			name: "no type with property and maxLength; data is map",
			data: map[string]interface{}{"foo": "bar"},
			schema: schematest.NewTypes(nil,
				schematest.WithProperty("foo", schematest.New("string")),
				schematest.WithMaxLength(10),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, v)
			},
		},
		{
			name: "no type with property and maxLength; data is string",
			data: "foobar",
			schema: schematest.NewTypes(nil,
				schematest.WithProperty("foo", schematest.New("string")),
				schematest.WithMaxLength(10),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foobar", v)
			},
		},
		{
			name: "no type with property and maxLength; data is string but too long",
			data: "foobar1234567",
			schema: schematest.NewTypes(nil,
				schematest.WithProperty("foo", schematest.New("string")),
				schematest.WithMaxLength(10),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/maxLength: string 'foobar1234567' exceeds maximum of 10")
			},
		},
		{
			name:   "null but not nullable",
			data:   nil,
			schema: schematest.New("string"),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/type: invalid type, expected string but got null")
			},
		},
		{
			name:   "null but with default",
			data:   nil,
			schema: schematest.New("string", schematest.WithDefault("foobar")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foobar", v)
			},
		},
		{
			name:   "const error",
			data:   "foo",
			schema: schematest.New("string", schematest.WithConst("bar")),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/const: value 'foo' does not match const 'bar'")
			},
		},
		{
			name:   "const does not match schema",
			data:   "foo",
			schema: schematest.New("string", schematest.WithConst(3)),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 2:\n\t- #/const: const value does not match schema\n\t\t- #/const/type: invalid type, expected string but got integer")
			},
		},
		{
			name:   "not string error",
			data:   "foo",
			schema: schematest.NewTypes(nil, schematest.WithNot(schematest.New("string"))),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/not: is valid against schema")
			},
		},
		{
			name:   "not string",
			data:   12,
			schema: schematest.NewTypes(nil, schematest.WithNot(schematest.New("string"))),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), v)
			},
		},
		{
			name: "array contains object with error",
			data: []interface{}{map[string]interface{}{"foo": 123}},
			schema: schematest.New("array",
				schematest.WithItems("object",
					schematest.WithProperty("foo", schematest.New("string")),
				),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n\t- #/items/0/foo/type: invalid type, expected string but got integer")
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
