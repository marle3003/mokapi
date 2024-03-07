package schema

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSchemaJson(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, s *Schema, err error)
	}{
		{
			name: "single type",
			data: `{"type": "string"}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"string"}, s.Type)
			},
		},
		{
			name: "two types",
			data: `{"type": ["string", "integer"] }`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"string", "integer"}, s.Type)
			},
		},
		{
			name: "type is not a string value",
			data: `{"type": ["string", 123] }`,
			test: func(t *testing.T, s *Schema, err error) {
				require.EqualError(t, err, "cannot unmarshal 123 into field type of type schema")
			},
		},
		{
			name: "one enum value",
			data: `{"enum": ["foo"]}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo"}, s.Enum)
			},
		},
		{
			name: "two enum values",
			data: `{"enum": ["foo", 123] }`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", float64(123)}, s.Enum)
			},
		},
		{
			name: "const value",
			data: `{"const": "foo"}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", s.Const)
			},
		},
		/*
		 * Numbers
		 */
		{
			name: "multipleOf",
			data: `{"multipleOf": 12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MultipleOf)
			},
		},
		{
			name: "multipleOf is not integer",
			data: `{"multipleOf": 12.5}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.EqualError(t, err, "cannot unmarshal 12.5 into field multipleOf of type schema")
			},
		},
		{
			name: "maximum",
			data: `{"maximum": 12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(12), *s.Maximum)
			},
		},
		{
			name: "exclusiveMaximum",
			data: `{"exclusiveMaximum": 12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(12), *s.ExclusiveMaximum)
			},
		},
		{
			name: "minimum",
			data: `{"minimum": 12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(12), *s.Minimum)
			},
		},
		{
			name: "exclusiveMinimum",
			data: `{"exclusiveMinimum": 12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(12), *s.ExclusiveMinimum)
			},
		},
		/*
		 * Strings
		 */
		{
			name: "maxLength",
			data: `{"maxLength": 12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MaxLength)
			},
		},
		{
			name: "maxLength negative",
			data: `{"maxLength": -12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.EqualError(t, err, "maxLength must be a non-negative integer: -12")
			},
		},
		{
			name: "minLength",
			data: `{"minLength": 12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MinLength)
			},
		},
		{
			name: "minLength negative",
			data: `{"minLength": -12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.EqualError(t, err, "minLength must be a non-negative integer: -12")
			},
		},
		{
			name: "pattern",
			data: `{"pattern": "[a-z]"}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "[a-z]", s.Pattern)
			},
		},
		{
			name: "format",
			data: `{"format": "date"}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, "date", s.Format)
			},
		},
		/*
		 * Arrays
		 */
		{
			name: "maxItems",
			data: `{"maxItems": 12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MaxItems)
			},
		},
		{
			name: "maxItems negative",
			data: `{"maxItems": -12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.EqualError(t, err, "maxItems must be a non-negative integer: -12")
			},
		},
		{
			name: "minItems",
			data: `{"minItems": 12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MinItems)
			},
		},
		{
			name: "minItems negative",
			data: `{"minItems": -12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.EqualError(t, err, "minItems must be a non-negative integer: -12")
			},
		},
		{
			name: "uniqueItems",
			data: `{"uniqueItems": true}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, true, s.UniqueItems)
			},
		},
		{
			name: "maxContains",
			data: `{"maxContains": 12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, s.MaxContains)
			},
		},
		{
			name: "maxContains negative",
			data: `{"maxContains": -12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.EqualError(t, err, "maxContains must be a non-negative integer: -12")
			},
		},
		{
			name: "minContains",
			data: `{"minContains": 12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, s.MinContains)
			},
		},
		{
			name: "minContains negative",
			data: `{"minContains": -12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.EqualError(t, err, "minContains must be a non-negative integer: -12")
			},
		},
		/*
		 * Objects
		 */
		{
			name: "maxProperties",
			data: `{"maxProperties": 12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MaxProperties)
			},
		},
		{
			name: "maxProperties negative",
			data: `{"maxProperties": -12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.EqualError(t, err, "maxProperties must be a non-negative integer: -12")
			},
		},
		{
			name: "minProperties",
			data: `{"minProperties": 12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, *s.MinProperties)
			},
		},
		{
			name: "minProperties negative",
			data: `{"minProperties": -12}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.EqualError(t, err, "minProperties must be a non-negative integer: -12")
			},
		},
		{
			name: "required",
			data: `{"required": ["foo", "bar"]}`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, s.Required)
			},
		},
		{
			name: "dependentRequired",
			data: `{"dependentRequired": {"foo": ["bar"]} }`,
			test: func(t *testing.T, s *Schema, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string][]string{"foo": {"bar"}}, s.DependentRequired)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var s *Schema
			err := json.Unmarshal([]byte(tc.data), &s)
			tc.test(t, s, err)
		})
	}
}
