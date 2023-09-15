package schema

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRef_UnmarshalJSON(t *testing.T) {
	for _, testcase := range []struct {
		name string
		s    string
		fn   func(t *testing.T, r *Ref)
	}{
		{
			name: "ref",
			s:    `{ "$ref": "#/components/schema/Foo" }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, "#/components/schema/Foo", r.Ref)
			},
		},
		{
			name: "default",
			s:    `{ "type": "string" }`,
			fn: func(t *testing.T, r *Ref) {
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
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, "string", r.Value.Type)
			},
		},
		{
			name: "description",
			s:    `{ "description": "foo" }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, "foo", r.Value.Description)
			},
		},
		{
			name: "anyOf",
			s:    `{ "anyOf": [ { "type": "string" }, { "type": "number" } ] }`,
			fn: func(t *testing.T, r *Ref) {
				require.Len(t, r.Value.AnyOf, 2)
				require.Equal(t, "string", r.Value.AnyOf[0].Value.Type)
				require.Equal(t, "number", r.Value.AnyOf[1].Value.Type)
			},
		},
		{
			name: "allOf",
			s:    `{ "allOf": [ { "type": "string" }, { "type": "number" } ] }`,
			fn: func(t *testing.T, r *Ref) {
				require.Len(t, r.Value.AllOf, 2)
				require.Equal(t, "string", r.Value.AllOf[0].Value.Type)
				require.Equal(t, "number", r.Value.AllOf[1].Value.Type)
			},
		},
		{
			name: "oneOf",
			s:    `{ "oneOf": [ { "type": "string" }, { "type": "number" } ] }`,
			fn: func(t *testing.T, r *Ref) {
				require.Len(t, r.Value.OneOf, 2)
				require.Equal(t, "string", r.Value.OneOf[0].Value.Type)
				require.Equal(t, "number", r.Value.OneOf[1].Value.Type)
			},
		},
		{
			name: "deprecated",
			s:    `{ "deprecated": true }`,
			fn: func(t *testing.T, r *Ref) {
				require.True(t, r.Value.Deprecated)
			},
		},
		{
			name: "example value",
			s:    `{ "example": 12 }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, float64(12), r.Value.Example)
			},
		},
		{
			name: "example array",
			s:    `{ "example": [1,2,3] }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, []interface{}{float64(1), float64(2), float64(3)}, r.Value.Example)
			},
		},
		{
			name: "example object",
			s:    `{ "example": { "id": 1, "name": "Jessica Smith" } }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, map[string]interface{}{"id": float64(1), "name": "Jessica Smith"}, r.Value.Example)
			},
		},
		{
			name: "enum value",
			s:    `{ "enum": [ 12 ] }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, []interface{}{float64(12)}, r.Value.Enum)
			},
		},
		{
			name: "enum array",
			s:    `{ "enum": [ [1,2,3], [9,8,7] ] }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, []interface{}{
					[]interface{}{float64(1), float64(2), float64(3)},
					[]interface{}{float64(9), float64(8), float64(7)},
				}, r.Value.Enum)
			},
		},
		{
			name: "enum object",
			s:    `{ "enum": [ { "id": 1, "name": "Jessica Smith" }, { "id": 2, "name": "Ron Stewart" } ] }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, []interface{}{
					map[string]interface{}{"id": float64(1), "name": "Jessica Smith"},
					map[string]interface{}{"id": float64(2), "name": "Ron Stewart"},
				}, r.Value.Enum)
			},
		},
		{
			name: "xml",
			s:    `{ "xml": { "wrapped": true, "name": "foo", "attribute": true, "prefix": "bar", "namespace": "ns1", "x-cdata": true} }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, &Xml{
					Wrapped:   true,
					Name:      "foo",
					Attribute: true,
					Prefix:    "bar",
					Namespace: "ns1",
					CData:     true,
				}, r.Value.Xml)
			},
		},
		{
			name: "format",
			s:    `{ "format": "foo" }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, "foo", r.Value.Format)
			},
		},
		{
			name: "nullable",
			s:    `{ "nullable": true }`,
			fn: func(t *testing.T, r *Ref) {
				require.True(t, r.Value.Nullable)
			},
		},
		{
			name: "pattern",
			s:    `{ "pattern": "[A-Z]{4}" }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, "[A-Z]{4}", r.Value.Pattern)
			},
		},
		{
			name: "minLength",
			s:    `{ "minLength": 3 }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, 3, *r.Value.MinLength)
			},
		},
		{
			name: "maxLength",
			s:    `{ "maxLength": 3 }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, 3, *r.Value.MaxLength)
			},
		},
		{
			name: "minimum",
			s:    `{ "minimum": 3 }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, float64(3), *r.Value.Minimum)
			},
		},
		{
			name: "maximum",
			s:    `{ "maximum": 3 }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, float64(3), *r.Value.Maximum)
			},
		},
		{
			name: "exclusiveMinimum",
			s:    `{ "exclusiveMinimum": true }`,
			fn: func(t *testing.T, r *Ref) {
				require.True(t, *r.Value.ExclusiveMinimum)
			},
		},
		{
			name: "exclusiveMaximum",
			s:    `{ "exclusiveMaximum": true }`,
			fn: func(t *testing.T, r *Ref) {
				require.True(t, *r.Value.ExclusiveMaximum)
			},
		},
		{
			name: "items",
			s:    `{ "items": { "type": "object" } }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, "object", r.Value.Items.Value.Type)
			},
		},
		{
			name: "uniqueItems",
			s:    `{ "uniqueItems": true }`,
			fn: func(t *testing.T, r *Ref) {
				require.True(t, r.Value.UniqueItems)
			},
		},
		{
			name: "minItems",
			s:    `{ "minItems": 3 }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, 3, *r.Value.MinItems)
			},
		},
		{
			name: "maxItems",
			s:    `{ "maxItems": 3 }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, 3, *r.Value.MaxItems)
			},
		},
		{
			name: "properties true",
			s:    `{ "type": "object", "properties": { "name": { "type": "string" } } }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, "object", r.Value.Type)
				require.Equal(t, 1, r.Value.Properties.Len())
				name := r.Value.Properties.Get("name")
				require.NotNil(t, name)
				require.Equal(t, "string", name.Value.Type)
			},
		},
		{
			name: "required",
			s:    `{ "required": ["name"] }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, []string{"name"}, r.Value.Required)
			},
		},
		{
			name: "additional properties true",
			s:    `{ "type": "object", "additionalProperties": true }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, "object", r.Value.Type)
				require.NotNil(t, r.Value.AdditionalProperties)
			},
		},
		{
			name: "additional properties {}",
			s:    `{ "type": "object", "additionalProperties": {} }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, "object", r.Value.Type)
				require.NotNil(t, r.Value.AdditionalProperties)
			},
		},
		{
			name: "minProperties",
			s:    `{ "minProperties": 3 }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, 3, *r.Value.MinProperties)
			},
		},
		{
			name: "maxProperties",
			s:    `{ "maxProperties": 3 }`,
			fn: func(t *testing.T, r *Ref) {
				require.Equal(t, 3, *r.Value.MaxProperties)
			},
		},
	} {
		test := testcase
		t.Run(test.name, func(t *testing.T) {
			r := &Ref{}
			err := json.Unmarshal([]byte(test.s), r)
			require.NoError(t, err)
			test.fn(t, r)
		})
	}
}
