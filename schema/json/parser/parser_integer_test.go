package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestParse_Integer(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Schema
		d    interface{}
		test func(t *testing.T, v interface{}, err error)

		skipValidationFormat  bool
		convertStringToNumber bool
	}{
		{
			name: "integer but string",
			s:    schematest.New("integer"),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected integer but got string\nschema path #/type")
			},
		},
		{
			name: "integer but map",
			s:    schematest.New("integer"),
			d:    map[string]interface{}{},
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected integer but got object\nschema path #/type")
			},
		},
		{
			name: "int32 but greater as max value",
			s:    schematest.New("integer", schematest.WithFormat("int32")),
			d:    int64(1e10),
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninteger '10000000000' does not match format 'int32'\nschema path #/format")
			},
		},
		{
			name: "int32 but lower as max value",
			s:    schematest.New("integer", schematest.WithFormat("int32")),
			d:    int64(-1e10),
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninteger '-10000000000' does not match format 'int32'\nschema path #/format")
			},
		},
		{
			name: "int32 and greater as max value but disabled",
			s:    schematest.New("integer", schematest.WithFormat("int32")),
			d:    int64(1e10),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(1e10), v)
			},

			skipValidationFormat: true,
		},
		{
			name: "int",
			s:    schematest.New("integer"),
			d:    1234,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(1234), v)
			},
		},
		{
			name: "int32",
			s:    schematest.New("integer"),
			d:    int32(1234),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(1234), v)
			},
		},
		{
			name: "float but with fraction",
			s:    schematest.New("integer"),
			d:    3.4,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected integer but got number\nschema path #/type")
			},
		},
		{
			name: "float",
			s:    schematest.New("integer"),
			d:    3.0,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(3), v)
			},
		},
		{
			name: "convert from string",
			s:    schematest.New("integer"),
			d:    "1234",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(1234), v)
			},

			convertStringToNumber: true,
		},
		{
			name: "convert from string error",
			s:    schematest.New("integer"),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected integer but got string\nschema path #/type")
			},

			convertStringToNumber: true,
		},
		{
			name: "convert from string to int32 error",
			s:    schematest.New("integer", schematest.WithFormat("int32")),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected integer but got string\nschema path #/type")
			},

			convertStringToNumber: true,
		},
		{
			name: "convert from string to int32",
			s:    schematest.New("integer", schematest.WithFormat("int32")),
			d:    "1234",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(1234), v)
			},

			convertStringToNumber: true,
		},
		{
			name: "integer multipleOf error",
			s:    schematest.New("integer", schematest.WithMultipleOf(3)),
			d:    8,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninteger 8 is not a multiple of 3\nschema path #/multipleOf")
			},
		},
		{
			name: "integer multipleOf",
			s:    schematest.New("integer", schematest.WithMultipleOf(3)),
			d:    12,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), v)
			},
		},
		{
			name: "integer minimum error",
			s:    schematest.New("integer", schematest.WithMinimum(3)),
			d:    1,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninteger 1 is less than minimum value of 3\nschema path #/minimum")
			},
		},
		{
			name: "integer minimum",
			s:    schematest.New("integer", schematest.WithMultipleOf(3)),
			d:    3,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(3), v)
			},
		},
		{
			name: "integer maximum error",
			s:    schematest.New("integer", schematest.WithMaximum(3)),
			d:    4,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninteger 4 exceeds maximum value of 3\nschema path #/maximum")
			},
		},
		{
			name: "integer maximum",
			s:    schematest.New("integer", schematest.WithMaximum(3)),
			d:    3,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(3), v)
			},
		},
		{
			name: "integer exclusive minimum less error",
			s:    schematest.New("integer", schematest.WithExclusiveMinimum(3)),
			d:    2,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninteger 2 is less than minimum value of 3\nschema path #/exclusiveMinimum")
			},
		},
		{
			name: "integer exclusive minimum error",
			s:    schematest.New("integer", schematest.WithExclusiveMinimum(3)),
			d:    3,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninteger 3 equals minimum value of 3\nschema path #/exclusiveMinimum")
			},
		},
		{
			name: "integer exclusive minimum",
			s:    schematest.New("integer", schematest.WithExclusiveMinimum(3)),
			d:    4,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(4), v)
			},
		},
		{
			name: "integer exclusive minimum and minimum error",
			s:    schematest.New("integer", schematest.WithExclusiveMinimumFlag(true), schematest.WithMinimum(3)),
			d:    3,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninteger 3 equals minimum value of 3 and exclusive minimum is true\nschema path #/minimum")
			},
		},
		{
			name: "integer exclusive minimum and minimum",
			s:    schematest.New("integer", schematest.WithExclusiveMinimumFlag(true), schematest.WithMinimum(3)),
			d:    4,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(4), v)
			},
		},
		{
			name: "integer exclusive maximum exceeds error",
			s:    schematest.New("integer", schematest.WithExclusiveMaximum(3)),
			d:    4,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninteger 4 exceeds maximum value of 3\nschema path #/exclusiveMaximum")
			},
		},
		{
			name: "integer exclusive maximum error",
			s:    schematest.New("integer", schematest.WithExclusiveMaximum(3)),
			d:    3,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninteger 3 equals maximum value of 3\nschema path #/exclusiveMaximum")
			},
		},
		{
			name: "integer exclusive maximum",
			s:    schematest.New("integer", schematest.WithExclusiveMaximum(3)),
			d:    2,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(2), v)
			},
		},
		{
			name: "integer exclusive maximum and maximum error",
			s:    schematest.New("integer", schematest.WithExclusiveMaximumFlag(true), schematest.WithMaximum(3)),
			d:    3,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninteger 3 equals maximum value of 3 and exclusive maximum is true\nschema path #/maximum")
			},
		},
		{
			name: "integer exclusive maximum and maximum",
			s:    schematest.New("integer", schematest.WithExclusiveMaximumFlag(true), schematest.WithMaximum(3)),
			d:    2,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(2), v)
			},
		},
		{
			name: "const error",
			s:    schematest.New("integer", schematest.WithConst(10)),
			d:    3,
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\nvalue '3' does not match const '10'\nschema path #/const")
			},
		},
		{
			name: "const",
			s:    schematest.New("integer", schematest.WithConst(10)),
			d:    10,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(10), v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parser.Parser{
				Schema:                      tc.s,
				SkipValidationFormatKeyword: tc.skipValidationFormat,
				ConvertStringToNumber:       tc.convertStringToNumber,
			}

			v, err := p.Parse(tc.d)
			tc.test(t, v, err)
		})
	}
}
