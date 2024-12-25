package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
	"testing"
)

func TestParser_ParseObject(t *testing.T) {
	testcases := []struct {
		name   string
		data   interface{}
		schema *schema.Schema
		test   func(t *testing.T, v interface{}, err error)
	}{
		{
			name:   "expect object but got integer",
			data:   12,
			schema: schematest.New("object"),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected object but got integer\nschema path #/type")
			},
		},
		{
			name:   "property invalid type",
			data:   map[string]interface{}{"foo": 1234},
			schema: schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected string but got integer\nschema path #/foo/type")
			},
		},
		{
			name: "two properties invalid type",
			data: map[string]interface{}{"foo": 1234, "bar": 1234},
			schema: schematest.New("object",
				schematest.WithProperty("foo", schematest.New("string")),
				schematest.WithProperty("bar", schematest.New("string")),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 2 errors:\ninvalid type, expected string but got integer\nschema path #/foo/type\ninvalid type, expected string but got integer\nschema path #/bar/type")
			},
		},
		{
			name: "pattern properties error",
			data: map[string]interface{}{"S_25": 1234},
			schema: schematest.New("object",
				schematest.WithPatternProperty("^S_", schematest.New("string")),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected string but got integer\nschema path #/patternProperties/^S_/type")
			},
		},
		{
			name: "pattern properties",
			data: map[string]interface{}{"S_25": "foo"},
			schema: schematest.New("object",
				schematest.WithPatternProperty("^S_", schematest.New("string")),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"S_25": "foo"}, v)
			},
		},
		{
			name: "additional properties false error",
			data: map[string]interface{}{"foo": "bar"},
			schema: schematest.New("object",
				schematest.WithPatternProperty("^S_", schematest.New("string")),
				schematest.WithFreeForm(false),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\nproperty 'foo' not defined and the schema does not allow additional properties\nschema path #/additionalProperties")
			},
		},
		{
			name: "additional properties true",
			data: map[string]interface{}{"foo": "bar"},
			schema: schematest.New("object",
				schematest.WithFreeForm(true),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, v)
			},
		},
		{
			name: "additional properties string error",
			data: map[string]interface{}{"number2": 1234},
			schema: schematest.New("object",
				schematest.WithProperty("number", schematest.New("number")),
				schematest.WithAdditionalProperties(schematest.New("string")),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected string but got integer\nschema path #/additionalProperties/number2/type")
			},
		},
		{
			name: "additional properties string",
			data: map[string]interface{}{"foo": "bar"},
			schema: schematest.New("object",
				schematest.WithProperty("number", schematest.New("number")),
				schematest.WithAdditionalProperties(schematest.New("string")),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, v)
			},
		},
		{
			name: "additional properties combination with properties and patternProperties",
			data: map[string]interface{}{"builtin": 42},
			schema: schematest.New("object",
				schematest.WithProperty("builtin", schematest.New("number")),
				schematest.WithPatternProperty("^S_", schematest.New("string")),
				schematest.WithPatternProperty("^I_", schematest.New("integer")),
				schematest.WithAdditionalProperties(schematest.New("string")),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"builtin": float64(42)}, v)
			},
		},
		{
			name: "additional properties combination with properties and patternProperties, not match regex but string",
			data: map[string]interface{}{"keyword": "value"},
			schema: schematest.New("object",
				schematest.WithProperty("builtin", schematest.New("number")),
				schematest.WithPatternProperty("^S_", schematest.New("string")),
				schematest.WithPatternProperty("^I_", schematest.New("integer")),
				schematest.WithAdditionalProperties(schematest.New("string")),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"keyword": "value"}, v)
			},
		},
		{
			name: "additional properties combination with properties and patternProperties, must be a string",
			data: map[string]interface{}{"keyword": 42},
			schema: schematest.New("object",
				schematest.WithProperty("builtin", schematest.New("number")),
				schematest.WithPatternProperty("^S_", schematest.New("string")),
				schematest.WithPatternProperty("^I_", schematest.New("integer")),
				schematest.WithAdditionalProperties(schematest.New("string")),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected string but got integer\nschema path #/additionalProperties/keyword/type")
			},
		},
		{
			name: "required string error",
			data: map[string]interface{}{},
			schema: schematest.New("object",
				schematest.WithProperty("foo", schematest.New("string")),
				schematest.WithRequired("foo")),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\nrequired properties are missing: foo\nschema path #/required")
			},
		},
		{
			name: "required string but it is empty",
			data: map[string]interface{}{"foo": ""},
			schema: schematest.New("object",
				schematest.WithProperty("foo", schematest.New("string")),
				schematest.WithRequired("foo")),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": ""}, v)
			},
		},
		{
			name: "propertyNames error",
			data: map[string]interface{}{"1_foo": ""},
			schema: schematest.New("object",
				schematest.WithPropertyNames(schematest.NewTypes(nil, schematest.WithPattern("^[a-z]+"))),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\nstring '1_foo' does not match regex pattern '^[a-z]+'\nschema path #/propertyNames/pattern")
			},
		},
		{
			name: "propertyNames",
			data: map[string]interface{}{"foo": ""},
			schema: schematest.New("object",
				schematest.WithPropertyNames(schematest.NewTypes(nil, schematest.WithPattern("^[a-z]+"))),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": ""}, v)
			},
		},
		{
			name: "minProperties error",
			data: map[string]interface{}{"foo": ""},
			schema: schematest.New("object",
				schematest.WithMinProperties(2),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\nproperty count 1 is less than minimum count of 2\nschema path #/minProperties")
			},
		},
		{
			name: "minProperties",
			data: map[string]interface{}{"foo": ""},
			schema: schematest.New("object",
				schematest.WithMinProperties(1),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": ""}, v)
			},
		},
		{
			name: "maxProperties error",
			data: map[string]interface{}{"foo": "", "bar": ""},
			schema: schematest.New("object",
				schematest.WithMaxProperties(1),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\nproperty count 2 exceeds maximum count of 1\nschema path #/maxProperties")
			},
		},
		{
			name: "maxProperties",
			data: map[string]interface{}{"foo": ""},
			schema: schematest.New("object",
				schematest.WithMaxProperties(1),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": ""}, v)
			},
		},
		{
			name:   "const error",
			schema: schematest.New("object", schematest.WithConst(map[string]interface{}{"foo": "bar"})),
			data:   map[string]interface{}{"foo": "foobar"},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\nvalue '{foo: foobar}' does not match const '{foo: bar}'\nschema path #/const")
			},
		},
		{
			name:   "const",
			schema: schematest.New("object", schematest.WithConst(map[string]interface{}{"foo": "bar"})),
			data:   map[string]interface{}{"foo": "bar"},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, v)
			},
		},
		{
			name:   "dependentRequired error",
			schema: schematest.New("object", schematest.WithDependentRequired("foo", "bar")),
			data:   map[string]interface{}{"foo": "foobar"},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\ndependencies for property 'foo' failed: missing required keys: bar.\nschema path #/dependentRequired")
			},
		},
		{
			name:   "dependentRequired",
			schema: schematest.New("object", schematest.WithDependentRequired("foo", "bar")),
			data:   map[string]interface{}{"foo": "foobar", "bar": 12},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "foobar", "bar": 12}, v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &parser.Parser{ValidateAdditionalProperties: true}
			v, err := p.Parse(tc.data, &schema.Ref{Value: tc.schema})
			tc.test(t, v, err)
		})
	}
}
