package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
	"testing"
)

func toFloatP(f float64) *float64 { return &f }
func toIntP(i int) *int           { return &i }

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
				require.Equal(t, "[string integer]", result.Type.String())
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
				require.Equal(t, "string", foo.Value.Type.String())
				bar := result.Properties.Get("bar")
				require.NotNil(t, bar)
				require.Equal(t, "number", bar.Value.Type.String())
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
				require.Equal(t, "[string number]", foo.Value.Type.String())
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
				require.Equal(t, "string", result.Items.Value.Type.String())
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
				require.Equal(t, "string", result.Items.Value.Type.String())
				require.Equal(t, "foo", result.Items.Value.Format)
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
			name: "patch example",
			schemas: []*schema.Schema{
				{},
				{Examples: []interface{}{"foo"}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, result.Examples)
			},
		},
		{
			name: "patch overwrite example",
			schemas: []*schema.Schema{
				{Examples: []interface{}{"bar"}},
				{Examples: []interface{}{"foo"}},
			},
			test: func(t *testing.T, result *schema.Schema) {
				require.Equal(t, []interface{}{"foo"}, result.Examples)
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
		schemas []*schema.Ref
		test    func(t *testing.T, result *schema.Ref)
	}{
		{
			name: "patch is nil",
			schemas: []*schema.Ref{
				{Value: schematest.New("object")},
				{},
			},
			test: func(t *testing.T, result *schema.Ref) {
				require.Equal(t, "object", result.Value.Type.String())
			},
		},
		{
			name: "source is nil",
			schemas: []*schema.Ref{
				{},
				{Value: schematest.New("object")},
			},
			test: func(t *testing.T, result *schema.Ref) {
				require.Equal(t, "object", result.Value.Type.String())
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