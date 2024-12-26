package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	jsonSchema "mokapi/schema/json/schema"
	"testing"
)

func TestConvert(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Schema
		test func(t *testing.T, s *jsonSchema.Schema)
	}{
		{
			name: "nil schema",
			s:    nil,
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Nil(t, s)
			},
		},
		{
			name: "schema",
			s:    schematest.New("string", schematest.WithSchema("foo")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, "foo", s.Schema)
			},
		},
		{
			name: "type",
			s:    schematest.New("string"),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, jsonSchema.Types{"string"}, s.Type)
			},
		},
		{
			name: "types",
			s:    schematest.NewTypes([]string{"string", "integer"}),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, jsonSchema.Types{"string", "integer"}, s.Type)
			},
		},
		{
			name: "enum",
			s:    schematest.New("string", schematest.WithEnum([]interface{}{"foo", "bar"})),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.Enum)
				require.Equal(t, []interface{}{"foo", "bar"}, s.Enum)
			},
		},
		{
			name: "const",
			s:    schematest.New("string", schematest.WithConst("foo")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.Const)
				require.Equal(t, "foo", *s.Const)
			},
		},
		{
			name: "multipleOf",
			s:    schematest.New("integer", schematest.WithMultipleOf(12)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.MultipleOf)
				require.Equal(t, float64(12), *s.MultipleOf)
			},
		},
		{
			name: "minimum",
			s:    schematest.New("integer", schematest.WithMinimum(12)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.Minimum)
				require.Equal(t, float64(12), *s.Minimum)
			},
		},
		{
			name: "maximum",
			s:    schematest.New("integer", schematest.WithMaximum(12)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.Maximum)
				require.Equal(t, float64(12), *s.Maximum)
			},
		},
		{
			name: "exclusiveMinimum",
			s:    schematest.New("integer", schematest.WithExclusiveMinimum(12)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.ExclusiveMinimum)
				require.Equal(t, float64(12), s.ExclusiveMinimum.A)
			},
		},
		{
			name: "exclusiveMaximum",
			s:    schematest.New("integer", schematest.WithExclusiveMaximum(12)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.ExclusiveMaximum)
				require.Equal(t, float64(12), s.ExclusiveMaximum.A)
			},
		},
		{
			name: "minLength",
			s:    schematest.New("string", schematest.WithMinLength(12)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.MinLength)
				require.Equal(t, 12, *s.MinLength)
			},
		},
		{
			name: "maxLength",
			s:    schematest.New("string", schematest.WithMaxLength(12)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.MaxLength)
				require.Equal(t, 12, *s.MaxLength)
			},
		},
		{
			name: "pattern",
			s:    schematest.New("string", schematest.WithPattern("foo")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, "foo", s.Pattern)
			},
		},
		{
			name: "format",
			s:    schematest.New("string", schematest.WithFormat("foo")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, "foo", s.Format)
			},
		},
		{
			name: "items",
			s:    schematest.New("array", schematest.WithItems("string")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.Items)
				require.Equal(t, jsonSchema.Types{"string"}, s.Items.Value.Type)
			},
		},
		{
			name: "prefixItems",
			s:    schematest.New("array", schematest.WithPrefixItems(schematest.New("string"), schematest.New("integer"))),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.PrefixItems)
				require.Equal(t, "string", s.PrefixItems[0].Value.Type.String())
				require.Equal(t, "integer", s.PrefixItems[1].Value.Type.String())
			},
		},
		{
			name: "unevaluatedItems",
			s:    schematest.New("array", schematest.WithUnevaluatedItems(schematest.NewRef("string"))),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.UnevaluatedItems)
				require.Equal(t, "string", s.UnevaluatedItems.Value.Type.String())
			},
		},
		{
			name: "contains",
			s:    schematest.New("array", schematest.WithContains(schematest.NewRef("string"))),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, "string", s.Contains.Value.Type.String())
			},
		},
		{
			name: "maxContains",
			s:    schematest.New("array", schematest.WithMaxContains(3)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, 3, *s.MaxContains)
			},
		},
		{
			name: "minContains",
			s:    schematest.New("array", schematest.WithMinContains(2)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, 2, *s.MinContains)
			},
		},
		{
			name: "UniqueItems",
			s:    schematest.New("array", schematest.WithUniqueItems()),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, true, s.UniqueItems)
			},
		},
		{
			name: "minItems",
			s:    schematest.New("array", schematest.WithMinItems(12)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.MinItems)
				require.Equal(t, 12, *s.MinItems)
			},
		},
		{
			name: "maxItems",
			s:    schematest.New("array", schematest.WithMaxItems(12)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.MaxItems)
				require.Equal(t, 12, *s.MaxItems)
			},
		},
		{
			name: "shuffleItems",
			s:    schematest.New("array", schematest.WithShuffleItems()),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, true, s.ShuffleItems)
			},
		},
		{
			name: "properties",
			s:    schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.Properties)
				foo := s.Properties.Get("foo")
				require.NotNil(t, foo)
				require.Equal(t, jsonSchema.Types{"string"}, foo.Value.Type)
			},
		},
		{
			name: "patternProperties",
			s:    schematest.New("object", schematest.WithPatternProperty("[a-z]*", schematest.New("string"))),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Contains(t, s.PatternProperties, "[a-z]*")
				require.Equal(t, "string", s.PatternProperties["[a-z]*"].Value.Type.String())
			},
		},
		{
			name: "required",
			s:    schematest.New("object", schematest.WithRequired("foo")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.Required)
				require.Equal(t, []string{"foo"}, s.Required)
			},
		},
		{
			name: "dependentRequired",
			s:    schematest.New("object", schematest.WithDependentRequired("foo", "bar", "yuh")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Len(t, s.DependentRequired, 1)
				require.Equal(t, []string{"bar", "yuh"}, s.DependentRequired["foo"])
			},
		},
		{
			name: "dependentSchemas",
			s:    schematest.New("object", schematest.WithDependentSchemas("foo", schematest.New("string"))),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Len(t, s.DependentSchemas, 1)
				require.Equal(t, "string", s.DependentSchemas["foo"].Value.Type.String())
			},
		},
		{
			name: "additionalProperties",
			s:    schematest.New("object", schematest.WithAdditionalProperties(schematest.New("string"))),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.AdditionalProperties.Value)
				require.Equal(t, jsonSchema.Types{"string"}, s.AdditionalProperties.Value.Type)
			},
		},
		{
			name: "unevaluatedProperties",
			s:    schematest.New("object", schematest.WithUnevaluatedProperties(schematest.NewRef("string"))),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, "string", s.UnevaluatedProperties.Value.Type.String())
			},
		},
		{
			name: "unevaluatedProperties false",
			s:    schematest.New("object", schematest.WithUnevaluatedProperties(&schema.Ref{Boolean: toBoolP(false)})),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, false, *s.UnevaluatedProperties.Boolean)
			},
		},
		{
			name: "additionalProperties forbidden",
			s:    schematest.New("object", schematest.WithFreeForm(false)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Nil(t, s.AdditionalProperties.Value)
				require.Equal(t, false, *s.AdditionalProperties.Boolean)
			},
		},
		{
			name: "additionalProperties free-form",
			s:    schematest.New("object", schematest.WithFreeForm(true)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Nil(t, s.AdditionalProperties.Value)
				require.Equal(t, true, *s.AdditionalProperties.Boolean)
			},
		},
		{
			name: "propertyNames",
			s:    schematest.New("object", schematest.WithPropertyNames(schematest.New("string"))),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, "string", s.PropertyNames.Value.Type.String())
			},
		},
		{
			name: "minProperties",
			s:    schematest.New("object", schematest.WithMinProperties(12)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.MinProperties)
				require.Equal(t, 12, *s.MinProperties)
			},
		},
		{
			name: "maxProperties",
			s:    schematest.New("object", schematest.WithMaxProperties(12)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.MaxProperties)
				require.Equal(t, 12, *s.MaxProperties)
			},
		},
		{
			name: "anyOf",
			s:    schematest.NewAny(schematest.New("string"), schematest.New("integer")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.AnyOf)
				require.Equal(t, jsonSchema.Types{"string"}, s.AnyOf[0].Value.Type)
				require.Equal(t, jsonSchema.Types{"integer"}, s.AnyOf[1].Value.Type)
			},
		},
		{
			name: "allOf",
			s:    schematest.NewAllOf(schematest.New("string"), schematest.New("integer")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.AllOf)
				require.Equal(t, jsonSchema.Types{"string"}, s.AllOf[0].Value.Type)
				require.Equal(t, jsonSchema.Types{"integer"}, s.AllOf[1].Value.Type)
			},
		},
		{
			name: "oneOf",
			s:    schematest.NewOneOf(schematest.New("string"), schematest.New("integer")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.OneOf)
				require.Equal(t, jsonSchema.Types{"string"}, s.OneOf[0].Value.Type)
				require.Equal(t, jsonSchema.Types{"integer"}, s.OneOf[1].Value.Type)
			},
		},
		{
			name: "not",
			s:    schematest.NewTypes(nil, schematest.WithNot(schematest.NewRef("integer"))),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.Not)
				require.Equal(t, jsonSchema.Types{"integer"}, s.Not.Value.Type)
			},
		},
		{
			name: "if",
			s:    schematest.NewTypes(nil, schematest.WithIf(schematest.NewRef("integer"))),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.If)
				require.Equal(t, jsonSchema.Types{"integer"}, s.If.Value.Type)
			},
		},
		{
			name: "then",
			s:    schematest.NewTypes(nil, schematest.WithThen(schematest.NewRef("integer"))),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.Then)
				require.Equal(t, jsonSchema.Types{"integer"}, s.Then.Value.Type)
			},
		},
		{
			name: "else",
			s:    schematest.NewTypes(nil, schematest.WithElse(schematest.NewRef("integer"))),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.NotNil(t, s.Else)
				require.Equal(t, jsonSchema.Types{"integer"}, s.Else.Value.Type)
			},
		},
		{
			name: "title",
			s:    schematest.New("object", schematest.WithTitle("foo")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, "foo", s.Title)
			},
		},
		{
			name: "description",
			s:    schematest.New("object", schematest.WithDescription("foo")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, "foo", s.Description)
			},
		},
		{
			name: "default",
			s:    schematest.New("object", schematest.WithDefault("foo")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, "foo", s.Default)
			},
		},
		{
			name: "deprecated",
			s:    schematest.New("object", schematest.WithDeprecated(true)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, true, s.Deprecated)
			},
		},
		{
			name: "examples",
			s:    schematest.New("object", schematest.WithExample(true)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, []interface{}{true}, s.Examples)
			},
		},
		{
			name: "examples",
			s:    schematest.New("object", schematest.WithExamples(true)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, []interface{}{true}, s.Examples)
			},
		},
		{
			name: "contentMediaType",
			s:    schematest.New("object", schematest.WithContentMediaType("foo")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, "foo", s.ContentMediaType)
			},
		},
		{
			name: "contentEncoding",
			s:    schematest.New("object", schematest.WithContentEncoding("foo")),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, "foo", s.ContentEncoding)
			},
		},
		{
			name: "nullable",
			s:    schematest.New("string", schematest.IsNullable(true)),
			test: func(t *testing.T, s *jsonSchema.Schema) {
				require.Equal(t, jsonSchema.Types{"string", "null"}, s.Type)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := schema.ConvertToJsonSchema(&schema.Ref{Value: tc.s})
			tc.test(t, r.Value)
		})
	}
}

func TestConvertToJsonSchema_Ref(t *testing.T) {
	testcases := []struct {
		name string
		r    *schema.Ref
		test func(t *testing.T, r *jsonSchema.Ref)
	}{
		{
			name: "nil",
			r:    nil,
			test: func(t *testing.T, r *jsonSchema.Ref) {
				require.Nil(t, r)
			},
		},
		{
			name: "bool true",
			r:    &schema.Ref{Boolean: toBoolP(true)},
			test: func(t *testing.T, r *jsonSchema.Ref) {
				require.Equal(t, true, *r.Boolean)
			},
		},
		{
			name: "bool false",
			r:    &schema.Ref{Boolean: toBoolP(false)},
			test: func(t *testing.T, r *jsonSchema.Ref) {
				require.Equal(t, false, *r.Boolean)
			},
		},
		{
			name: "schema",
			r:    &schema.Ref{Value: schematest.New("string")},
			test: func(t *testing.T, r *jsonSchema.Ref) {
				require.Nil(t, r.Boolean)
				require.NotNil(t, r.Value)
				require.Equal(t, "string", r.Value.Type.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := schema.ConvertToJsonSchema(tc.r)
			tc.test(t, r)
		})
	}
}
