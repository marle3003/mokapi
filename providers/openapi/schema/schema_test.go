package schema_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"testing"
)

func TestSchema_UnmarshalJSON(t *testing.T) {
	for _, testcase := range []struct {
		name string
		s    string
		fn   func(t *testing.T, r *schema.Ref)
	}{
		{
			name: "default",
			s:    `{ "type": "string" }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Nil(t, r.Value.MinLength)
				require.Nil(t, r.Value.MaxLength)
				require.Nil(t, r.Value.Minimum)
				require.Nil(t, r.Value.Maximum)
				require.Nil(t, r.Value.ExclusiveMinimum)
				require.Nil(t, r.Value.ExclusiveMaximum)
				require.Nil(t, r.Value.MinItems)
				require.Nil(t, r.Value.MaxItems)
				require.Nil(t, r.Value.MinProperties)
				require.Nil(t, r.Value.MaxProperties)
			},
		},
		{
			name: "type",
			s:    `{ "type": "string" }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, "string", r.Value.Type.String())
			},
		},
		{
			name: "description",
			s:    `{ "description": "foo" }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, "foo", r.Value.Description)
			},
		},
		{
			name: "anyOf",
			s:    `{ "anyOf": [ { "type": "string" }, { "type": "number" } ] }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Len(t, r.Value.AnyOf, 2)
				require.Equal(t, "string", r.Value.AnyOf[0].Value.Type.String())
				require.Equal(t, "number", r.Value.AnyOf[1].Value.Type.String())
			},
		},
		{
			name: "allOf",
			s:    `{ "allOf": [ { "type": "string" }, { "type": "number" } ] }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Len(t, r.Value.AllOf, 2)
				require.Equal(t, "string", r.Value.AllOf[0].Value.Type.String())
				require.Equal(t, "number", r.Value.AllOf[1].Value.Type.String())
			},
		},
		{
			name: "oneOf",
			s:    `{ "oneOf": [ { "type": "string" }, { "type": "number" } ] }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Len(t, r.Value.OneOf, 2)
				require.Equal(t, "string", r.Value.OneOf[0].Value.Type.String())
				require.Equal(t, "number", r.Value.OneOf[1].Value.Type.String())
			},
		},
		{
			name: "deprecated",
			s:    `{ "deprecated": true }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.True(t, r.Value.Deprecated)
			},
		},
		{
			name: "example value",
			s:    `{ "example": 12 }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, float64(12), r.Value.Example)
			},
		},
		{
			name: "example array",
			s:    `{ "example": [1,2,3] }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, []interface{}{float64(1), float64(2), float64(3)}, r.Value.Example)
			},
		},
		{
			name: "example object",
			s:    `{ "example": { "id": 1, "name": "Jessica Smith" } }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, map[string]interface{}{"id": float64(1), "name": "Jessica Smith"}, r.Value.Example)
			},
		},
		{
			name: "enum value",
			s:    `{ "enum": [ 12 ] }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, []interface{}{float64(12)}, r.Value.Enum)
			},
		},
		{
			name: "enum array",
			s:    `{ "enum": [ [1,2,3], [9,8,7] ] }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, []interface{}{
					[]interface{}{float64(1), float64(2), float64(3)},
					[]interface{}{float64(9), float64(8), float64(7)},
				}, r.Value.Enum)
			},
		},
		{
			name: "enum object",
			s:    `{ "enum": [ { "id": 1, "name": "Jessica Smith" }, { "id": 2, "name": "Ron Stewart" } ] }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, []interface{}{
					map[string]interface{}{"id": float64(1), "name": "Jessica Smith"},
					map[string]interface{}{"id": float64(2), "name": "Ron Stewart"},
				}, r.Value.Enum)
			},
		},
		{
			name: "xml",
			s:    `{ "xml": { "wrapped": true, "name": "foo", "attribute": true, "prefix": "bar", "namespace": "ns1"} }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, &schema.Xml{
					Wrapped:   true,
					Name:      "foo",
					Attribute: true,
					Prefix:    "bar",
					Namespace: "ns1",
				}, r.Value.Xml)
			},
		},
		{
			name: "format",
			s:    `{ "format": "foo" }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, "foo", r.Value.Format)
			},
		},
		{
			name: "nullable",
			s:    `{ "nullable": true }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.True(t, r.Value.Nullable)
			},
		},
		{
			name: "pattern",
			s:    `{ "pattern": "[A-Z]{4}" }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, "[A-Z]{4}", r.Value.Pattern)
			},
		},
		{
			name: "minLength",
			s:    `{ "minLength": 3 }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, 3, *r.Value.MinLength)
			},
		},
		{
			name: "maxLength",
			s:    `{ "maxLength": 3 }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, 3, *r.Value.MaxLength)
			},
		},
		{
			name: "minimum",
			s:    `{ "minimum": 3 }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, float64(3), *r.Value.Minimum)
			},
		},
		{
			name: "maximum",
			s:    `{ "maximum": 3 }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, float64(3), *r.Value.Maximum)
			},
		},
		{
			name: "exclusiveMinimum",
			s:    `{ "exclusiveMinimum": true }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.True(t, *r.Value.ExclusiveMinimum)
			},
		},
		{
			name: "exclusiveMaximum",
			s:    `{ "exclusiveMaximum": true }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.True(t, *r.Value.ExclusiveMaximum)
			},
		},
		{
			name: "items",
			s:    `{ "items": { "type": "object" } }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, "object", r.Value.Items.Value.Type.String())
			},
		},
		{
			name: "uniqueItems",
			s:    `{ "uniqueItems": true }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.True(t, r.Value.UniqueItems)
			},
		},
		{
			name: "minItems",
			s:    `{ "minItems": 3 }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, 3, *r.Value.MinItems)
			},
		},
		{
			name: "maxItems",
			s:    `{ "maxItems": 3 }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, 3, *r.Value.MaxItems)
			},
		},
		{
			name: "properties true",
			s:    `{ "type": "object", "properties": { "name": { "type": "string" } } }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, "object", r.Value.Type.String())
				require.Equal(t, 1, r.Value.Properties.Len())
				name := r.Value.Properties.Get("name")
				require.NotNil(t, name)
				require.Equal(t, "string", name.Value.Type.String())
			},
		},
		{
			name: "required",
			s:    `{ "required": ["name"] }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, []string{"name"}, r.Value.Required)
			},
		},
		{
			name: "additional properties true",
			s:    `{ "type": "object", "additionalProperties": true }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, "object", r.Value.Type.String())
				require.NotNil(t, r.Value.AdditionalProperties)
			},
		},
		{
			name: "additional properties {}",
			s:    `{ "type": "object", "additionalProperties": {} }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, "object", r.Value.Type.String())
				require.NotNil(t, r.Value.AdditionalProperties)
			},
		},
		{
			name: "minProperties",
			s:    `{ "minProperties": 3 }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, 3, *r.Value.MinProperties)
			},
		},
		{
			name: "maxProperties",
			s:    `{ "maxProperties": 3 }`,
			fn: func(t *testing.T, r *schema.Ref) {
				require.Equal(t, 3, *r.Value.MaxProperties)
			},
		},
	} {
		test := testcase
		t.Run(test.name, func(t *testing.T) {
			r := &schema.Ref{}
			err := json.Unmarshal([]byte(test.s), r)
			require.NoError(t, err)
			test.fn(t, r)
		})
	}
}

