package schema_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestSchemaJson(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, s *schema.Schema, err error)
	}{
		{
			name: "schema",
			data: `{"$schema": "http://json-schema.org/draft-07/schema#"}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://json-schema.org/draft-07/schema#", s.Schema)
			},
		},
		{
			name: "single type",
			data: `{"type": "string"}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, schema.Types{"string"}, s.Type)
			},
		},
		{
			name: "two types",
			data: `{"type": ["string", "integer"] }`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, schema.Types{"string", "integer"}, s.Type)
			},
		},
		{
			name: "type null",
			data: `{"type": "null" }`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, schema.Types{"null"}, s.Type)
			},
		},
		{
			name: "type is not a string value",
			data: `{"type": ["string", 123] }`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.EqualError(t, err, "cannot unmarshal 123 into field type of type schema")
			},
		},
		{
			name: "one enum value",
			data: `{"enum": ["foo"]}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo"}, s.Enum)
			},
		},
		{
			name: "two enum values",
			data: `{"enum": ["foo", 123] }`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", float64(123)}, s.Enum)
			},
		},
		{
			name: "const value",
			data: `{"const": "foo"}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", *s.Const)
			},
		},
		/*
		 * Numbers
		 */
		{
			name: "multipleOf",
			data: `{"multipleOf": 12}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12.0, *s.MultipleOf)
			},
		},
		{
			name: "multipleOf can be a floating point number",
			data: `{"multipleOf": 12.5}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12.5, *s.MultipleOf)
			},
		},
		{
			name: "maximum",
			data: `{"maximum": 12}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(12), *s.Maximum)
			},
		},
		{
			name: "exclusiveMaximum",
			data: `{"exclusiveMaximum": 12}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(12), s.ExclusiveMaximum.A)
			},
		},
		{
			name: "minimum",
			data: `{"minimum": 12}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(12), *s.Minimum)
			},
		},
		{
			name: "exclusiveMinimum",
			data: `{"exclusiveMinimum": 12}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(12), s.ExclusiveMinimum.A)
			},
		},
		/*
		 * Strings
		 */
		{
			name: "maxLength",
			data: `{"maxLength": 12}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MaxLength)
			},
		},
		{
			name: "minLength",
			data: `{"minLength": 12}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MinLength)
			},
		},
		{
			name: "pattern",
			data: `{"pattern": "[a-z]"}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "[a-z]", s.Pattern)
			},
		},
		{
			name: "format",
			data: `{"format": "date"}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "date", s.Format)
			},
		},
		/*
		 * Arrays
		 */
		{
			name: "maxItems",
			data: `{"maxItems": 12}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MaxItems)
			},
		},
		{
			name: "minItems",
			data: `{"minItems": 12}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MinItems)
			},
		},
		{
			name: "uniqueItems",
			data: `{"uniqueItems": true}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, true, s.UniqueItems)
			},
		},
		{
			name: "maxContains",
			data: `{"maxContains": 12}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MaxContains)
			},
		},
		{
			name: "minContains",
			data: `{"minContains": 12}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MinContains)
			},
		},
		/*
		 * Objects
		 */
		{
			name: "properties",
			data: `{"properties": {"name": {"type": "string"} }}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 1, s.Properties.Len())
			},
		},
		{
			name: "maxProperties",
			data: `{"maxProperties": 12}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MaxProperties)
			},
		},
		{
			name: "minProperties",
			data: `{"minProperties": 12}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MinProperties)
			},
		},
		{
			name: "required",
			data: `{"required": ["foo", "bar"]}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, s.Required)
			},
		},
		{
			name: "dependentRequired",
			data: `{"dependentRequired": {"foo": ["bar"]} }`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string][]string{"foo": {"bar"}}, s.DependentRequired)
			},
		},
		// Media
		{
			name: "contentMediaType",
			data: `{"contentMediaType": "text/html" }`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "text/html", s.ContentMediaType)
			},
		},
		{
			name: "contentMediaType",
			data: `{"contentEncoding": "base64" }`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "base64", s.ContentEncoding)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var s *schema.Schema
			err := json.Unmarshal([]byte(tc.data), &s)
			tc.test(t, s, err)
		})
	}
}

func TestSchema_MarshalJSON(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Schema
		test func(t *testing.T, s string, err error)
	}{
		{
			name: "empty type",
			s:    &schema.Schema{},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, "{}", s)
			},
		},
		{
			name: "one type",
			s:    &schema.Schema{Type: schema.Types{"string"}},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"type":"string"}`, s)
			},
		},
		{
			name: "two types",
			s:    &schema.Schema{Type: schema.Types{"string", "number"}},
			test: func(t *testing.T, s string, err error) {
				require.NoError(t, err)
				require.Equal(t, `{"type":["string","number"]}`, s)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b, err := json.Marshal(tc.s)
			tc.test(t, string(b), err)
		})
	}
}

func TestJson_Structuring(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "JSON pointer",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schemas/address": {
							Data: schematest.New("object",
								schematest.WithProperty("street_address", schematest.New("string")),
							),
						},
					},
				}
				r := &schema.Schema{}
				err := dynamic.Resolve("https://example.com/schemas/address#/properties/street_address", &r, &dynamic.Config{Data: &schema.Schema{}}, reader)
				require.NoError(t, err)
				require.Equal(t, "string", r.Type.String())
			},
		},
		{
			name: "$anchor",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schemas/address": {
							Data: schematest.New("object",
								schematest.WithProperty("street_address",
									schematest.New("string", schematest.WithAnchor("street_address"))),
							),
						},
					},
				}
				r := &schema.Schema{}
				err := dynamic.Resolve("https://example.com/schemas/address#street_address", &r, &dynamic.Config{Data: &schema.Schema{}}, reader)
				require.NoError(t, err)
				require.Equal(t, "string", r.Type.String())
			},
		},
		{
			name: "relative to $id",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schemas/address": {
							Data: schematest.New("object",
								schematest.WithProperty("street_address", schematest.New("string")),
							),
						},
					},
				}

				cfg := &dynamic.Config{Data: &schema.Schema{Id: "https://example.com/schemas/customer"}}

				r := &schema.Schema{}
				err := dynamic.Resolve("/schemas/address", &r, cfg, reader)
				require.NoError(t, err)
				require.NotNil(t, r)
				require.Equal(t, "object", r.Type.String())
			},
		},
		{
			name: "$defs",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithPropertyRef("first_name", "#/$defs/name"),
					schematest.WithDef("name", schematest.New("string")),
				)

				err := s.Parse(&dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				require.Equal(t, "string", s.Properties.Get("first_name").Type.String())
			},
		},
		{
			name: "recursion",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("children",
						schematest.New("array", schematest.WithItemsRefString("#")),
					),
				)

				err := s.Parse(&dynamic.Config{Data: s}, &dynamictest.Reader{})

				require.NoError(t, err)
				children := s.Properties.Get("children")
				require.Equal(t, s, children.Items)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t)
		})
	}
}
