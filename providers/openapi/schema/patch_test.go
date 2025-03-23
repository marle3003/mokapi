package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	jsonSchema "mokapi/schema/json/schema"
	"testing"
)

func TestSchema_Patch(t *testing.T) {
	testcases := []struct {
		name    string
		schemas []*schema.Schema
		test    func(t *testing.T, result *schema.Schema)
	}{
		{
			name: "patch type",
			schemas: []*schema.Schema{
				{},
				schematest.New("integer"),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "integer", result.Type.String())
			},
		},
		{
			name: "patch type merge",
			schemas: []*schema.Schema{
				schematest.New("string"),
				schematest.New("integer"),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "[string, integer]", result.Type.String())
			},
		},
		{
			name: "patch types result list should be unique",
			schemas: []*schema.Schema{
				schematest.New("string"),
				schematest.New("string"),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "string", result.Type.String())
			},
		},
		{
			name: "patch enum",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithEnumValues("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, result.Enum)
			},
		},
		{
			name: "patch overwrite enum",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithEnumValues("bar")),
				schematest.New("", schematest.WithEnumValues("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, result.Enum)
			},
		},
		{
			name: "patch const",
			schemas: []*schema.Schema{
				{},
				schematest.New("array", schematest.WithConst([]interface{}{"foo"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, *result.Const)
			},
		},
		{
			name: "patch overwrite const",
			schemas: []*schema.Schema{
				schematest.New("array", schematest.WithConst([]interface{}{"bar"})),
				schematest.New("array", schematest.WithConst([]interface{}{"foo"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, *result.Const)
			},
		},
		{
			name: "patch xml",
			schemas: []*schema.Schema{
				{},
				schematest.New("string",
					schematest.WithXml(&schema.Xml{Name: "foo"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.Xml)
				require.Equal(t, "foo", result.Xml.Name)
			},
		},
		{
			name: "patch xml overwrite name",
			schemas: []*schema.Schema{
				schematest.New("string",
					schematest.WithXml(&schema.Xml{Name: "foo"})),
				schematest.New("string",
					schematest.WithXml(&schema.Xml{Name: "bar"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.Xml)
				require.Equal(t, "bar", result.Xml.Name)
			},
		},
		{
			name: "patch xml overwrite prefix",
			schemas: []*schema.Schema{
				schematest.New("string",
					schematest.WithXml(&schema.Xml{Prefix: "foo"})),
				schematest.New("string",
					schematest.WithXml(&schema.Xml{Prefix: "bar"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.Xml)
				require.Equal(t, "bar", result.Xml.Prefix)
			},
		},
		{
			name: "patch xml overwrite namespace",
			schemas: []*schema.Schema{
				schematest.New("string",
					schematest.WithXml(&schema.Xml{Namespace: "foo"})),
				schematest.New("string",
					schematest.WithXml(&schema.Xml{Namespace: "bar"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.Xml)
				require.Equal(t, "bar", result.Xml.Namespace)
			},
		},
		{
			name: "patch xml overwrite wrapped",
			schemas: []*schema.Schema{
				schematest.New("string",
					schematest.WithXml(&schema.Xml{Wrapped: false})),
				schematest.New("string",
					schematest.WithXml(&schema.Xml{Wrapped: true})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.Xml)
				require.True(t, result.Xml.Wrapped)
			},
		},
		{
			name: "patch xml overwrite attribute",
			schemas: []*schema.Schema{
				schematest.New("string",
					schematest.WithXml(&schema.Xml{Attribute: false})),
				schematest.New("string",
					schematest.WithXml(&schema.Xml{Attribute: true})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.Xml)
				require.True(t, result.Xml.Attribute)
			},
		},
		{
			name: "patch format",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithFormat("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Format)
			},
		},
		{
			name: "patch overwrite format",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithFormat("bar")),
				schematest.New("", schematest.WithFormat("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Format)
			},
		},
		{
			name: "patch nullable",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.IsNullable(true)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, true, result.Nullable)
			},
		},
		{
			name: "patch overwrite format",
			schemas: []*schema.Schema{
				schematest.New("", schematest.IsNullable(true)),
				schematest.New("", schematest.IsNullable(false)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, false, result.Nullable)
			},
		},
		{
			name: "patch pattern",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithPattern("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Pattern)
			},
		},
		{
			name: "patch overwrite pattern",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithPattern("bar")),
				schematest.New("", schematest.WithPattern("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Pattern)
			},
		},
		{
			name: "patch minLength",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithMinLength(3)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MinLength)
			},
		},
		{
			name: "patch overwrite minLength",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithMinLength(10)),
				schematest.New("", schematest.WithMinLength(3)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MinLength)
			},
		},
		{
			name: "patch maxLength",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithMaxLength(3)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MaxLength)
			},
		},
		{
			name: "patch overwrite maxLength",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithMaxLength(10)),
				schematest.New("", schematest.WithMaxLength(3)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MaxLength)
			},
		},
		{
			name: "patch multipleOf",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithMultipleOf(3)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.MultipleOf)
				require.Equal(t, float64(3), *result.MultipleOf)
			},
		},
		{
			name: "patch overwrite multipleOf",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithMultipleOf(10)),
				schematest.New("", schematest.WithMultipleOf(3)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, float64(3), *result.MultipleOf)
			},
		},
		{
			name: "patch minimum",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithMinimum(2)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, float64(2), *result.Minimum)
			},
		},
		{
			name: "patch overwrite minimum",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithMinimum(2)),
				schematest.New("", schematest.WithMinimum(5)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, float64(5), *result.Minimum)
			},
		},
		{
			name: "patch maximum",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithMaximum(2)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, float64(2), *result.Maximum)
			},
		},
		{
			name: "patch overwrite maximum",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithMaximum(2)),
				schematest.New("", schematest.WithMaximum(5)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, float64(5), *result.Maximum)
			},
		},
		{
			name: "patch exclusive minimum",
			schemas: []*schema.Schema{
				{},
				{SubSchema: &schema.SubSchema{ExclusiveMinimum: jsonSchema.NewUnionTypeB[float64, bool](true)}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.True(t, result.ExclusiveMinimum.B)
			},
		},
		{
			name: "patch overwrite minimum",
			schemas: []*schema.Schema{
				{SubSchema: &schema.SubSchema{ExclusiveMinimum: jsonSchema.NewUnionTypeB[float64, bool](true)}},
				{SubSchema: &schema.SubSchema{ExclusiveMinimum: jsonSchema.NewUnionTypeB[float64, bool](false)}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.False(t, result.ExclusiveMinimum.B)
			},
		},
		{
			name: "patch exclusive maximum",
			schemas: []*schema.Schema{
				{},
				{SubSchema: &schema.SubSchema{ExclusiveMaximum: jsonSchema.NewUnionTypeB[float64, bool](true)}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.True(t, result.ExclusiveMaximum.B)
			},
		},
		{
			name: "patch overwrite maximum",
			schemas: []*schema.Schema{
				{SubSchema: &schema.SubSchema{ExclusiveMaximum: jsonSchema.NewUnionTypeB[float64, bool](true)}},
				{SubSchema: &schema.SubSchema{ExclusiveMaximum: jsonSchema.NewUnionTypeB[float64, bool](false)}},
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
				schematest.New("", schematest.WithUniqueItems()),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.True(t, result.UniqueItems)
			},
		},
		{
			name: "patch overwrite uniqueItems",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithUniqueItems()),
				schematest.New(""),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.False(t, result.UniqueItems)
			},
		},
		{
			name: "patch minItems",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithMinItems(3)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MinItems)
			},
		},
		{
			name: "patch overwrite minItems",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithMinItems(3)),
				schematest.New("", schematest.WithMinItems(5)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 5, *result.MinItems)
			},
		},
		{
			name: "patch maxItems",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithMaxItems(3)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MaxItems)
			},
		},
		{
			name: "patch overwrite maxItems",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithMaxItems(3)),
				schematest.New("", schematest.WithMaxItems(5)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 5, *result.MaxItems)
			},
		},
		{
			name: "patch exclusive shuffleItems",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithShuffleItems()),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.True(t, result.ShuffleItems)
			},
		},
		{
			name: "patch overwrite uniqueItems",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithShuffleItems()),
				schematest.New(""),
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
				schematest.New("", schematest.WithRequired("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []string{"foo"}, result.Required)
			},
		},
		{
			name: "patch overwrite required",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithRequired("bar")),
				schematest.New("", schematest.WithRequired("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []string{"foo"}, result.Required)
			},
		},
		{
			name: "patch additionalProperties",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithAdditionalProperties(&schema.Schema{SubSchema: &schema.SubSchema{Boolean: toBoolP(false)}})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.AdditionalProperties)
				require.Equal(t, false, *result.AdditionalProperties.Boolean)
			},
		},
		{
			name: "patch overwrite additionalProperties",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithAdditionalProperties(&schema.Schema{SubSchema: &schema.SubSchema{Boolean: toBoolP(false)}})),
				schematest.New("", schematest.WithAdditionalProperties(&schema.Schema{SubSchema: &schema.SubSchema{Boolean: toBoolP(true)}})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, true, *result.AdditionalProperties.Boolean)
			},
		},
		{
			name: "patch overwrite additionalProperties schema",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithAdditionalProperties(schematest.New("string"))),
				schematest.New("", schematest.WithAdditionalProperties(schematest.New("integer"))),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.AdditionalProperties.Ref)
				require.Equal(t, jsonSchema.Types{"string", "integer"}, result.AdditionalProperties.Type)
			},
		},
		{
			name: "patch minProperties",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithMinProperties(3)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MinProperties)
			},
		},
		{
			name: "patch overwrite minProperties",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithMinProperties(3)),
				schematest.New("", schematest.WithMinProperties(5)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 5, *result.MinProperties)
			},
		},
		{
			name: "patch maxProperties",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithMaxProperties(3)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 3, *result.MaxProperties)
			},
		},
		{
			name: "patch overwrite maxProperties",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithMaxProperties(3)),
				schematest.New("", schematest.WithMaxProperties(5)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, 5, *result.MaxProperties)
			},
		},
		{
			name: "patch title",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithTitle("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Title)
			},
		},
		{
			name: "patch overwrite title",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithTitle("bar")),
				schematest.New("", schematest.WithTitle("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Title)
			},
		},
		{
			name: "patch description",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithDescription("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Description)
			},
		},
		{
			name: "patch overwrite description",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithDescription("bar")),
				schematest.New("", schematest.WithDescription("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.Description)
			},
		},
		{
			name: "patch default",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithDefault([]string{"foo"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []string{"foo"}, result.Default)
			},
		},
		{
			name: "patch overwrite default",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithDefault([]string{"bar"})),
				schematest.New("", schematest.WithDefault([]string{"foo"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []string{"foo"}, result.Default)
			},
		},
		{
			name: "patch deprecated",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithDeprecated(true)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, true, result.Deprecated)
			},
		},
		{
			name: "patch overwrite default",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithDeprecated(true)),
				schematest.New(""),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, false, result.Deprecated)
			},
		},
		{
			name: "patch examples",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithExamples("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, result.Examples)
			},
		},
		{
			name: "patch overwrite examples",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithExamples("bar")),
				schematest.New("", schematest.WithExamples("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, result.Examples)
			},
		},
		{
			name: "patch example",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithExample([]interface{}{"foo"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, result.Example)
			},
		},
		{
			name: "patch overwrite example",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithExample([]interface{}{"bar"})),
				schematest.New("", schematest.WithExample([]interface{}{"foo"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, result.Example)
			},
		},
		{
			name: "patch contentMediaType",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithContentMediaType("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.ContentMediaType)
			},
		},
		{
			name: "patch overwrite contentMediaType",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithContentMediaType("foo")),
				schematest.New("", schematest.WithContentMediaType("bar")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "bar", result.ContentMediaType)
			},
		},
		{
			name: "patch contentEncoding",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithContentEncoding("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "foo", result.ContentEncoding)
			},
		},
		{
			name: "patch overwrite contentEncoding",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithContentEncoding("foo")),
				schematest.New("", schematest.WithContentEncoding("bar")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "bar", result.ContentEncoding)
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

