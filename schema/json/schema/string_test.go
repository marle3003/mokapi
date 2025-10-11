package schema_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestSchema_String(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Schema
		exp  string
	}{
		{
			name: "empty",
			s:    &schema.Schema{},
			exp:  "empty schema",
		},
		{
			name: "schema always fails validation",
			s:    &schema.Schema{Boolean: toBoolP(false)},
			exp:  "schema (always invalid)",
		},
		{
			name: "schema always valid",
			s:    &schema.Schema{Boolean: toBoolP(true)},
			exp:  "schema (always valid)",
		},
		{
			name: "any of",
			s:    schematest.NewAny(schematest.New("string"), schematest.New("integer")),
			exp:  "any of (schema type=string, schema type=integer)",
		},
		{
			name: "all of",
			s:    schematest.NewAllOf(schematest.New("string"), schematest.New("integer")),
			exp:  "all of (schema type=string, schema type=integer)",
		},
		{
			name: "one of",
			s:    schematest.NewOneOf(schematest.New("string"), schematest.New("integer")),
			exp:  "one of (schema type=string, schema type=integer)",
		},
		{
			name: "not",
			s:    schematest.NewTypes(nil, schematest.WithNot(schematest.New("string"))),
			exp:  "schema not (schema type=string)",
		},
		{
			name: "type string",
			s:    schematest.New("string"),
			exp:  "schema type=string",
		},
		{
			name: "format",
			s:    schematest.New("string", schematest.WithFormat("date-time")),
			exp:  "schema type=string format=date-time",
		},
		{
			name: "pattern",
			s:    schematest.New("string", schematest.WithPattern("abc")),
			exp:  "schema type=string pattern=abc",
		},
		{
			name: "minLength",
			s:    schematest.New("string", schematest.WithMinLength(12)),
			exp:  "schema type=string minLength=12",
		},
		{
			name: "maxLength",
			s:    schematest.New("string", schematest.WithMaxLength(12)),
			exp:  "schema type=string maxLength=12",
		},
		{
			name: "minimum",
			s:    schematest.New("string", schematest.WithMinimum(12)),
			exp:  "schema type=string minimum=12",
		},
		{
			name: "maximum",
			s:    schematest.New("string", schematest.WithMaximum(12)),
			exp:  "schema type=string maximum=12",
		},
		{
			name: "multipleOf",
			s:    schematest.New("string", schematest.WithMultipleOf(12)),
			exp:  "schema type=string multipleOf=12",
		},
		{
			name: "exclusiveMinimum value",
			s:    schematest.New("string", schematest.WithExclusiveMinimum(12)),
			exp:  "schema type=string exclusiveMinimum=12",
		},
		{
			name: "exclusiveMinimum bool",
			s:    schematest.New("string", schematest.WithExclusiveMinimumFlag(true)),
			exp:  "schema type=string exclusiveMinimum=true",
		},
		{
			name: "exclusiveMaximum value",
			s:    schematest.New("string", schematest.WithExclusiveMaximum(12)),
			exp:  "schema type=string exclusiveMaximum=12",
		},
		{
			name: "exclusiveMaximum bool",
			s:    schematest.New("string", schematest.WithExclusiveMaximumFlag(true)),
			exp:  "schema type=string exclusiveMaximum=true",
		},
		{
			name: "minItems",
			s:    schematest.New("string", schematest.WithMinItems(12)),
			exp:  "schema type=string minItems=12",
		},
		{
			name: "maxItems",
			s:    schematest.New("string", schematest.WithMaxItems(12)),
			exp:  "schema type=string maxItems=12",
		},
		{
			name: "minProperties",
			s:    schematest.New("string", schematest.WithMinProperties(12)),
			exp:  "schema type=string minProperties=12",
		},
		{
			name: "maxProperties",
			s:    schematest.New("string", schematest.WithMaxProperties(12)),
			exp:  "schema type=string maxProperties=12",
		},
		{
			name: "uniqueItems",
			s:    schematest.New("string", schematest.WithUniqueItems()),
			exp:  "schema type=string unique-items",
		},
		{
			name: "properties",
			s: schematest.New("object",
				schematest.WithProperty("foo", schematest.New("string")),
				schematest.WithProperty("bar", schematest.New("integer")),
			),
			exp: "schema type=object properties=[foo, bar]",
		},
		{
			name: "items",
			s: schematest.New("array",
				schematest.WithItems("string"),
			),
			exp: "schema type=array items=(schema type=string)",
		},
		{
			name: "required",
			s:    schematest.New("string", schematest.WithRequired("foo", "bar")),
			exp:  "schema type=string required=[foo bar]",
		},
		{
			name: "not free form",
			s:    schematest.New("object", schematest.WithFreeForm(false)),
			exp:  "schema type=object free-form=false",
		},
		{
			name: "title",
			s:    &schema.Schema{Title: "foo"},
			exp:  "schema title=foo",
		},
		{
			name: "description",
			s:    &schema.Schema{Description: "foo"},
			exp:  "schema description=foo",
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.exp, tc.s.String())
		})
	}
}
