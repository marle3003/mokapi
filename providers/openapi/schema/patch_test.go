package schema_test

import (
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	jsonSchema "mokapi/schema/json/schema"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchema_PatchSameSchema(t *testing.T) {
	// The same scheme is used for a reference ($ref) and should not create
	// duplicate entries when patching arrays like allOf.
	s := schematest.NewAllOf(schematest.New("string"))
	s.Patch(s)
	require.Len(t, s.AllOf, 1)
}

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
				{ExclusiveMinimum: jsonSchema.NewUnionTypeB[float64, bool](true)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.True(t, result.ExclusiveMinimum.B)
			},
		},
		{
			name: "patch overwrite minimum",
			schemas: []*schema.Schema{
				{ExclusiveMinimum: jsonSchema.NewUnionTypeB[float64, bool](true)},
				{ExclusiveMinimum: jsonSchema.NewUnionTypeB[float64, bool](false)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.False(t, result.ExclusiveMinimum.B)
			},
		},
		{
			name: "patch exclusive maximum",
			schemas: []*schema.Schema{
				{},
				{ExclusiveMaximum: jsonSchema.NewUnionTypeB[float64, bool](true)},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.True(t, result.ExclusiveMaximum.B)
			},
		},
		{
			name: "patch overwrite maximum",
			schemas: []*schema.Schema{
				{ExclusiveMaximum: jsonSchema.NewUnionTypeB[float64, bool](true)},
				{ExclusiveMaximum: jsonSchema.NewUnionTypeB[float64, bool](false)},
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
				schematest.New("", schematest.WithUniqueItems(true)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.True(t, *result.UniqueItems)
			},
		},
		{
			name: "patch overwrite uniqueItems",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithUniqueItems(true)),
				schematest.New("", schematest.WithUniqueItems(false)),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.False(t, *result.UniqueItems)
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
				schematest.New("", schematest.WithAdditionalProperties(&schema.Schema{Boolean: toBoolP(false)})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.NotNil(t, result.AdditionalProperties)
				require.Equal(t, false, *result.AdditionalProperties.Boolean)
			},
		},
		{
			name: "patch overwrite additionalProperties",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithAdditionalProperties(&schema.Schema{Boolean: toBoolP(false)})),
				schematest.New("", schematest.WithAdditionalProperties(&schema.Schema{Boolean: toBoolP(true)})),
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
				require.Equal(t, []jsonSchema.Example{{"foo"}}, result.Examples)
			},
		},
		{
			name: "patch overwrite examples",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithExamples("bar")),
				schematest.New("", schematest.WithExamples("foo")),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []jsonSchema.Example{{"foo"}}, result.Examples)
			},
		},
		{
			name: "patch example",
			schemas: []*schema.Schema{
				{},
				schematest.New("", schematest.WithExample([]interface{}{"foo"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, &jsonSchema.Example{Value: []interface{}{"foo"}}, result.Example)
			},
		},
		{
			name: "patch overwrite example",
			schemas: []*schema.Schema{
				schematest.New("", schematest.WithExample([]interface{}{"bar"})),
				schematest.New("", schematest.WithExample([]interface{}{"foo"})),
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, &jsonSchema.Example{Value: []interface{}{"foo"}}, result.Example)
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

func TestSchema_Patch_Keywords(t *testing.T) {
	str := func() *schema.Schema { return &schema.Schema{Type: jsonSchema.Types{"string"}} }
	num := func() *schema.Schema { return &schema.Schema{Type: jsonSchema.Types{"number"}} }
	two := 2

	t.Run("keywords are taken over", func(t *testing.T) {
		s := &schema.Schema{}
		s.Patch(&schema.Schema{
			Schema: "draft", DynamicAnchor: "a", Not: str(), If: str(), Then: str(), Else: str(),
			UnevaluatedItems: str(), Contains: str(), MaxContains: &two, MinContains: &two,
			UnevaluatedProperties: str(), PropertyNames: str(),
			PatternProperties: map[string]*schema.Schema{"^x": str()},
			DependentRequired: map[string][]string{"a": {"b"}},
			DependentSchemas:  map[string]*schema.Schema{"a": str()},
		})

		require.Equal(t, "draft", s.Schema)
		require.Equal(t, "a", s.DynamicAnchor)
		require.Equal(t, str(), s.Not)
		require.Equal(t, str(), s.If)
		require.Equal(t, str(), s.Then)
		require.Equal(t, str(), s.Else)
		require.Equal(t, str(), s.UnevaluatedItems)
		require.Equal(t, str(), s.Contains)
		require.Equal(t, 2, *s.MaxContains)
		require.Equal(t, 2, *s.MinContains)
		require.Equal(t, str(), s.UnevaluatedProperties)
		require.Equal(t, str(), s.PropertyNames)
		require.Equal(t, str(), s.PatternProperties["^x"])
		require.Equal(t, []string{"b"}, s.DependentRequired["a"])
		require.Equal(t, str(), s.DependentSchemas["a"])
	})

	t.Run("nested keywords are merged", func(t *testing.T) {
		s := &schema.Schema{
			Not:               str(),
			PrefixItems:       []*schema.Schema{str()},
			PatternProperties: map[string]*schema.Schema{"^x": str()},
			DependentRequired: map[string][]string{"a": {"b"}},
		}
		s.Patch(&schema.Schema{
			Not:               num(),
			PrefixItems:       []*schema.Schema{num(), str()},
			PatternProperties: map[string]*schema.Schema{"^x": num(), "^y": str()},
			DependentRequired: map[string][]string{"a": {"c"}},
		})

		require.Equal(t, jsonSchema.Types{"string", "number"}, s.Not.Type)
		require.Len(t, s.PrefixItems, 2)
		require.Equal(t, jsonSchema.Types{"string", "number"}, s.PrefixItems[0].Type)
		require.Equal(t, str(), s.PrefixItems[1])
		require.Equal(t, jsonSchema.Types{"string", "number"}, s.PatternProperties["^x"].Type)
		require.Equal(t, str(), s.PatternProperties["^y"])
		require.Equal(t, []string{"c"}, s.DependentRequired["a"])
	})
}

func TestSchema_Patch_DoesNotShareWithPatch(t *testing.T) {
	t.Run("types", func(t *testing.T) {
		patch := &schema.Schema{Type: jsonSchema.Types{"string"}}
		s := &schema.Schema{}
		s.Patch(patch)
		s.Patch(&schema.Schema{Type: jsonSchema.Types{"number"}})

		require.Equal(t, jsonSchema.Types{"string", "number"}, s.Type)
		require.Equal(t, jsonSchema.Types{"string"}, patch.Type)
	})

	t.Run("merging types does not write into the origin array", func(t *testing.T) {
		origin := make(jsonSchema.Types, 1, 10)
		origin[0] = "string"
		s := &schema.Schema{Type: origin}

		s.Patch(&schema.Schema{Type: jsonSchema.Types{"number"}})

		require.Equal(t, jsonSchema.Types{"string", "number"}, s.Type)
		require.Equal(t, "", origin[:2][1])
	})

	t.Run("nested schemas, compositions and pointers", func(t *testing.T) {
		build := func() *schema.Schema {
			max := 3
			s := schematest.NewAllOf(schematest.New("string"))
			s.Items = schematest.New("integer")
			s.MaxLength = &max
			s.Xml = &schema.Xml{Name: "foo"}
			s.Properties = &schema.Schemas{}
			s.Properties.Set("id", schematest.New("string"))
			return s
		}
		patch := build()
		s := &schema.Schema{}
		s.Patch(patch)

		s.AllOf[0].Type[0] = "changed"
		s.Items.Type[0] = "changed"
		*s.MaxLength = 99
		s.Xml.Name = "changed"
		s.Properties.Get("id").Type[0] = "changed"

		require.Equal(t, build(), patch)
	})

	t.Run("composition entry is added to existing composition", func(t *testing.T) {
		patch := &schema.Schema{AnyOf: []*schema.Schema{{Title: "a", Type: jsonSchema.Types{"string"}}}}
		s := &schema.Schema{AnyOf: []*schema.Schema{{Title: "b"}}}
		s.Patch(patch)

		s.AnyOf[1].Type[0] = "changed"

		require.Equal(t, jsonSchema.Types{"string"}, patch.AnyOf[0].Type)
	})
}

func TestSchema_Patch_CyclicSchema(t *testing.T) {
	cyclic := parseCyclic(t)

	result := &schema.Schema{}
	result.Patch(cyclic)

	// Sub is derived by Parse and not patched, the cycle is reached through the properties
	child := result.Properties.Get("child")
	node := child.Sub
	require.NotNil(t, node)
	require.NotSame(t, cyclic.Sub, node)
	require.Same(t, node, node.Properties.Get("child").Sub, "the copy keeps the cycle")

	result.Title = "changed"
	node.Title = "changed"
	node.Properties.Get("child").Title = "changed"
	node.Type[0] = "changed"
	require.Equal(t, "node", cyclic.Sub.Title)
	require.Equal(t, jsonSchema.Types{"object"}, cyclic.Sub.Type)
	require.Equal(t, "node", cyclic.Sub.Properties.Get("child").Title)
	require.NotEqual(t, "changed", cyclic.Title)
}

// newCyclicNode builds the graph Parse leaves behind for a schema that refers to
// itself (node.properties.child: $ref node): the resolved reference shares the
// properties of node, so node.Properties -> child -> child.Properties == node.Properties.
func newCyclicNode() *schema.Schema {
	props := &schema.Schemas{}
	props.Set("name", schematest.New("string"))
	node := schematest.New("object", schematest.WithTitle("node"))
	node.Properties = props
	props.Set("child", &schema.Schema{Type: jsonSchema.Types{"object"}, Properties: props})
	return node
}

// Two files with the same title that both define the same recursive schema are merged.
// Until this is fixed, Patch recurses forever: fatal error "stack overflow" (not recoverable,
// it aborts the whole test binary).
func TestSchema_Patch_CyclicSchemaIntoCyclicSchema(t *testing.T) {
	s := newCyclicNode()

	s.Patch(newCyclicNode())

	child := s.Properties.Get("child")
	require.NotNil(t, child)
	require.Equal(t, "string", s.Properties.Get("name").Type.String())
	require.Same(t, s.Properties, child.Properties)
}

// The target cycles over one schema, the patch over two, both with the same keys:
// the pairs (node, a) and (node, b) recur. The properties of the longer cycle must
// still reach the target.
func TestSchema_Patch_CyclesOfDifferentLength(t *testing.T) {
	s := newCyclicNode()

	aProps := &schema.Schemas{}
	a := &schema.Schema{Type: jsonSchema.Types{"object"}, Properties: aProps}
	bProps := &schema.Schemas{}
	b := &schema.Schema{Type: jsonSchema.Types{"object"}, Properties: bProps}
	aProps.Set("name", schematest.New("string"))
	aProps.Set("child", b)
	bProps.Set("extra", schematest.New("integer"))
	bProps.Set("child", a)

	s.Patch(a)

	require.Equal(t, "string", s.Properties.Get("name").Type.String())
	require.Equal(t, "integer", s.Properties.Get("extra").Type.String())
	require.NotNil(t, s.Properties.Get("child"))
}

// A schema that is referenced twice is patched once, so an untitled composition
// entry (which is appended, not merged) is not added twice.
func TestSchema_Patch_SharedSchemaIsPatchedOnce(t *testing.T) {
	newRoot := func(shared *schema.Schema) *schema.Schema {
		return schematest.New("object",
			schematest.WithProperty("a", shared),
			schematest.WithProperty("b", shared))
	}
	s := newRoot(schematest.NewAny(schematest.New("string")))

	s.Patch(newRoot(schematest.NewAny(schematest.New("integer"))))

	require.Len(t, s.Properties.Get("a").AnyOf, 2)
	require.Same(t, s.Properties.Get("a"), s.Properties.Get("b"))
}

func TestSchemas_Patch_CyclicSchemaIntoCyclicSchema(t *testing.T) {
	target := newCyclicNode().Properties
	patch := newCyclicNode().Properties

	target.Patch(patch)

	require.Equal(t, "string", target.Get("name").Type.String())
	require.NotNil(t, target.Get("child"))
	require.Same(t, target, target.Get("child").Properties)
}
