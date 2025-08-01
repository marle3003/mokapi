package schema_test

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/schema/json/schema"
	"testing"
)

func TestSchemaYaml(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, s *schema.Schema, err error)
	}{
		{
			name: "schema",
			data: `$schema: http://json-schema.org/draft-07/schema#`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "http://json-schema.org/draft-07/schema#", s.Schema)
			},
		},
		{
			name: "single type",
			data: `type: string`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, schema.Types{"string"}, s.Type)
			},
		},
		{
			name: "two types",
			data: `type: [string, integer]`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, schema.Types{"string", "integer"}, s.Type)
			},
		},
		{
			name: "type null",
			data: `type: "null"`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, schema.Types{"null"}, s.Type)
			},
		},
		{
			name: "type is not a string value",
			data: `type: [string, 123]`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.EqualError(t, err, "cannot unmarshal 123 into field type of type schema")
			},
		},
		{
			name: "one enum value",
			data: `enum: [foo]`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo"}, s.Enum)
			},
		},
		{
			name: "two enum values",
			data: `enum: [foo, 123]`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", 123}, s.Enum)
			},
		},
		{
			name: "const value",
			data: `const: foo`,
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
			data: `multipleOf: 12`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12.0, *s.MultipleOf)
			},
		},
		{
			name: "multipleOf can be a floating point number",
			data: `multipleOf: 12.5`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12.5, *s.MultipleOf)
			},
		},
		{
			name: "maximum",
			data: `maximum: 12`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(12), *s.Maximum)
			},
		},
		{
			name: "exclusiveMaximum",
			data: `exclusiveMaximum: 12`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(12), s.ExclusiveMaximum.A)
			},
		},
		{
			name: "minimum",
			data: `minimum: 12`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(12), *s.Minimum)
			},
		},
		{
			name: "exclusiveMinimum",
			data: `exclusiveMinimum: 12`,
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
			data: `maxLength: 12`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MaxLength)
			},
		},
		{
			name: "minLength",
			data: `minLength: 12`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MinLength)
			},
		},
		{
			name: "pattern",
			data: `pattern: "[a-z]"`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "[a-z]", s.Pattern)
			},
		},
		{
			name: "format",
			data: `format: date`,
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
			data: `maxItems: 12`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MaxItems)
			},
		},
		{
			name: "minItems",
			data: `minItems: 12`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MinItems)
			},
		},
		{
			name: "uniqueItems",
			data: `uniqueItems: true`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, true, s.UniqueItems)
			},
		},
		{
			name: "maxContains",
			data: `maxContains: 12`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MaxContains)
			},
		},
		{
			name: "minContains",
			data: `minContains: 12`,
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
			data: `
properties: 
  name: 
    type: string`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 1, s.Properties.Len())
			},
		},
		{
			name: "maxProperties",
			data: `maxProperties: 12`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MaxProperties)
			},
		},
		{
			name: "minProperties",
			data: `minProperties: 12`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MinProperties)
			},
		},
		{
			name: "required",
			data: `required: ["foo", "bar"]`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, s.Required)
			},
		},
		{
			name: "dependentRequired",
			data: `dependentRequired: {foo: [bar]}`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string][]string{"foo": {"bar"}}, s.DependentRequired)
			},
		},
		// Media
		{
			name: "contentMediaType",
			data: `contentMediaType: text/html`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "text/html", s.ContentMediaType)
			},
		},
		{
			name: "contentMediaType",
			data: `contentEncoding: base64`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "base64", s.ContentEncoding)
			},
		},
		{
			name: "not false",
			data: `not: false`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, false, *s.Not.Boolean)
			},
		},
		{
			name: "examples",
			data: `examples: [foo, bar]`,
			test: func(t *testing.T, s *schema.Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, []schema.Example{{Value: "foo"}, {Value: "bar"}}, s.Examples)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var s *schema.Schema
			err := yaml.Unmarshal([]byte(tc.data), &s)
			tc.test(t, s, err)
		})
	}
}

func TestYaml_Structuring(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{

		{
			name: "generic list of strings",
			test: func(t *testing.T) {
				reader := &dynamictest.Reader{
					Data: map[string]*dynamic.Config{
						"https://example.com/schemas/list-of-t": {
							Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/list-of-t")),
							Raw: []byte(`
$defs: 
  content: 
    $dynamicAnchor: T 
    not: true
type: array
items: 
  $dynamicRef: '#T'
`),
						},
					},
				}

				person := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(dynamictest.WithUrl("https://example.com/schemas/list-of-string")),
					Data: &schema.Schema{
						Defs: map[string]*schema.Schema{
							"string-items": {
								DynamicAnchor: "T",
								Type:          schema.Types{"string"},
							},
						},
						Ref: "https://example.com/schemas/list-of-t",
					},
				}

				err := person.Data.(*schema.Schema).Parse(person, reader)
				require.NoError(t, err)

				require.NoError(t, err)
				require.Equal(t, "string", person.Data.(*schema.Schema).Items.Type.String())
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
