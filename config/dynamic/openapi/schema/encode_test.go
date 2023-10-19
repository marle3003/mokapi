package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/media"
	"mokapi/sortedmap"
	"testing"
)

func TestRef_Marshal(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Ref
		data   interface{}
		ct     media.ContentType
		exp    string
	}{
		{
			"no schema",
			&schema.Ref{},
			"foo",
			media.ParseContentType("application/json"),
			`"foo"`,
		},
		{
			"number",
			&schema.Ref{Value: schematest.New("number")},
			3.141,
			media.ParseContentType("application/json"),
			`3.141`,
		},
		{
			"string",
			&schema.Ref{Value: schematest.New("string")},
			"12",
			media.ParseContentType("application/json"),
			`"12"`,
		},
		{
			"integer as string",
			&schema.Ref{Value: schematest.New("integer")},
			"12",
			media.ParseContentType("application/json"),
			`12`,
		},
		{
			"array of integer",
			&schema.Ref{Value: schematest.New("array", schematest.WithItems(schematest.New("integer")))},
			[]interface{}{12, 13},
			media.ParseContentType("application/json"),
			`[12,13]`,
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := tc.schema.Marshal(tc.data, tc.ct)
			require.NoError(t, err)
			require.Equal(t, tc.exp, string(b))
		})
	}
}

func TestRef_Marshal_Object(t *testing.T) {
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
				require.EqualError(t, err, "serialize data to 'application/json' failed: encode '12' to schema type=object properties=[name, value] failed")
				require.Len(t, result, 0)
			},
		},
		{
			name: "struct",
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer"))),
			data: struct {
				Name  string
				Value int
			}{"foo", 12},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"name":"foo","value":12}`, result)
			},
		},
		{
			name: "struct pointer",
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer"))),
			data: &struct {
				Name  string
				Value int
			}{"foo", 12},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"name":"foo","value":12}`, result)
			},
		},
		{
			name: "struct but property matches not schema",
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer"))),
			data: struct {
				Name  string
				Value string
			}{"foo", "foo"},
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "serialize data to 'application/json' failed: encode property 'value' failed: could not parse 'foo' as int, expected schema type=integer")
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
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := schema.Ref{Value: tc.schema}
			b, err := r.Marshal(tc.data, media.ParseContentType("application/json"))
			tc.test(t, string(b), err)
		})
	}
}

func TestRef_Marshal_AnyOf(t *testing.T) {
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
			data: "12",
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, "12", result)
			},
		},
		{
			name: "string or integer, first matching is integer",
			schema: schematest.NewAny(
				schematest.New("string"),
				schematest.New("integer"),
			),
			data: "12",
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, "\"12\"", result)
			},
		},
		{
			name: "any",
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string")), schematest.WithFreeForm(false)),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("string")), schematest.WithFreeForm(false)),
			),
			data: map[string]interface{}{"foo": "foo", "value": 12},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"foo":"foo"}`, result)
			},
		},
		{
			name: "any matches both",
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string")), schematest.WithFreeForm(false)),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("string")), schematest.WithFreeForm(false)),
			),
			data: map[string]interface{}{"foo": "foo", "bar": "bar", "value": 12},
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
				require.Equal(t, `{"foo":"foo","bar":"bar","value":12}`, result)
			},
		},
		{
			name: "any with one error should be skipped",
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("integer")), schematest.WithFreeForm(false)),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("string")), schematest.WithFreeForm(false)),
			),
			data: map[string]interface{}{"foo": "foo", "bar": "bar", "value": 12},
			test: func(t *testing.T, result string, err error) {
				s := &schema.Ref{Value: schematest.New("",
					schematest.Any(
						schematest.New("object", schematest.WithProperty("foo", schematest.New("integer")), schematest.WithFreeForm(false)),
						schematest.New("object", schematest.WithProperty("bar", schematest.New("string")), schematest.WithFreeForm(false)),
					))}
				data := map[string]interface{}{"foo": "foo", "bar": "bar", "value": 12}

				b, err := s.Marshal(data, media.ParseContentType("application/json"))
				require.NoError(t, err)
				require.Equal(t, `{"bar":"bar"}`, string(b))
			},
		},
		{
			name: "anyOf with additional field with free-form",
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("number"))),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("integer"))),
			),
			data: map[string]interface{}{"foo": "3.141", "bar": "12", "name": "foobar"},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"foo":3.141,"bar":12,"name":"foobar"}`, result)
			},
		},
		{
			name: "anyOf ignores additional properties when free-form is false",
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string")), schematest.WithFreeForm(false)),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("string")), schematest.WithFreeForm(false)),
			),
			data: map[string]interface{}{"foo": "foo", "bar": "bar", "value": "test"},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"foo":"foo","bar":"bar"}`, result)
			},
		},
		{
			name: "schema is nil and reference is nil",
			schema: schematest.NewAnyRef(
				&schema.Ref{},
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

			r := &schema.Ref{Value: tc.schema}
			b, err := r.Marshal(tc.data, media.ParseContentType("application/json"))

			tc.test(t, string(b), err)
		})
	}
}

func TestRef_Marshal_OneOf(t *testing.T) {
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
						schematest.WithEnum([]interface{}{"Dingo", "Husky", "Retriever", "Shepherd"})),
					),
				),
				schematest.New("object",
					schematest.WithProperty("hunts", schematest.New("boolean")),
					schematest.WithProperty("age", schematest.New("integer")),
				),
			),
			data: map[string]interface{}{"bark": true, "breed": "Dingo"},
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "serialize data to 'application/json' failed: oneOf can only match exactly one schema")
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
			name: "example from swagger.io but cat does not allow additional properties",
			schema: schematest.NewOneOf(
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
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"bark":true,"breed":"Dingo"}`, result)
			},
		},
		{
			name: "schema is nil = free form",
			schema: schematest.NewOneOf(
				nil,
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
				require.EqualError(t, err, "serialize data to 'application/json' failed: oneOf can only match exactly one schema")
				require.Len(t, result, 0)
			},
		},
		{
			name: "one reference is nil",
			schema: schematest.NewOneOfRef(
				&schema.Ref{Value: schematest.New("object",
					schematest.WithProperty("bark", schematest.New("boolean")),
					schematest.WithProperty("breed", schematest.New("string",
						schematest.WithEnum([]interface{}{"Dingo", "Husky", "Retriever", "Shepherd"})),
					),
				)},
				nil,
			),
			data: map[string]interface{}{"bark": true, "breed": "Dingo"},
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "serialize data to 'application/json' failed: oneOf can only match exactly one schema")
				require.Len(t, result, 0)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &schema.Ref{Value: tc.schema}
			b, err := r.Marshal(tc.data, media.ParseContentType("application/json"))

			tc.test(t, string(b), err)
		})
	}
}

