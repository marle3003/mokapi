package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
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
				require.EqualError(t, err, "found 1 error:\nstring 'foobar1234567' exceeds maximum of 10\nschema path #/maxLength")
			},
		},
		{
			name:   "null but not nullable",
			data:   nil,
			schema: schematest.New("string"),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\nparse NULL failed, expected schema type=string")
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
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &parser.Parser{}
			v, err := p.Parse(tc.data, &schema.Ref{Value: tc.schema})
			tc.test(t, v, err)
		})
	}
}
