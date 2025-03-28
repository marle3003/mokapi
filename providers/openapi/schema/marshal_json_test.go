package schema_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/media"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/sortedmap"
	"testing"
)

func TestRef_Marshal_Json(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		data   interface{}
		test   func(t *testing.T, result string, err error)
	}{
		{
			name:   "no schema",
			schema: &schema.Schema{},
			data:   "foo",
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `"foo"`, result)
			},
		},
		{
			name:   "NULL no schema",
			schema: &schema.Schema{},
			data:   nil,
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, "null", result)
			},
		},
		{
			name:   "NULL not allowed",
			schema: schematest.New("number"),
			data:   nil,
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "encoding data to 'application/json' failed: error count 1:\n\t- #/type: invalid type, expected number but got null")
			},
		},
		{
			name:   "NULL and nullable=true",
			schema: schematest.New("number", schematest.IsNullable(true)),
			data:   nil,
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, "null", result)
			},
		},
		{
			name:   "NULL and type=null",
			schema: schematest.New("number", schematest.And("null")),
			data:   nil,
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, "null", result)
			},
		},
		{
			name:   "number",
			schema: schematest.New("number"),
			data:   3.141,
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, "3.141", result)
			},
		},
		{
			name:   "string",
			schema: schematest.New("string"),
			data:   "12",
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `"12"`, result)
			},
		},
		{
			name:   "nullable string",
			schema: schematest.New("string", schematest.IsNullable(true)),
			data:   nil,
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, "null", result)
			},
		},
		{
			name:   "integer as string",
			schema: schematest.New("integer"),
			data:   "12",
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "encoding data to 'application/json' failed: error count 1:\n\t- #/type: invalid type, expected integer but got string")
			},
		},
		{
			name:   "array of integer",
			schema: schematest.New("array", schematest.WithItems("integer")),
			data:   []interface{}{12, 13},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, "[12,13]", result)
			},
		},
		{
			name:   "array of string",
			schema: schematest.New("array", schematest.WithItems("string")),
			data:   []interface{}{"foo", "bar"},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `["foo","bar"]`, result)
			},
		},
		{
			name:   "accept strings or numbers, value is string",
			schema: schematest.New("string", schematest.And("number")),
			data:   "foo",
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `"foo"`, result)
			},
		},
		{
			name:   "accept strings or numbers, value is number",
			schema: schematest.New("string", schematest.And("number")),
			data:   12,
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `12`, result)
			},
		},
		{
			name:   "accept strings or numbers, value is boolean",
			schema: schematest.New("string", schematest.And("number")),
			data:   true,
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "encoding data to 'application/json' failed: error count 1:\n\t- #/type: invalid type, expected number but got boolean")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := tc.schema.Marshal(tc.data, media.ParseContentType("application/json"))
			tc.test(t, string(b), err)
		})
	}
}