func TestSchema_UnmarshalYAML(t *testing.T) {
	testcases := []struct {
		name string
		s    string
		fn   func(t *testing.T, schema *schema.Schema)
	}{
		{
			"empty",
			"",
			func(t *testing.T, schema *schema.Schema) {
				require.Equal(t, "", schema.Type.String())
			},
		},
		{
			name: "format",
			s: `
format: foo
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, "foo", r.Format)
			},
		},
		{
			name: "nullable",
			s: `
nullable: true
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.True(t, r.Nullable)
			},
		},
		{
			name: "pattern",
			s: `
pattern: '[A-Z]{4}'
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, "[A-Z]{4}", r.Pattern)
			},
		},
		{
			name: "minLength",
			s: `
minLength: 3
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, 3, *r.MinLength)
			},
		},
		{
			name: "maxLength",
			s: `
maxLength: 3
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, 3, *r.MaxLength)
			},
		},
		{
			name: "minimum",
			s: `
minimum: 3
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, float64(3), *r.Minimum)
			},
		},
		{
			name: "maximum",
			s: `
maximum: 3
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, float64(3), *r.Maximum)
			},
		},
		{
			name: "items",
			s: `
items:
  type: object
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, "object", r.Items.Value.Type.String())
			},
		},
		{
			name: "uniqueItems",
			s: `
uniqueItems: true
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.True(t, r.UniqueItems)
			},
		},
		{
			name: "minItems",
			s: `
minItems: 3
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, 3, *r.MinItems)
			},
		},
		{
			name: "maxItems",
			s: `
maxItems: 3
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, 3, *r.MaxItems)
			},
		},
		{
			name: "exclusiveMinimum",
			s: `
exclusiveMinimum: true
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.True(t, *r.ExclusiveMinimum)
			},
		},
		{
			name: "exclusiveMaximum",
			s: `
exclusiveMaximum: true
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.True(t, *r.ExclusiveMaximum)
			},
		},
		{
			name: "properties true",
			s: `
type: object
properties:
  name:
    type: string
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, "object", r.Type.String())
				require.Equal(t, 1, r.Properties.Len())
				name := r.Properties.Get("name")
				require.NotNil(t, name)
				require.Equal(t, "string", name.Value.Type.String())
			},
		},
		{
			"additional properties false",
			`
type: object
additionalProperties: false
properties:
  name:
    type: string
`,
			func(t *testing.T, schema *schema.Schema) {
				require.Equal(t, "object", schema.Type.String())
				require.False(t, schema.IsFreeForm(), "object should not be free form")
				require.False(t, schema.IsDictionary())
			},
		},
		{
			"additional properties true",
			`
type: object
additionalProperties: true
properties:
  name:
    type: string
`,
			func(t *testing.T, schema *schema.Schema) {
				require.Equal(t, "object", schema.Type.String())
				require.True(t, schema.IsFreeForm(), "object should be free form")
				require.False(t, schema.IsDictionary())
			},
		},
		{
			"additional properties",
			`
type: object
additionalProperties: {}
`,
			func(t *testing.T, schema *schema.Schema) {
				require.Equal(t, "object", schema.Type.String())
				require.True(t, schema.IsFreeForm(), "object should be free form")
			},
		},
		{
			"additional properties",
			`
type: object
additionalProperties:
  type: string
properties:
  name:
    type: string
`,
			func(t *testing.T, schema *schema.Schema) {
				require.Equal(t, "object", schema.Type.String())
				require.False(t, schema.IsFreeForm())
				require.Equal(t, "string", schema.AdditionalProperties.Value.Type.String())
			},
		},
		{
			name: "anyOf",
			s: `
anyOf:
  - type: string
  - type: number
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Len(t, r.AnyOf, 2)
				require.Equal(t, "string", r.AnyOf[0].Value.Type.String())
				require.Equal(t, "number", r.AnyOf[1].Value.Type.String())
			},
		},

		{
			name: "oneOf",
			s: `
oneOf:
  - type: string
  - type: number
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Len(t, r.OneOf, 2)
				require.Equal(t, "string", r.OneOf[0].Value.Type.String())
				require.Equal(t, "number", r.OneOf[1].Value.Type.String())
			},
		},
		{
			name: "allOf",
			s: `
allOf:
  - type: string
  - type: number
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Len(t, r.AllOf, 2)
				require.Equal(t, "string", r.AllOf[0].Value.Type.String())
				require.Equal(t, "number", r.AllOf[1].Value.Type.String())
			},
		},
		{
			name: "required",
			s: `
required: [name]
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, []string{"name"}, r.Required)
			},
		},
		{
			name: "minProperties",
			s: `
minProperties: 3
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, 3, *r.MinProperties)
			},
		},
		{
			name: "maxProperties",
			s: `
maxProperties: 3
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, 3, *r.MaxProperties)
			},
		},
		{
			name: "enum value",
			s: `
enum: [12]
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, []interface{}{12}, r.Enum)
			},
		},
		{
			name: "enum array",
			s: `
enum: 
  - [1,2,3]
  - [9,8,7]
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, []interface{}{
					[]interface{}{1, 2, 3},
					[]interface{}{9, 8, 7},
				}, r.Enum)
			},
		},
		{
			"enum object",
			`
type: object
enum:
  - name: alice
    age: 29
`,
			func(t *testing.T, schema *schema.Schema) {
				require.Equal(t, "object", schema.Type.String())
				require.Len(t, schema.Enum, 1)
				require.Equal(t, map[string]interface{}{"name": "alice", "age": 29}, schema.Enum[0])
			},
		},
		{
			name: "example value",
			s: `
example: 12
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, 12, r.Example)
			},
		},
		{
			name: "example array",
			s: `
example: [1,2,3]
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, []interface{}{1, 2, 3}, r.Example)
			},
		},
		{
			name: "example object",
			s: `
example: 
  id: 1
  name: Jessica Smith
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, map[string]interface{}{"id": 1, "name": "Jessica Smith"}, r.Example)
			},
		},
		{
			name: "enum object",
			s: `
example: 
  - id: 1
    name: Jessica Smith
  - id: 2
    name: Ron Stewart
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.Equal(t, []interface{}{
					map[string]interface{}{"id": 1, "name": "Jessica Smith"},
					map[string]interface{}{"id": 2, "name": "Ron Stewart"},
				}, r.Example)
			},
		},
		{
			name: "deprecated",
			s: `
deprecated: true
`,
			fn: func(t *testing.T, r *schema.Schema) {
				require.True(t, r.Deprecated)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := &schema.Schema{}
			err := yaml.Unmarshal([]byte(tc.s), &s)
			require.NoError(t, err)
			tc.fn(t, s)
		})
	}
}

