package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestParse_Number(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Schema
		d    interface{}
		test func(t *testing.T, v interface{}, err error)

		skipValidationFormat  bool
		convertStringToNumber bool
	}{
		{
			name: "number but string",
			s:    schematest.New("number"),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected number but got string\nschema path #/type")
			},
		},
		{
			name: "number but map",
			s:    schematest.New("number"),
			d:    map[string]interface{}{},
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected number but got object\nschema path #/type")
			},
		},
		{
			name: "float but double",
			s:    schematest.New("number", schematest.WithFormat("float")),
			d:    -3942.2,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\nnumber '-3942.2' does not match format 'float'\nschema path #/format")
			},
		},
		{
			name: "float but double but disabled",
			s:    schematest.New("number", schematest.WithFormat("float")),
			d:    -3942.2,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, -3942.2, v)
			},

			skipValidationFormat: true,
		},
		{
			name: "double",
			s:    schematest.New("number"),
			d:    1234567890.5,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 1234567890.5, v)
			},
		},
		{
			name: "float",
			s:    schematest.New("number"),
			d:    1234.5,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 1234.5, v)
			},
		},
		{
			name: "int",
			s:    schematest.New("number"),
			d:    1234567890,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(1234567890), v)
			},
		},
		{
			name: "int64",
			s:    schematest.New("number"),
			d:    int64(1234567890),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(1234567890), v)
			},
		},
		{
			name: "convert from string",
			s:    schematest.New("number"),
			d:    "1234.5",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 1234.5, v)
			},

			convertStringToNumber: true,
		},
		{
			name: "convert from string error",
			s:    schematest.New("number"),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected number but got string\nschema path #/type")
			},

			convertStringToNumber: true,
		},
		{
			name: "convert from string to float error",
			s:    schematest.New("number", schematest.WithFormat("float")),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\ninvalid type, expected number but got string\nschema path #/type")
			},

			convertStringToNumber: true,
		},
		{
			name: "convert from string to float",
			s:    schematest.New("number", schematest.WithFormat("float")),
			d:    "1234.5",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 1234.5, v)
			},

			convertStringToNumber: true,
		},
		{
			name: "number multipleOf error",
			s:    schematest.New("number", schematest.WithMultipleOf(3.5)),
			d:    9.5,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\nnumber 9.5 is not a multiple of 3.5\nschema path #/multipleOf")
			},
		},
		{
			name: "number multipleOf",
			s:    schematest.New("number", schematest.WithMultipleOf(3.5)),
			d:    10.5,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 10.5, v)
			},
		},
		{
			name: "number minimum error",
			s:    schematest.New("number", schematest.WithMinimum(3.5)),
			d:    1,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\nnumber 1 is less than minimum value of 3.5\nschema path #/minimum")
			},
		},
		{
			name: "number minimum",
			s:    schematest.New("number", schematest.WithMultipleOf(3.5)),
			d:    3.5,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 3.5, v)
			},
		},
		{
			name: "number maximum error",
			s:    schematest.New("number", schematest.WithMaximum(3.5)),
			d:    4,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\nnumber 4 exceeds maximum value of 3.5\nschema path #/maximum")
			},
		},
		{
			name: "number maximum",
			s:    schematest.New("number", schematest.WithMaximum(3.5)),
			d:    3.5,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 3.5, v)
			},
		},
		{
			name: "number exclusive minimum less error",
			s:    schematest.New("number", schematest.WithExclusiveMinimum(3.5)),
			d:    3.0,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\nnumber 3 is less than minimum value of 3.5\nschema path #/exclusiveMinimum")
			},
		},
		{
			name: "number exclusive minimum error",
			s:    schematest.New("number", schematest.WithExclusiveMinimum(3.5)),
			d:    3.5,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\nnumber 3.5 equals minimum value of 3.5\nschema path #/exclusiveMinimum")
			},
		},
		{
			name: "number exclusive minimum",
			s:    schematest.New("number", schematest.WithExclusiveMinimum(3)),
			d:    4,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 4.0, v)
			},
		},
		{
			name: "number exclusive minimum and minimum error",
			s:    schematest.New("number", schematest.WithExclusiveMinimumFlag(true), schematest.WithMinimum(3)),
			d:    3,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\nnumber 3 equals minimum value of 3 and exclusive minimum is true\nschema path #/minimum")
			},
		},
		{
			name: "number exclusive minimum and minimum",
			s:    schematest.New("number", schematest.WithExclusiveMinimumFlag(true), schematest.WithMinimum(3)),
			d:    4,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 4.0, v)
			},
		},
		{
			name: "number exclusive maximum exceeds error",
			s:    schematest.New("number", schematest.WithExclusiveMaximum(3)),
			d:    4,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\nnumber 4 exceeds maximum value of 3\nschema path #/exclusiveMaximum")
			},
		},
		{
			name: "number exclusive maximum error",
			s:    schematest.New("number", schematest.WithExclusiveMaximum(3)),
			d:    3,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\nnumber 3 equals maximum value of 3\nschema path #/exclusiveMaximum")
			},
		},
		{
			name: "number exclusive maximum",
			s:    schematest.New("number", schematest.WithExclusiveMaximum(3)),
			d:    2,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 2.0, v)
			},
		},
		{
			name: "number exclusive maximum and maximum error",
			s:    schematest.New("number", schematest.WithExclusiveMaximumFlag(true), schematest.WithMaximum(3)),
			d:    3,
			test: func(t *testing.T, v interface{}, err error) {
				require.IsType(t, &parser.Error{}, err)
				require.EqualError(t, err, "found 1 error:\nnumber 3 equals maximum value of 3 and exclusive maximum is true\nschema path #/maximum")
			},
		},
		{
			name: "number exclusive maximum and maximum",
			s:    schematest.New("number", schematest.WithExclusiveMaximumFlag(true), schematest.WithMaximum(3)),
			d:    2,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 2.0, v)
			},
		},
		{
			name: "const error",
			s:    schematest.New("number", schematest.WithConst(10.5)),
			d:    3,
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "found 1 error:\nvalue '3' does not match const '10.5'\nschema path #/const")
			},
		},
		{
			name: "const",
			s:    schematest.New("number", schematest.WithConst(10.5)),
			d:    10.5,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 10.5, v)
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