func TestRef_Marshal_Json_Object(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		data   interface{}
		test   func(t *testing.T, result string, err error)
	}{
		{
			name: "data is not struct or map",
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer"))),
			data: 12,
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "encoding data to 'application/json' failed: error count 1:\n\t- #/type: invalid type, expected object but got integer")
				require.Len(t, result, 0)
			},
		},
		{
			name: "map with key string",
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer"))),
			data: map[string]interface{}{"name": "foo", "value": 12},
			test: func(t *testing.T, result string, err error) {
				require.Equal(t, `{"name":"foo","value":12}`, result)
			},
		},
		{
			name: "map with key interface{}",
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer"))),
			data: map[interface{}]interface{}{"name": "foo", "value": 12},
			test: func(t *testing.T, result string, err error) {
				require.Equal(t, `{"name":"foo","value":12}`, result)
			},
		},
		{
			name:   "map with key interface{} and empty properties",
			schema: schematest.New("object"),
			data:   map[interface{}]interface{}{"name": "foo", "value": 12},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				// order of properties is not guaranteed
				require.True(t, result == `{"name":"foo","value":12}` || result == `{"value":12,"name":"foo"}`, result)
			},
		},
		{
			name: "map as map",
			schema: schematest.New("object",
				schematest.WithAdditionalProperties(schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")))),
			),
			data: map[interface{}]interface{}{"x": map[string]string{"name": "x"}, "y": map[string]string{"name": "y"}},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.True(t, result == `{"x":{"name":"x"},"y":{"name":"y"}}` || result == `{"y":{"name":"y"},"x":{"name":"x"}}`, result)
			},
		},
		{
			name: "map not free-form",
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithFreeForm(false),
			),
			data: map[interface{}]interface{}{"name": "foo", "value": 12},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"name":"foo"}`, result)
			},
		},
		{
			name: "order by schema property definition",
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer")),
			),
			data: map[string]interface{}{"value": 12, "name": "foo"},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"name":"foo","value":12}`, result)
			},
		},
		{
			name: "dictionary",
			schema: schematest.New("object",
				schematest.WithAdditionalProperties(schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("value", schematest.New("integer")),
					schematest.WithProperty("bar", schematest.New("integer")),
				)),
			),
			data: map[string]interface{}{"foo": map[string]interface{}{"bar": 11, "value": 12, "name": "foo"}},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"foo":{"name":"foo","value":12,"bar":11}}`, result)
			},
		},
		{
			name: "additional properties false but data has additional",
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer")),
				schematest.WithFreeForm(false),
			),
			data: map[string]interface{}{"bar": 11, "value": 12, "name": "foo"},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"name":"foo","value":12}`, result)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := tc.schema.Marshal(tc.data, media.ParseContentType("application/json"))
			tc.test(t, string(b), err)
		})
	}
}