func TestRef_Marshal_AllOf(t *testing.T) {
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
				return map[string]interface{}{"foo": "3.141", "bar": "12"}
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
				data.Set("foo", "3.141")
				data.Set("bar", "12")
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
				data.Set("foo", "3.141")
				data.Set("bar", "12")
				data.Set("name", "foobar")
				return data
			},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"foo":3.141,"bar":12,"name":"foobar"}`, result)
			},
		},
		{
			name: "allOf ignores additional properties when free-form is false",
			schema: schematest.NewAllOf(
				schematest.New("object", schematest.WithProperty("foo", schematest.New("string")), schematest.WithFreeForm(false)),
				schematest.New("object", schematest.WithProperty("bar", schematest.New("string")), schematest.WithFreeForm(false)),
			),
			data: func() interface{} {
				return map[string]interface{}{"foo": "foo", "bar": "bar", "value": "test"}
			},
			test: func(t *testing.T, result string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"foo":"foo","bar":"bar"}`, result)
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
				require.EqualError(t, err, "serialize data to 'application/json' failed: does not match schema type=object properties=[foo] required=[foo]: missing required field 'foo'")
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
				require.EqualError(t, err, "serialize data to 'application/json' failed: type of 'integer' is not allowed: allOf only supports type of object")
				require.Len(t, result, 0)
			},
		},
		{
			name: "schema is nil and reference is nil",
			schema: schematest.NewAllOfRefs(
				&schema.Ref{},
				nil,
			),
			data: func() interface{} {
				return map[string]interface{}{"bar": "bar"}
			},
			test: func(t *testing.T, result string, err error) {
				require.EqualError(t, err, "serialize data to 'application/json' failed: schema is not defined: allOf only supports type of object")
				require.Len(t, result, 0)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &schema.Ref{Value: tc.schema}
			b, err := r.Marshal(tc.data(), media.ParseContentType("application/json"))

			tc.test(t, string(b), err)
		})
	}
}

func TestRef_Marshal_Invalid(t *testing.T) {
	testcases := []struct {
		name   string
		schema *schema.Ref
		data   interface{}
		ct     media.ContentType
		exp    string
	}{
		{
			"number",
			&schema.Ref{Value: schematest.New("number")},
			"foo",
			media.ParseContentType("application/json"),
			"serialize data to 'application/json' failed: could not parse 'foo' as floating number, expected schema type=number",
		},
		{
			"min array",
			&schema.Ref{Value: schematest.New("array", schematest.WithItems(schematest.New("integer")), schematest.WithMinItems(3))},
			[]interface{}{12, 13},
			media.ParseContentType("application/json"),
			`serialize data to 'application/json' failed: does not match schema type=array minItems=3 items=schema type=integer: should NOT have less than 3 items`,
		},
		{
			"max array",
			&schema.Ref{Value: schematest.New("array", schematest.WithItems(schematest.New("integer")), schematest.WithMaxItems(1))},
			[]interface{}{12, 13},
			media.ParseContentType("application/json"),
			`serialize data to 'application/json' failed: does not match schema type=array maxItems=1 items=schema type=integer: should NOT have more than 1 items`,
		},
		{
			"map missing required property",
			&schema.Ref{Value: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("value", schematest.New("integer")),
				schematest.WithRequired("value"),
			)},
			map[interface{}]interface{}{"name": "foo"},
			media.ParseContentType("application/json"),
			`serialize data to 'application/json' failed: does not match schema type=object properties=[name, value] required=[value]: missing required field 'value'`,
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tc.schema.Marshal(tc.data, tc.ct)
			require.EqualError(t, err, tc.exp)
		})
	}
}
