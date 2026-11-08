package schema_test

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// parse unmarshals the schema and resolves its references like a loaded config.
func parse(t *testing.T, data string) *schema.Schema {
	t.Helper()
	s := &schema.Schema{}
	require.NoError(t, yaml.Unmarshal([]byte(data), &s))
	require.NoError(t, s.Parse(&dynamic.Config{Data: s}, &dynamictest.Reader{}))
	return s
}

func TestSchema_Clone(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "base",
			test: func(t *testing.T) {
				s := &schema.Schema{
					Id:        "id",
					Reference: dynamic.Reference[*schema.Schema]{Ref: "ref", DynamicRef: "dynamicRef"},
					Schema:    "schema",
					Boolean:   toBoolP(true),
					Anchor:    "anchor",
					Type:      schema.Types{"object"},
					Enum:      []any{"one", "two", "three"},
					Const: func() *any {
						var v any
						v = "const"
						return &v
					}(),
					ContentEncoding:  "utf-8",
					ContentMediaType: "text/plain",
				}
				s2 := s.Clone()

				require.Equal(t, "id", s2.Id)
				require.Equal(t, "ref", s2.Ref)
				require.Equal(t, "dynamicRef", s2.DynamicRef)
				require.Equal(t, "schema", s2.Schema)
				require.Equal(t, true, *s2.Boolean)
				require.Equal(t, "anchor", s2.Anchor)
				require.Equal(t, schema.Types{"object"}, s2.Type)
				require.Equal(t, []any{"one", "two", "three"}, s2.Enum)
				require.Equal(t, "const", *s2.Const)
				require.Equal(t, "utf-8", s2.ContentEncoding)
				require.Equal(t, "text/plain", s2.ContentMediaType)

				s2.Id = "id2"
				s2.Ref = "ref2"
				s2.DynamicRef = "dynamicRef2"
				s2.Schema = "schema2"
				s2.Boolean = toBoolP(false)
				s2.Anchor = "anchor"
				s2.Type[0] = "boolean"
				s2.Enum[0] = "foo"
				s2.Const = func() *any {
					var v any
					v = "const2"
					return &v
				}()
				s2.ContentEncoding = "foo"
				s2.ContentMediaType = "application/json"

				require.Equal(t, "id", s.Id)
				require.Equal(t, "ref", s.Ref)
				require.Equal(t, "dynamicRef", s.DynamicRef)
				require.Equal(t, "schema", s.Schema)
				require.Equal(t, true, *s.Boolean)
				require.Equal(t, "anchor", s.Anchor)
				require.Equal(t, schema.Types{"object"}, s.Type)
				require.Equal(t, []any{"one", "two", "three"}, s.Enum)
				require.Equal(t, "const", *s.Const)
				require.Equal(t, "utf-8", s.ContentEncoding)
				require.Equal(t, "text/plain", s.ContentMediaType)
			},
		},
		{
			name: "number",
			test: func(t *testing.T) {
				s := &schema.Schema{
					MultipleOf:       toFloatP(12),
					Maximum:          toFloatP(12),
					Minimum:          toFloatP(12),
					ExclusiveMaximum: schema.NewUnionTypeA[float64, bool](12),
					ExclusiveMinimum: schema.NewUnionTypeA[float64, bool](12),
				}
				s2 := s.Clone()

				require.Equal(t, 12.0, *s2.MultipleOf)
				require.Equal(t, 12.0, *s2.Maximum)
				require.Equal(t, 12.0, *s2.Minimum)
				require.Equal(t, 12.0, s2.ExclusiveMaximum.A)
				require.Equal(t, 12.0, s2.ExclusiveMinimum.A)

				s2.MultipleOf = toFloatP(1)
				s2.Maximum = toFloatP(1)
				s2.Minimum = toFloatP(1)
				s2.ExclusiveMaximum.A = 1
				s2.ExclusiveMinimum.A = 1

				require.Equal(t, 12.0, *s.MultipleOf)
				require.Equal(t, 12.0, *s.Maximum)
				require.Equal(t, 12.0, *s.Minimum)
				require.Equal(t, 12.0, s.ExclusiveMaximum.A)
				require.Equal(t, 12.0, s.ExclusiveMinimum.A)
			},
		},
		{
			name: "string",
			test: func(t *testing.T) {
				s := &schema.Schema{
					MaxLength: toIntP(12),
					MinLength: toIntP(12),
					Pattern:   "pattern",
					Format:    "format",
				}
				s2 := s.Clone()

				require.Equal(t, 12, *s.MaxLength)
				require.Equal(t, 12, *s.MinLength)
				require.Equal(t, "pattern", s.Pattern)
				require.Equal(t, "format", s.Format)

				s2.MaxLength = toIntP(1)
				s2.MinLength = toIntP(1)
				s2.Pattern = "foo"
				s2.Format = "bar"

				require.Equal(t, 12, *s.MaxLength)
				require.Equal(t, 12, *s.MinLength)
				require.Equal(t, "pattern", s.Pattern)
				require.Equal(t, "format", s.Format)
			},
		},
		{
			name: "array",
			test: func(t *testing.T) {
				s := &schema.Schema{
					Items:            schematest.New("string"),
					PrefixItems:      []*schema.Schema{schematest.New("string")},
					UnevaluatedItems: schematest.New("string"),
					Contains:         schematest.New("string"),
					MaxContains:      toIntP(12),
					MinContains:      toIntP(12),
					MaxItems:         toIntP(12),
					MinItems:         toIntP(12),
					UniqueItems:      toBoolP(true),
					ShuffleItems:     true,
				}
				s2 := s.Clone()

				require.Equal(t, schematest.New("string"), s2.Items)
				require.Equal(t, schematest.New("string"), s2.PrefixItems[0])
				require.Equal(t, schematest.New("string"), s2.UnevaluatedItems)
				require.Equal(t, schematest.New("string"), s2.Contains)
				require.Equal(t, 12, *s2.MaxContains)
				require.Equal(t, 12, *s2.MinContains)
				require.Equal(t, 12, *s2.MaxItems)
				require.Equal(t, 12, *s2.MinItems)
				require.Equal(t, true, *s2.UniqueItems)
				require.Equal(t, true, s2.ShuffleItems)

				s2.Items.Type[0] = "integer"
				s2.PrefixItems[0].Type[0] = "integer"
				s2.UnevaluatedItems.Type[0] = "integer"
				s2.Contains.Type[0] = "integer"
				s2.MaxContains = toIntP(1)
				s2.MinContains = toIntP(1)
				s2.MaxItems = toIntP(1)
				s2.MinItems = toIntP(1)
				s2.UniqueItems = toBoolP(false)
				s2.ShuffleItems = false

				require.Equal(t, schematest.New("string"), s.Items)
				require.Equal(t, schematest.New("string"), s.PrefixItems[0])
				require.Equal(t, schematest.New("string"), s.UnevaluatedItems)
				require.Equal(t, schematest.New("string"), s.Contains)
				require.Equal(t, 12, *s.MaxContains)
				require.Equal(t, 12, *s.MinContains)
				require.Equal(t, 12, *s.MaxItems)
				require.Equal(t, 12, *s.MinItems)
				require.Equal(t, true, *s.UniqueItems)
				require.Equal(t, true, s.ShuffleItems)
			},
		},
		{
			name: "object",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithProperty("foo", schematest.New("string")),
					schematest.WithPatternProperty("foo", schematest.New("string")),
					schematest.WithMaxProperties(12),
					schematest.WithMinProperties(12),
					schematest.WithRequired("foo"),
					schematest.WithDependentRequired("foo", "bar"),
					schematest.WithDependentSchemas("foo", schematest.New("string")),
					schematest.WithAdditionalProperties(schematest.New("string")),
					schematest.WithUnevaluatedProperties(schematest.New("string")),
					schematest.WithPropertyNames(schematest.New("string")),
				)
				s2 := s.Clone()

				require.Equal(t, schematest.New("string"), s2.Properties.Get("foo"))
				require.Equal(t, schematest.New("string"), s2.PatternProperties["foo"])
				require.Equal(t, 12, *s2.MaxProperties)
				require.Equal(t, 12, *s2.MinProperties)
				require.Equal(t, []string{"foo"}, s2.Required)
				require.Equal(t, []string{"bar"}, s2.DependentRequired["foo"])
				require.Equal(t, schematest.New("string"), s2.DependentSchemas["foo"])
				require.Equal(t, schematest.New("string"), s2.AdditionalProperties)
				require.Equal(t, schematest.New("string"), s2.UnevaluatedProperties)
				require.Equal(t, schematest.New("string"), s2.PropertyNames)

				s2.Properties.Get("foo").Type[0] = "integer"
				s2.PatternProperties["foo"].Type[0] = "integer"
				s2.MaxProperties = toIntP(1)
				s2.MinProperties = toIntP(1)
				s2.Required[0] = "yuh"
				s2.DependentRequired["foo"][0] = "yuh"
				s2.DependentSchemas["foo"].Type[0] = "integer"
				s2.AdditionalProperties.Type[0] = "integer"
				s2.UnevaluatedProperties.Type[0] = "integer"
				s2.PropertyNames.Type[0] = "integer"

				require.Equal(t, schematest.New("string"), s.Properties.Get("foo"))
				require.Equal(t, schematest.New("string"), s.PatternProperties["foo"])
				require.Equal(t, 12, *s.MaxProperties)
				require.Equal(t, 12, *s.MinProperties)
				require.Equal(t, []string{"foo"}, s.Required)
				require.Equal(t, []string{"bar"}, s.DependentRequired["foo"])
				require.Equal(t, schematest.New("string"), s.DependentSchemas["foo"])
				require.Equal(t, schematest.New("string"), s.AdditionalProperties)
				require.Equal(t, schematest.New("string"), s.UnevaluatedProperties)
				require.Equal(t, schematest.New("string"), s.PropertyNames)
			},
		},
		{
			name: "conditional",
			test: func(t *testing.T) {
				s := schematest.NewTypes(nil,
					schematest.WithIf(schematest.New("string")),
					schematest.WithThen(schematest.New("string")),
					schematest.WithElse(schematest.New("string")),
					schematest.Any(schematest.New("string")),
					schematest.WithOneOf(schematest.New("string")),
					schematest.WithAllOf(schematest.New("string")),
				)
				s2 := s.Clone()

				require.Equal(t, schematest.New("string"), s2.If)
				require.Equal(t, schematest.New("string"), s2.Then)
				require.Equal(t, schematest.New("string"), s2.Else)
				require.Equal(t, schematest.New("string"), s2.AnyOf[0])
				require.Equal(t, schematest.New("string"), s2.OneOf[0])
				require.Equal(t, schematest.New("string"), s2.AllOf[0])

				s2.If.Type[0] = "integer"
				s2.Then.Type[0] = "integer"
				s2.Else.Type[0] = "integer"
				s2.AnyOf[0].Type[0] = "integer"
				s2.OneOf[0].Type[0] = "integer"
				s2.AllOf[0].Type[0] = "integer"

				require.Equal(t, schematest.New("string"), s.If)
				require.Equal(t, schematest.New("string"), s.Then)
				require.Equal(t, schematest.New("string"), s.Else)
				require.Equal(t, schematest.New("string"), s.AnyOf[0])
				require.Equal(t, schematest.New("string"), s.OneOf[0])
				require.Equal(t, schematest.New("string"), s.AllOf[0])
			},
		},
		{
			name: "definitions and defs",
			test: func(t *testing.T) {
				s := &schema.Schema{
					Definitions: map[string]*schema.Schema{"foo": schematest.New("string")},
					Defs:        map[string]*schema.Schema{"bar": schematest.New("string")},
				}
				s2 := s.Clone()

				require.Equal(t, schematest.New("string"), s2.Definitions["foo"])
				require.Equal(t, schematest.New("string"), s2.Defs["bar"])

				s2.Definitions["foo"].Type[0] = "integer"
				s2.Definitions["baz"] = schematest.New("string")
				s2.Defs["bar"].Type[0] = "integer"
				s2.Defs["baz"] = schematest.New("string")

				require.Equal(t, schematest.New("string"), s.Definitions["foo"])
				require.Len(t, s.Definitions, 1)
				require.Equal(t, schematest.New("string"), s.Defs["bar"])
				require.Len(t, s.Defs, 1)
			},
		},
		{
			name: "exclusive limits keep their variant",
			test: func(t *testing.T) {
				s := &schema.Schema{
					ExclusiveMaximum: schema.NewUnionTypeB[float64, bool](true),
					ExclusiveMinimum: schema.NewUnionTypeA[float64, bool](1),
				}
				s2 := s.Clone()

				require.False(t, s2.ExclusiveMaximum.IsA())
				require.Equal(t, true, s2.ExclusiveMaximum.B)
				require.True(t, s2.ExclusiveMinimum.IsA())
				require.Equal(t, 1.0, s2.ExclusiveMinimum.A)
			},
		},
		{
			name: "annotations",
			test: func(t *testing.T) {
				s := &schema.Schema{
					Title:       "title",
					Description: "description",
					Deprecated:  true,
					Default:     "default",
					Examples:    []schema.Example{{Value: "foo"}},
				}
				s2 := s.Clone()

				require.Equal(t, "title", s2.Title)
				require.Equal(t, "description", s2.Description)
				require.True(t, s2.Deprecated)
				require.Equal(t, "default", s2.Default)
				require.Equal(t, []schema.Example{{Value: "foo"}}, s2.Examples)

				s2.Examples[0] = schema.Example{Value: "bar"}
				require.Equal(t, []schema.Example{{Value: "foo"}}, s.Examples)
			},
		},
		{
			name: "resolved reference",
			test: func(t *testing.T) {
				s := parse(t, `
$defs:
  foo:
    type: string
properties:
  a:
    $ref: '#/$defs/foo'
`)
				a := s.Properties.Get("a")
				require.NotNil(t, a.Sub)

				s2 := s.Clone()

				a2 := s2.Properties.Get("a")
				require.Equal(t, "#/$defs/foo", a2.Ref)
				require.NotNil(t, a2.Sub)
				require.NotSame(t, a.Sub, a2.Sub)
				require.Equal(t, a.Sub.Type, a2.Sub.Type)

				a2.Sub.Type[0] = "integer"
				require.Equal(t, schema.Types{"string"}, a.Sub.Type)
			},
		},
		{
			name: "cyclic reference",
			test: func(t *testing.T) {
				s := parse(t, `
type: object
properties:
  child:
    $ref: '#'
`)
				child := s.Properties.Get("child")
				require.Same(t, s, child.Sub)

				s2 := s.Clone()

				child2 := s2.Properties.Get("child")
				require.NotSame(t, child, child2)
				require.Same(t, s2, child2.Sub, "the copy has the same cycle and does not point into the original")
			},
		},
		{
			name: "keywords next to reference are kept when the copy is parsed",
			test: func(t *testing.T) {
				s := &schema.Schema{}
				err := yaml.Unmarshal([]byte(`
$defs:
  foo:
    type: string
    description: from definition
properties:
  a:
    $ref: '#/$defs/foo'
    description: from reference
`), &s)
				require.NoError(t, err)

				s2 := s.Clone()

				err = s2.Parse(&dynamic.Config{Data: s2}, &dynamictest.Reader{})
				require.NoError(t, err)
				a := s2.Properties.Get("a")
				require.Equal(t, schema.Types{"string"}, a.Type)
				require.Equal(t, "from reference", a.Description)
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