func TestRef_Marshal_Json_AnyOf(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		data   interface{}
		test   func(t *testing.T, result string, err error)
	}{
		{
			name: "integer or string, first matching is integer",
			schema: schematest.NewAny(
				schematest.New("integer"),
				schematest.New("string"),
			),
			data: 12,
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, "12", result)
			},
		},
		{
			name: "string or integer, integer not in first place",
			schema: schematest.NewAny(
				schematest.New("string"),
				schematest.New("integer"),
			),
			data: 12,
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, "12", result)
			},
		},
		{
			name: "any",
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("string"))),
			),
			data: map[string]interface{}{"foo": "foo"},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"foo":"foo"}`, result)
			},
		},
		{
			name: "any matches both",
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("string"))),
			),
			data: map[string]interface{}{"foo": "foo", "bar": "bar"},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"foo":"foo","bar":"bar"}`, result)
			},
		},
		{
			name: "any with free-form",
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string")), schematest.WithFreeForm(true)),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("string")), schematest.WithFreeForm(false)),
			),
			data: map[string]interface{}{"foo": "foo", "bar": "bar", "value": 12},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				var m map[string]interface{}
				err = json.Unmarshal([]byte(result), &m)
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"bar": "bar", "foo": "foo", "value": float64(12)}, m)
			},
		},
		{
			name: "any with one error should be skipped",
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("string"))),
			),
			data: map[string]interface{}{"foo": "foo", "bar": "bar"},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				// parse as map because order is random with free-form
				var m map[string]interface{}
				err = json.Unmarshal([]byte(result), &m)
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "foo", "bar": "bar"}, m)
			},
		},
		{
			name: "anyOf with additional field with free-form",
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("number"))),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("integer"))),
			),
			data: map[string]interface{}{"foo": 3.141, "bar": 12, "name": "foobar"},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				// parse as map because order is random with free-form
				var m map[string]interface{}
				err = json.Unmarshal([]byte(result), &m)
				require.Equal(t, map[string]interface{}{"foo": 3.141, "bar": float64(12), "name": "foobar"}, m)
			},
		},
		{
			name: "anyOf both invalid",
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string")), schematest.WithFreeForm(false)),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("string")), schematest.WithFreeForm(false)),
			),
			data: map[string]interface{}{"foo": "foo", "bar": "bar", "value": "test"},
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "encoding data to 'application/json' failed: error count 2:\n\t- #/anyOf/1/additionalProperties: does not match any schemas: property 'foo' not defined and the schema does not allow additional properties\n\t- #/anyOf/1/additionalProperties: does not match any schemas: property 'value' not defined and the schema does not allow additional properties")
			},
		},
		{
			name: "schema is nil and reference is nil",
			schema: schematest.NewAny(
				schematest.New(""),
				nil,
			),
			data: map[string]interface{}{"bar": "bar"},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"bar":"bar"}`, result)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := tc.schema.Marshal(tc.data, media.ParseContentType("application/json"))

			tc.test(t, string(b), err)
		})
	}
}

func TestRef_Marshal_Json_OneOf(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		data   interface{}
		test   func(t *testing.T, result string, err error)
	}{
		// examples on swagger documentation for oneOf are incorrect
		// https://swagger.io/docs/specification/data-models/oneof-anyof-allof-not/
		// https://github.com/swagger-api/swagger.io/issues/253
		{
			name: "example from swagger.io but throws an error because both matches",
			schema: schematest.NewOneOf(
				schematest.New("object",
					schematest.WithProperty("bark", schematest.New("boolean")),
					schematest.WithProperty("breed", schematest.New("string",
						schematest.WithEnumValues("Dingo", "Husky", "Retriever", "Shepherd")),
					),
				),
				schematest.New("object",
					schematest.WithProperty("hunts", schematest.New("boolean")),
					schematest.WithProperty("age", schematest.New("integer")),
				),
			),
			data: map[string]interface{}{"bark": true, "breed": "Dingo"},
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "encoding data to 'application/json' failed: error count 1:\n\t- #/oneOf: valid against more than one schema: valid schema indexes: 0, 1")
				require.Len(t, result, 0)
			},
		},
		{
			name: "example from swagger.io but with required properties",
			schema: schematest.NewOneOf(
				schematest.New("object",
					schematest.WithProperty("bark", schematest.New("boolean")),
					schematest.WithProperty("breed", schematest.New("string",
						schematest.WithEnum([]interface{}{"Dingo", "Husky", "Retriever", "Shepherd"})),
					),
					schematest.WithRequired("bark", "breed"),
				),
				schematest.New("object",
					schematest.WithProperty("hunts", schematest.New("boolean")),
					schematest.WithProperty("age", schematest.New("integer")),
					schematest.WithRequired("hunts", "age"),
				),
			),
			data: map[string]interface{}{"bark": true, "breed": "Dingo"},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"bark":true,"breed":"Dingo"}`, result)
			},
		},
		{
			name: "one schema is nil",
			schema: schematest.NewOneOf(
				schematest.New("object",
					schematest.WithProperty("bark", schematest.New("boolean")),
					schematest.WithProperty("breed", schematest.New("string",
						schematest.WithEnum([]interface{}{"Dingo", "Husky", "Retriever", "Shepherd"})),
					),
				),
				nil,
			),
			data: map[string]interface{}{"bark": true, "breed": "Dingo"},
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "encoding data to 'application/json' failed: error count 1:\n\t- #/oneOf: valid against more than one schema: valid schema indexes: 0, 1")
				require.Len(t, result, 0)
			},
		},
		{
			name: "one reference is nil",
			schema: schematest.NewOneOf(
				schematest.New("object",
					schematest.WithProperty("bark", schematest.New("boolean")),
					schematest.WithProperty("breed", schematest.New("string",
						schematest.WithEnum([]interface{}{"Dingo", "Husky", "Retriever", "Shepherd"})),
					),
				),
				nil,
			),
			data: map[string]interface{}{"bark": true, "breed": "Dingo"},
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "encoding data to 'application/json' failed: error count 1:\n\t- #/oneOf: valid against more than one schema: valid schema indexes: 0, 1")
				require.Len(t, result, 0)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := tc.schema.Marshal(tc.data, media.ParseContentType("application/json"))

			tc.test(t, string(b), err)
		})
	}
}

