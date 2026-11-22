package schema

import (
	encodingjson "encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/dynamictest"
	json "mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestConvert(t *testing.T) {
	testcases := []struct {
		name string
		s    *Schema
		test func(t *testing.T, js *json.Schema)
	}{
		{
			name: "empty",
			s:    &Schema{},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, &json.Schema{}, js)
			},
		},
		{
			name: "string",
			s:    &Schema{Type: []interface{}{"string"}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.New("string"), js)
			},
		},
		{
			name: "int",
			s:    &Schema{Type: []interface{}{"int"}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.New("integer", schematest.WithFormat("int32")), js)
			},
		},
		{
			name: "int wrapped",
			s: &Schema{Type: []interface{}{
				&Schema{Type: []interface{}{"int"}},
			}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.NewAny(schematest.New("integer", schematest.WithFormat("int32"))), js)
			},
		},
		{
			name: "long",
			s:    &Schema{Type: []interface{}{"long"}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.New("integer"), js)
			},
		},
		{
			name: "float",
			s:    &Schema{Type: []interface{}{"float"}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.New("number"), js)
			},
		},
		{
			name: "double",
			s:    &Schema{Type: []interface{}{"double"}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.New("number"), js)
			},
		},
		{
			name: "record",
			s:    &Schema{Type: []interface{}{"record"}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.New("object"), js)
			},
		},
		{
			name: "enum",
			s:    &Schema{Type: []interface{}{"enum"}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.New("string"), js)
			},
		},
		{
			name: "array",
			s:    &Schema{Type: []interface{}{"array"}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.New("array"), js)
			},
		},
		{
			name: "map",
			s:    &Schema{Type: []interface{}{"map"}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.New("object"), js)
			},
		},
		{
			name: "union",
			s:    &Schema{Type: []interface{}{"string", "int"}},
			test: func(t *testing.T, js *json.Schema) {
				require.Len(t, js.AnyOf, 2)
				require.Equal(t, "string", js.AnyOf[0].Type.String())
				require.Equal(t, "integer", js.AnyOf[1].Type.String())
			},
		},
		{
			name: "union schema",
			s: &Schema{Type: []interface{}{
				&Schema{Type: []interface{}{"string"}},
				&Schema{Type: []interface{}{"int"}},
			}},
			test: func(t *testing.T, js *json.Schema) {
				require.Len(t, js.AnyOf, 2)
				require.Equal(t, "string", js.AnyOf[0].Type.String())
				require.Equal(t, "integer", js.AnyOf[1].Type.String())
			},
		},
		{
			name: "fixed",
			s:    &Schema{Type: []interface{}{"fixed"}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.New("string", schematest.WithMinLength(0), schematest.WithMaxLength(0)), js)
			},
		},
		{
			name: "null",
			s:    &Schema{Type: []interface{}{"null"}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.New("null"), js)
			},
		},
		{
			name: "fields",
			s:    &Schema{Fields: []*Schema{{Name: "foo", Type: []interface{}{"string"}}}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.NewTypes(nil,
					schematest.WithProperty("foo",
						schematest.New("string", schematest.WithTitle("foo"))),
					schematest.WithRequired("foo")),
					js)
			},
		},
		{
			name: "symbols",
			s:    &Schema{Symbols: []string{"foo", "bar"}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.NewTypes(nil,
					schematest.WithEnum([]interface{}{"foo", "bar"}),
				), js)
			},
		},
		{
			name: "items",
			s:    &Schema{Items: &Schema{Type: []interface{}{"string"}}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.NewTypes(nil,
					schematest.WithItems("string"),
				), js)
			},
		},
		{
			name: "values",
			s:    &Schema{Values: &Schema{Type: []interface{}{"string"}}},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.NewTypes(nil,
					schematest.WithAdditionalProperties(schematest.New("string")),
				), js)
			},
		},
		{
			name: "fixed",
			s:    &Schema{Type: []interface{}{"fixed"}, Size: 16},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, schematest.New("string",
					schematest.WithMinLength(16),
					schematest.WithMaxLength(16)),
					js)
			},
		},
		{
			name: "enum wrapped",
			s: &Schema{Type: []interface{}{
				&Schema{
					Type:    []interface{}{"enum"},
					Symbols: []string{"foo", "bar"},
				},
			}},
			test: func(t *testing.T, js *json.Schema) {
				require.Len(t, js.AnyOf, 1)
				require.Equal(t, "string", js.AnyOf[0].Type.String())
				require.Equal(t, []interface{}{"foo", "bar"}, js.AnyOf[0].Enum)
			},
		},
		{
			name: "named enum",
			s: &Schema{
				Type: []interface{}{"record"},
				Fields: []*Schema{
					{
						Name: "f1",
						Type: []interface{}{
							&Schema{
								Name:    "foo",
								Type:    []interface{}{"enum"},
								Symbols: []string{"foo", "bar", "yuh"},
							},
						},
					},
					{
						Name: "f2",
						Type: []interface{}{"foo"},
					},
				},
			},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, "object", js.Type.String())

				f1 := js.Properties.Get("f1")
				require.Equal(t, "", f1.Type.String())
				require.Len(t, f1.AnyOf, 1)
				require.Equal(t, "string", f1.AnyOf[0].Type.String())
				require.Equal(t, []interface{}{"foo", "bar", "yuh"}, f1.AnyOf[0].Enum)

				f2 := js.Properties.Get("f2")
				require.Equal(t, "", f2.Type.String())
				require.Len(t, f2.AnyOf, 1)
				require.Equal(t, "string", f2.AnyOf[0].Type.String())
				require.Equal(t, []interface{}{"foo", "bar", "yuh"}, f2.AnyOf[0].Enum)
			},
		},
		{
			name: "record fields are required",
			s: &Schema{
				Type: []interface{}{"record"},
				Fields: []*Schema{
					{Name: "id", Type: []interface{}{"int"}},
					{Name: "name", Type: []interface{}{"null", "string"}},
				},
			},
			test: func(t *testing.T, js *json.Schema) {
				// Avro binary encoding writes every field of a record,
				// so the generator must never omit one
				require.Equal(t, []string{"id", "name"}, js.Required)
			},
		},
		{
			name: "nested record fields are required",
			s: &Schema{
				Type: []interface{}{"record"},
				Fields: []*Schema{
					{
						Name: "address",
						Type: []interface{}{
							&Schema{
								Type: []interface{}{"record"},
								Fields: []*Schema{
									{Name: "street", Type: []interface{}{"string"}},
									{Name: "zip", Type: []interface{}{"string"}},
								},
							},
						},
					},
				},
			},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, []string{"address"}, js.Required)
				address := js.Properties.Get("address").AnyOf[0]
				require.Equal(t, []string{"street", "zip"}, address.Required)
			},
		},
		{
			name: "null union field is required but allows null",
			s: &Schema{
				Type: []interface{}{"record"},
				Fields: []*Schema{
					{Name: "nickname", Type: []interface{}{"null", "string"}},
				},
			},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, []string{"nickname"}, js.Required)
				anyOf := js.Properties.Get("nickname").AnyOf
				require.Len(t, anyOf, 2)
				require.Equal(t, "null", anyOf[0].Type.String())
				require.Equal(t, "string", anyOf[1].Type.String())
			},
		},
		{
			name: "loop",
			s: &Schema{
				Type: []interface{}{"record"},
				Name: "List",
				Fields: []*Schema{
					{
						Name: "next",
						Type: []interface{}{"null", "List"},
					},
				},
			},
			test: func(t *testing.T, js *json.Schema) {
				require.Equal(t, "object", js.Type.String())
				require.Equal(t, 1, js.Properties.Len())
				anyOf := js.Properties.Get("next").AnyOf
				require.Equal(t, "schema type=null", anyOf[0].String())
				require.Equal(t, js, anyOf[1])
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			err := tc.s.Parse(dynamictest.NewConfig(tc.s, dynamictest.WithInfo(dynamictest.NewConfigInfo())), &dynamictest.Reader{})
			require.NoError(t, err)

			js := ConvertToJsonSchema(tc.s)
			tc.test(t, js)
		})
	}
}

func TestConvert_FieldWithDefaultIsRequired(t *testing.T) {
	var s Schema
	err := encodingjson.Unmarshal([]byte(`{"type":"record","name":"U","fields":[{"name":"role","type":"string","default":"guest"}]}`), &s)
	require.NoError(t, err)

	js := ConvertToJsonSchema(&s)

	require.Equal(t, []string{"role"}, js.Required)
}