func TestSchema_IsFreeForm(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"number",
			func(t *testing.T) {
				s := schematest.New("number")
				require.False(t, s.IsFreeForm())
			},
		},
		{
			"object with property and additionalProperty null",
			func(t *testing.T) {
				s := schematest.New("object", schematest.WithProperty("foo", schematest.New("string")))
				require.True(t, s.IsFreeForm())
			},
		},
		{
			"object without property",
			func(t *testing.T) {
				s := schematest.New("object")
				require.True(t, s.IsFreeForm())
			},
		},
		{
			"object with empty additional properties",
			func(t *testing.T) {
				s := schematest.New("object")
				s.AdditionalProperties = &schema.AdditionalProperties{}
				require.True(t, s.IsFreeForm())
			},
		},
		{
			"object with property additional false",
			func(t *testing.T) {
				s := schematest.New("object", schematest.WithProperty("foo", schematest.New("string")), schematest.WithFreeForm(false))
				require.False(t, s.IsFreeForm())
			},
		},
		{
			"object with property additional true",
			func(t *testing.T) {
				s := schematest.New("object", schematest.WithProperty("foo", schematest.New("string")), schematest.WithFreeForm(true))
				require.True(t, s.IsFreeForm())
			},
		},
	}
	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t)
		})
	}
}
