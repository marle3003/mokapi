package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func toFloatP(f float64) *float64 { return &f }
func toIntP(i int) *int           { return &i }
func toBoolP(b bool) *bool        { return &b }

func TestSchema_Patch(t *testing.T) {
	testcases := []struct {
		name    string
		schemas []*schema.Schema
		test    func(t *testing.T, result *schema.Schema)
	}{
		{
			name: "patch id",
			schemas: []*schema.Schema{
				{},
				{Id: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Id)
			},
		},
		{
			name: "patch id pattern",
			schemas: []*schema.Schema{
				{Id: "bar"},
				{Id: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Id)
			},
		},
		{
			name: "patch anchor",
			schemas: []*schema.Schema{
				{},
				{Anchor: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Anchor)
			},
		},
		{
			name: "patch id anchor",
			schemas: []*schema.Schema{
				{Anchor: "bar"},
				{Anchor: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Anchor)
			},
		},
		{
			name: "patch boolean",
			schemas: []*schema.Schema{
				{},
				{Boolean: toBoolP(true)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, true, *result.Boolean)
			},
		},
		{
			name: "patch id anchor",
			schemas: []*schema.Schema{
				{Boolean: toBoolP(true)},
				{Boolean: toBoolP(false)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, false, *result.Boolean)
			},
		},
		{
			name: "patch type",
			schemas: []*schema.Schema{
				{},
				{Type: schema.Types{"integer"}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "integer", result.Type.String())
			},
		},
		{
			name: "patch type merge",
			schemas: []*schema.Schema{
				{Type: schema.Types{"string"}},
				{Type: schema.Types{"integer"}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "[string, integer]", result.Type.String())
			},
		},
		{
			name: "patch types result list should be unique",
			schemas: []*schema.Schema{
				{Type: schema.Types{"string"}},
				{Type: schema.Types{"string"}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "string", result.Type.String())
			},
		},
		{
			name: "patch enum",
			schemas: []*schema.Schema{
				{},
				{Enum: []interface{}{"foo"}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, result.Enum)
			},
		},
		{
			name: "patch overwrite enum",
			schemas: []*schema.Schema{
				{Enum: []interface{}{"bar"}},
				{Enum: []interface{}{"foo"}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, result.Enum)
			},
		},
		{
			name: "patch const",
			schemas: []*schema.Schema{
				{},
				schematest.NewTypes(nil, schematest.WithConst([]interface{}{"foo"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, *result.Const)
			},
		},
		{
			name: "patch overwrite const",
			schemas: []*schema.Schema{
				schematest.NewTypes(nil, schematest.WithConst([]interface{}{"bar"})),
				schematest.NewTypes(nil, schematest.WithConst([]interface{}{"foo"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, *result.Const)
			},
		},
		{
			name: "patch format",
			schemas: []*schema.Schema{
				{},
				{Format: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Format)
			},
		},
		{
			name: "patch overwrite format",
			schemas: []*schema.Schema{
				{Format: "bar"},
				{Format: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Format)
			},
		},
		{
			name: "patch pattern",
			schemas: []*schema.Schema{
				{},
				{Pattern: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Pattern)
			},
		},
		{
			name: "patch overwrite pattern",
			schemas: []*schema.Schema{
				{Pattern: "bar"},
				{Pattern: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Pattern)
			},
		},
		{
			name: "patch minLength",
			schemas: []*schema.Schema{
				{},
				{MinLength: toIntP(3)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MinLength)
			},
		},
		{
			name: "patch overwrite minLength",
			schemas: []*schema.Schema{
				{MinLength: toIntP(10)},
				{MinLength: toIntP(3)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MinLength)
			},
		},
		{
			name: "patch maxLength",
			schemas: []*schema.Schema{
				{},
				{MaxLength: toIntP(3)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MaxLength)
			},
		},
		{
			name: "patch overwrite maxLength",
			schemas: []*schema.Schema{
				{MaxLength: toIntP(10)},
				{MaxLength: toIntP(3)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MaxLength)
			},
		},
		{
			name: "patch multipleOf",
			schemas: []*schema.Schema{
				{},
				{MultipleOf: toFloatP(3)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.MultipleOf)
				require.Equal(t, float64(3), *result.MultipleOf)
			},
		},
		{
			name: "patch overwrite multipleOf",
			schemas: []*schema.Schema{
				{MultipleOf: toFloatP(10)},
				{MultipleOf: toFloatP(3)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, float64(3), *result.MultipleOf)
			},
		},
		{
			name: "patch minimum",
			schemas: []*schema.Schema{
				{},
				{Minimum: toFloatP(2)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, float64(2), *result.Minimum)
			},
		},
		{
			name: "patch overwrite minimum",
			schemas: []*schema.Schema{
				{Minimum: toFloatP(2)},
				{Minimum: toFloatP(5)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, float64(5), *result.Minimum)
			},
		},
		{
			name: "patch maximum",
			schemas: []*schema.Schema{
				{},
				{Maximum: toFloatP(2)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, float64(2), *result.Maximum)
			},
		},
		{
			name: "patch overwrite maximum",
			schemas: []*schema.Schema{
				{Maximum: toFloatP(2)},
				{Maximum: toFloatP(5)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, float64(5), *result.Maximum)
			},
		},
		{
			name: "patch exclusive minimum",
			schemas: []*schema.Schema{
				{},
				{ExclusiveMinimum: schema.NewUnionTypeB[float64, bool](true)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.True(t, result.ExclusiveMinimum.B)
			},
		},
		{
			name: "patch overwrite minimum",
			schemas: []*schema.Schema{
				{ExclusiveMinimum: schema.NewUnionTypeB[float64, bool](true)},
				{ExclusiveMinimum: schema.NewUnionTypeB[float64, bool](false)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.False(t, result.ExclusiveMinimum.B)
			},
		},
		{
			name: "patch exclusive maximum",
			schemas: []*schema.Schema{
				{},
				{ExclusiveMaximum: schema.NewUnionTypeB[float64, bool](true)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.True(t, result.ExclusiveMaximum.B)
			},
		},
		{
			name: "patch overwrite maximum",
			schemas: []*schema.Schema{
				{ExclusiveMaximum: schema.NewUnionTypeB[float64, bool](true)},
				{ExclusiveMaximum: schema.NewUnionTypeB[float64, bool](false)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.False(t, result.ExclusiveMaximum.B)
			},
		},
		{
			name: "patch array items",
			schemas: []*schema.Schema{
				{},
				schematest.New("array",
					schematest.WithItems("string")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "array", result.Type.String())
				require.Equal(t, "string", result.Items.Type.String())
			},
		},
		{
			name: "patch array items format",
			schemas: []*schema.Schema{
				schematest.New("array",
					schematest.WithItems("string")),
				schematest.New("array",
					schematest.WithItems("string", schematest.WithFormat("foo"))),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "array", result.Type.String())
				require.Equal(t, "string", result.Items.Type.String())
				require.Equal(t, "foo", result.Items.Format)
			},
		},
		{
			name: "patch exclusive uniqueItems",
			schemas: []*schema.Schema{
				{},
				{UniqueItems: true},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.True(t, result.UniqueItems)
			},
		},
		{
			name: "patch overwrite uniqueItems",
			schemas: []*schema.Schema{
				{UniqueItems: true},
				{UniqueItems: false},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.False(t, result.UniqueItems)
			},
		},
		{
			name: "patch minItems",
			schemas: []*schema.Schema{
				{},
				{MinItems: toIntP(3)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MinItems)
			},
		},
		{
			name: "patch overwrite minItems",
			schemas: []*schema.Schema{
				{MinItems: toIntP(3)},
				{MinItems: toIntP(5)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 5, *result.MinItems)
			},
		},
		{
			name: "patch maxItems",
			schemas: []*schema.Schema{
				{},
				{MaxItems: toIntP(3)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MaxItems)
			},
		},
		{
			name: "patch overwrite maxItems",
			schemas: []*schema.Schema{
				{MaxItems: toIntP(3)},
				{MaxItems: toIntP(5)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 5, *result.MaxItems)
			},
		},
		{
			name: "patch exclusive shuffleItems",
			schemas: []*schema.Schema{
				{},
				{ShuffleItems: true},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.True(t, result.ShuffleItems)
			},
		},
		{
			name: "patch overwrite uniqueItems",
			schemas: []*schema.Schema{
				{ShuffleItems: true},
				{ShuffleItems: false},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.False(t, result.ShuffleItems)
			},
		},
		{
			name: "patch properties",
			schemas: []*schema.Schema{
				{},
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("string"))),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 1, result.Properties.Len())
				require.NotNil(t, result.Properties.Get("foo"))
			},
		},
		{
			name: "patch extend properties",
			schemas: []*schema.Schema{
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("string"))),
				schematest.New("object",
					schematest.WithProperty("bar", schematest.New("number"))),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 2, result.Properties.Len())
				foo := result.Properties.Get("foo")
				require.NotNil(t, foo)
				require.Equal(t, "string", foo.Type.String())
				bar := result.Properties.Get("bar")
				require.NotNil(t, bar)
				require.Equal(t, "number", bar.Type.String())
			},
		},
		{
			name: "patch change property type",
			schemas: []*schema.Schema{
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("string"))),
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("number"))),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 1, result.Properties.Len())
				foo := result.Properties.Get("foo")
				require.NotNil(t, foo)
				require.Equal(t, "[string, number]", foo.Type.String())
			},
		},
		{
			name: "patch required",
			schemas: []*schema.Schema{
				{},
				{Required: []string{"foo"}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []string{"foo"}, result.Required)
			},
		},
		{
			name: "patch overwrite required",
			schemas: []*schema.Schema{
				{Required: []string{"bar"}},
				{Required: []string{"foo"}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []string{"foo"}, result.Required)
			},
		},
		{
			name: "patch additionalProperties",
			schemas: []*schema.Schema{
				{},
				schematest.NewTypes(nil, schematest.WithFreeForm(false)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.AdditionalProperties)
				require.Equal(t, false, result.IsFreeForm())
			},
		},
		{
			name: "patch overwrite additionalProperties",
			schemas: []*schema.Schema{
				schematest.NewTypes(nil, schematest.WithFreeForm(false)),
				schematest.NewTypes(nil, schematest.WithFreeForm(true)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, true, result.AdditionalProperties.IsFreeForm())
			},
		},
		{
			name: "patch overwrite additionalProperties schema",
			schemas: []*schema.Schema{
				schematest.NewTypes(nil, schematest.WithAdditionalProperties(schematest.New("string"))),
				schematest.NewTypes(nil, schematest.WithAdditionalProperties(schematest.New("integer"))),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.AdditionalProperties.Ref)
				require.Equal(t, schema.Types{"string", "integer"}, result.AdditionalProperties.Type)
			},
		},
		{
			name: "patch minProperties",
			schemas: []*schema.Schema{
				{},
				{MinProperties: toIntP(3)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MinProperties)
			},
		},
		{
			name: "patch overwrite minProperties",
			schemas: []*schema.Schema{
				{MinProperties: toIntP(3)},
				{MinProperties: toIntP(5)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 5, *result.MinProperties)
			},
		},
		{
			name: "patch maxProperties",
			schemas: []*schema.Schema{
				{},
				{MaxProperties: toIntP(3)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MaxProperties)
			},
		},
		{
			name: "patch overwrite maxProperties",
			schemas: []*schema.Schema{
				{MaxProperties: toIntP(3)},
				{MaxProperties: toIntP(5)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 5, *result.MaxProperties)
			},
		},
		{
			name: "patch title",
			schemas: []*schema.Schema{
				{},
				{Title: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Title)
			},
		},
		{
			name: "patch overwrite title",
			schemas: []*schema.Schema{
				{Title: "bar"},
				{Title: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Title)
			},
		},
		{
			name: "patch description",
			schemas: []*schema.Schema{
				{},
				{Description: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Description)
			},
		},
		{
			name: "patch overwrite description",
			schemas: []*schema.Schema{
				{Description: "bar"},
				{Description: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Description)
			},
		},
		{
			name: "patch default",
			schemas: []*schema.Schema{
				{},
				{Default: []string{"foo"}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []string{"foo"}, result.Default)
			},
		},
		{
			name: "patch overwrite default",
			schemas: []*schema.Schema{
				{Default: []string{"bar"}},
				{Default: []string{"foo"}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []string{"foo"}, result.Default)
			},
		},
		{
			name: "patch deprecated",
			schemas: []*schema.Schema{
				{},
				{Deprecated: true},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, true, result.Deprecated)
			},
		},
		{
			name: "patch overwrite default",
			schemas: []*schema.Schema{
				{Deprecated: true},
				{Deprecated: false},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, false, result.Deprecated)
			},
		},
		{
			name: "patch examples",
			schemas: []*schema.Schema{
				{},
				{Examples: []interface{}{"foo"}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, result.Examples)
			},
		},
		{
			name: "patch overwrite examples",
			schemas: []*schema.Schema{
				{Examples: []interface{}{"bar"}},
				{Examples: []interface{}{"foo"}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, result.Examples)
			},
		},
		{
			name: "patch contentMediaType",
			schemas: []*schema.Schema{
				{},
				{ContentMediaType: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.ContentMediaType)
			},
		},
		{
			name: "patch overwrite contentMediaType",
			schemas: []*schema.Schema{
				{ContentMediaType: "foo"},
				{ContentMediaType: "bar"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "bar", result.ContentMediaType)
			},
		},
		{
			name: "patch contentEncoding",
			schemas: []*schema.Schema{
				{},
				{ContentEncoding: "foo"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.ContentEncoding)
			},
		},
		{
			name: "patch overwrite contentEncoding",
			schemas: []*schema.Schema{
				{ContentEncoding: "foo"},
				{ContentEncoding: "bar"},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "bar", result.ContentEncoding)
			},
		},
		{
			name: "patch definitions",
			schemas: []*schema.Schema{
				{},
				{Definitions: map[string]*schema.Schema{"foo": schematest.New("string")}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "string", result.Definitions["foo"].Type.String())
			},
		},
		{
			name: "patch overwrite definitions",
			schemas: []*schema.Schema{
				{Definitions: map[string]*schema.Schema{"foo": schematest.New("string")}},
				{Definitions: map[string]*schema.Schema{"foo": schematest.New("integer")}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "[string, integer]", result.Definitions["foo"].Type.String())
			},
		},
		{
			name: "patch $defs",
			schemas: []*schema.Schema{
				{},
				{Defs: map[string]*schema.Schema{"foo": schematest.New("string")}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "string", result.Defs["foo"].Type.String())
			},
		},
		{
			name: "patch overwrite $defs",
			schemas: []*schema.Schema{
				{Defs: map[string]*schema.Schema{"foo": schematest.New("string")}},
				{Defs: map[string]*schema.Schema{"foo": schematest.New("integer")}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "[string, integer]", result.Defs["foo"].Type.String())
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := tc.schemas[0]
			for _, p := range tc.schemas[1:] {
				s.Patch(p)
			}
			tc.test(t, s)
		})
	}
}
