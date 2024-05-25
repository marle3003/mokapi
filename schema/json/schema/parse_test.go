package schema

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func toFloatP(f float64) *float64 {
	return &f
}

func toIntP(i int) *int {
	return &i
}

func TestSchema_Parse(t *testing.T) {
	testcases := []struct {
		name string
		s    *Schema
		test func(t *testing.T, err error)
	}{
		{
			name: "multipleOf negative",
			s:    &Schema{MultipleOf: toFloatP(-1)},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "multipleOf must be greater than 0: -1")
			},
		},
		{
			name: "maxLength negative",
			s:    &Schema{MaxLength: toIntP(-1)},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "maxLength must be a non-negative integer: -1")
			},
		},
		{
			name: "minLength negative",
			s:    &Schema{MinLength: toIntP(-1)},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "minLength must be a non-negative integer: -1")
			},
		},
		{
			name: "minLength greater maxLength",
			s:    &Schema{MinLength: toIntP(3), MaxLength: toIntP(1)},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "minLength cannot be greater than maxLength: 3, 1")
			},
		},
		{
			name: "maxItems negative",
			s:    &Schema{MaxItems: toIntP(-1)},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "maxItems must be a non-negative integer: -1")
			},
		},
		{
			name: "minItems negative",
			s:    &Schema{MinItems: toIntP(-1)},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "minItems must be a non-negative integer: -1")
			},
		},
		{
			name: "minItems greater maxItems",
			s:    &Schema{MinItems: toIntP(3), MaxItems: toIntP(1)},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "minItems cannot be greater than maxItems: 3, 1")
			},
		},
		{
			name: "maxContains negative",
			s:    &Schema{MaxContains: toIntP(-1)},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "maxContains must be a non-negative integer: -1")
			},
		},
		{
			name: "minContains negative",
			s:    &Schema{MinContains: toIntP(-1)},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "minContains must be a non-negative integer: -1")
			},
		},
		{
			name: "maxProperties negative",
			s:    &Schema{MaxProperties: toIntP(-1)},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "maxProperties must be a non-negative integer: -1")
			},
		},
		{
			name: "minProperties negative",
			s:    &Schema{MinProperties: toIntP(-1)},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "minProperties must be a non-negative integer: -1")
			},
		},
		{
			name: "minProperties greater maxProperties",
			s:    &Schema{MinProperties: toIntP(3), MaxProperties: toIntP(1)},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "minProperties cannot be greater than maxProperties: 3, 1")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.s.Parse()
			tc.test(t, err)
		})
	}
}