func TestRef_Patch(t *testing.T) {
	testcases := []struct {
		name    string
		schemas []*schema.Schema
		test    func(t *testing.T, result *schema.Schema)
	}{
		{
			name: "patch is nil",
			schemas: []*schema.Schema{
				schematest.New("object"),
				{},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "object", result.Type.String())
			},
		},
		{
			name: "source is nil",
			schemas: []*schema.Schema{
				{},
				schematest.New("object"),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, "object", result.Type.String())
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

func TestPatch_Composition(t *testing.T) {
	testcases := []struct {
		name    string
		schemas []*schema.Schema
		test    func(t *testing.T, result *schema.Schema)
	}{
		{
			name: "add anyOf",
			schemas: []*schema.Schema{
				schematest.NewAny(),
				schematest.NewAny(schematest.New("string")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.AnyOf)
				require.Len(t, result.AnyOf, 1)
				require.Equal(t, "string", result.AnyOf[0].Type.String())
			},
		},
		{
			name: "append anyOf",
			schemas: []*schema.Schema{
				schematest.NewAny(schematest.New("string")),
				schematest.NewAny(schematest.New("string")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.AnyOf)
				require.Len(t, result.AnyOf, 2)
				require.Equal(t, "string", result.AnyOf[0].Type.String())
				require.Equal(t, "string", result.AnyOf[1].Type.String())
			},
		},
		{
			name: "patch anyOf",
			schemas: []*schema.Schema{
				schematest.NewAny(schematest.New("string", schematest.WithTitle("foo"))),
				schematest.NewAny(schematest.New("integer", schematest.WithTitle("foo"))),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.AnyOf)
				require.Len(t, result.AnyOf, 1)
				require.Equal(t, jsonSchema.Types{"string", "integer"}, result.AnyOf[0].Type)
			},
		},
		{
			name: "add allOf",
			schemas: []*schema.Schema{
				schematest.NewAny(),
				schematest.NewAllOf(schematest.New("string")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.AllOf)
				require.Len(t, result.AllOf, 1)
				require.Equal(t, "string", result.AllOf[0].Type.String())
			},
		},
		{
			name: "append allOf",
			schemas: []*schema.Schema{
				schematest.NewAllOf(schematest.New("string")),
				schematest.NewAllOf(schematest.New("string")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.AllOf)
				require.Len(t, result.AllOf, 2)
				require.Equal(t, "string", result.AllOf[0].Type.String())
				require.Equal(t, "string", result.AllOf[1].Type.String())
			},
		},
		{
			name: "patch allOf",
			schemas: []*schema.Schema{
				schematest.NewAllOf(schematest.New("string", schematest.WithTitle("foo"))),
				schematest.NewAllOf(schematest.New("integer", schematest.WithTitle("foo"))),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.AllOf)
				require.Len(t, result.AllOf, 1)
				require.Equal(t, jsonSchema.Types{"string", "integer"}, result.AllOf[0].Type)
			},
		},
		{
			name: "add OneOf",
			schemas: []*schema.Schema{
				schematest.NewAny(),
				schematest.NewOneOf(schematest.New("string")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.OneOf)
				require.Len(t, result.OneOf, 1)
				require.Equal(t, "string", result.OneOf[0].Type.String())
			},
		},
		{
			name: "append OneOf",
			schemas: []*schema.Schema{
				schematest.NewOneOf(schematest.New("string")),
				schematest.NewOneOf(schematest.New("string")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.OneOf)
				require.Len(t, result.OneOf, 2)
				require.Equal(t, "string", result.OneOf[0].Type.String())
				require.Equal(t, "string", result.OneOf[1].Type.String())
			},
		},
		{
			name: "patch OneOf",
			schemas: []*schema.Schema{
				schematest.NewOneOf(schematest.New("string", schematest.WithTitle("foo"))),
				schematest.NewOneOf(schematest.New("integer", schematest.WithTitle("foo"))),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.OneOf)
				require.Len(t, result.OneOf, 1)
				require.Equal(t, jsonSchema.Types{"string", "integer"}, result.OneOf[0].Type)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var s *schema.Schema
			for _, p := range tc.schemas {
				if s == nil {
					s = p
				} else {
					s.Patch(p)
				}
			}
			tc.test(t, s)
		})
	}
}