func TestRef_Marshal_Json_AllOf(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		data   func() interface{}
		test   func(t *testing.T, result string, err error)
	}{
		{
			name: "all from a map",
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("number"))),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("integer"))),
			),
			data: func() interface{} {
				return map[string]interface{}{"foo": 3.141, "bar": 12}
			},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"foo":3.141,"bar":12}`, result)
			},
		},
		{
			name: "all from a linked map",
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("number"))),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("integer"))),
			),
			data: func() interface{} {
				data := sortedmap.NewLinkedHashMap()
				data.Set("foo", 3.141)
				data.Set("bar", 12)
				return data
			},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"foo":3.141,"bar":12}`, result)
			},
		},
		{
			name: "all with additional field with free-form",
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("number"))),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("integer"))),
			),
			data: func() interface{} {
				data := sortedmap.NewLinkedHashMap()
				data.Set("foo", 3.141)
				data.Set("bar", 12)
				data.Set("name", "foobar")
				return data
			},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"foo":3.141,"bar":12,"name":"foobar"}`, result)
			},
		},
		{
			name: "allOf missing required property",
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string")), schematest.WithRequired("foo")),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("string"))),
			),
			data: func() interface{} {
				return map[string]interface{}{"bar": "bar"}
			},
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "encoding data to 'application/json' failed: error count 2:\n\t- #/allOf: does not match all schema\n\t\t- #/allOf/0/required: required properties are missing: foo")
				require.Len(t, result, 0)
			},
		},
		{
			name: "must be type of object",
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
				schematest.New("integer"),
			),
			data: func() interface{} {
				return map[string]interface{}{"bar": "bar"}
			},
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "encoding data to 'application/json' failed: error count 2:\n\t- #/allOf: does not match all schema\n\t\t- #/allOf/1/type: invalid type, expected integer but got object")
				require.Len(t, result, 0)
			},
		},
		{
			name: "schema is nil and reference is nil",
			schema: schematest.NewAllOfRefs(
				schematest.New(""),
				nil,
			),
			data: func() interface{} {
				return map[string]interface{}{"bar": "bar"}
			},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"bar":"bar"}`, result)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := tc.schema.Marshal(tc.data(), media.ParseContentType("application/json"))

			tc.test(t, string(b), err)
		})
	}
}

func TestRef_Marshal_Json_Invalid(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Schema
		data   interface{}
		exp    string
	}{
		{
			name:   "number",
			schema: schematest.New("number"),
			data:   "foo",
			exp:    "encoding data to 'application/json' failed: error count 1:\n\t- #/type: invalid type, expected number but got string",
		},
		{
			name:   "exclusiveMinimum",
			schema: schematest.New("integer", schematest.WithExclusiveMinimum(3)),
			data:   3,
			exp:    "encoding data to 'application/json' failed: error count 1:\n\t- #/exclusiveMinimum: integer 3 equals minimum value of 3",
		},
		{
			name:   "min array",
			schema: schematest.New("array", schematest.WithItems("integer"), schematest.WithMinItems(3)),
			data:   []interface{}{12, 13},
			exp:    "encoding data to 'application/json' failed: error count 1:\n\t- #/minItems: item count 2 is less than minimum count of 3",
		},
		{
			name:   "max array",
			schema: schematest.New("array", schematest.WithItems("integer"), schematest.WithMaxItems(1)),
			data:   []interface{}{12, 13},
			exp:    "encoding data to 'application/json' failed: error count 1:\n\t- #/maxItems: item count 2 exceeds maximum count of 1",
		},
		{
			name: "map missing required property",
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer")),
				schematest.WithRequired("value"),
			),
			data: map[interface{}]interface{}{"name": "foo"},
			exp:  "encoding data to 'application/json' failed: error count 1:\n\t- #/required: required properties are missing: value",
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.schema.Marshal(tc.data, media.ParseContentType("application/json"))
			require.EqualError(t, err, tc.exp)
		})
	}
}
