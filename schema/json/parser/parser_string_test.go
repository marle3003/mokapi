package parser_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/parser"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"mokapi/version"
	"testing"
)

func TestParse_String(t *testing.T) {
	testcases := []struct {
		name string
		s    *schema.Schema
		d    interface{}
		test func(t *testing.T, v interface{}, err error)

		skipValidationFormat bool
	}{
		{
			name: "string",
			s:    schematest.New("string"),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "version",
			s:    schematest.New("string"),
			d:    version.New("1.0.0"),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "1.0.0", v)
			},
		},
		{
			name: "null",
			s:    schematest.NewTypes([]string{"string", "null"}),
			d:    nil,
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, nil, v)
			},
		},
		{
			name: "not string",
			s:    schematest.New("string"),
			d:    12,
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/type: invalid type, expected string but got integer")
			},
		},
		{
			name: "nil schema",
			s:    nil,
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "empty schema",
			s:    &schema.Schema{},
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "maxLength error",
			s:    schematest.New("string", schematest.WithMaxLength(2)),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/maxLength: string 'foo' exceeds maximum of 2")
			},
		},
		{
			name: "maxLength",
			s:    schematest.New("string", schematest.WithMaxLength(3)),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "minLength error",
			s:    schematest.New("string", schematest.WithMinLength(4)),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/minLength: string 'foo' is less than minimum of 4")
			},
		},
		{
			name: "minLength",
			s:    schematest.New("string", schematest.WithMinLength(3)),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
		{
			name: "pattern syntax error",
			s:    schematest.New("string", schematest.WithPattern("[")),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/pattern: validate string 'foo' with regex pattern '[' failed: error parsing regex: missing closing ]")
			},
		},
		{
			name: "pattern error",
			s:    schematest.New("string", schematest.WithPattern("[0-9]{4}")),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/pattern: string 'foo' does not match regex pattern '[0-9]{4}'")
			},
		},
		{
			name: "pattern valid but not maxLength",
			s:    schematest.New("string", schematest.WithPattern("[0-9]*"), schematest.WithMaxLength(3)),
			d:    "1234",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/maxLength: string '1234' exceeds maximum of 3")
			},
		},
		{
			name: "pattern",
			s:    schematest.New("string", schematest.WithPattern("[0-9]{4}")),
			d:    "1234",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "1234", v)
			},
		},
		{
			name: "format date error",
			s:    schematest.New("string", schematest.WithFormat("date")),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string 'foo' does not match format 'date'")
			},
		},
		{
			name: "format date",
			s:    schematest.New("string", schematest.WithFormat("date")),
			d:    "2018-11-13",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2018-11-13", v)
			},
		},
		{
			name: "format date-time error",
			s:    schematest.New("string", schematest.WithFormat("date-time")),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string 'foo' does not match format 'date-time'")
			},
		},
		{
			name: "format date-time",
			s:    schematest.New("string", schematest.WithFormat("date-time")),
			d:    "2018-11-13T20:20:39+00:00",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2018-11-13T20:20:39+00:00", v)
			},
		},
		{
			name: "format time error",
			s:    schematest.New("string", schematest.WithFormat("time")),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string 'foo' does not match format 'time'")
			},
		},
		{
			name: "format time",
			s:    schematest.New("string", schematest.WithFormat("time")),
			d:    "20:20:39+00:00",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "20:20:39+00:00", v)
			},
		},
		{
			name: "format duration error",
			s:    schematest.New("string", schematest.WithFormat("duration")),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string 'foo' does not match format 'duration'")
			},
		},
		{
			name: "format duration",
			s:    schematest.New("string", schematest.WithFormat("duration")),
			d:    "P3D",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "P3D", v)
			},
		},
		{
			name: "format email error",
			s:    schematest.New("string", schematest.WithFormat("email")),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string 'foo' does not match format 'email'")
			},
		},
		{
			name: "format email",
			s:    schematest.New("string", schematest.WithFormat("email")),
			d:    "foo@bar.com",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo@bar.com", v)
			},
		},
		{
			name: "format uuid error",
			s:    schematest.New("string", schematest.WithFormat("uuid")),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string 'foo' does not match format 'uuid'")
			},
		},
		{
			name: "format uuid",
			s:    schematest.New("string", schematest.WithFormat("uuid")),
			d:    "3e4666bf-d5e5-4aa7-b8ce-cefe41c7568a",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "3e4666bf-d5e5-4aa7-b8ce-cefe41c7568a", v)
			},
		},
		{
			name: "format ipv4 error",
			s:    schematest.New("string", schematest.WithFormat("ipv4")),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string 'foo' does not match format 'ipv4'")
			},
		},
		{
			name: "format ipv4 but ipv6",
			s:    schematest.New("string", schematest.WithFormat("ipv4")),
			d:    "1080:0:0:0:8:800:200C:417A",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string '1080:0:0:0:8:800:200C:417A' does not match format 'ipv4'")
			},
		},
		{
			name: "format ipv4",
			s:    schematest.New("string", schematest.WithFormat("ipv4")),
			d:    "192.168.1.1",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "192.168.1.1", v)
			},
		},
		{
			name: "format ipv6 error",
			s:    schematest.New("string", schematest.WithFormat("ipv6")),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string 'foo' does not match format 'ipv6'")
			},
		},
		{
			name: "format ipv6 but ipv4",
			s:    schematest.New("string", schematest.WithFormat("ipv6")),
			d:    "192.168.1.1",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string '192.168.1.1' does not match format 'ipv6'")
			},
		},
		{
			name: "format ipv6",
			s:    schematest.New("string", schematest.WithFormat("ipv6")),
			d:    "1080:0:0:0:8:800:200C:417A",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "1080:0:0:0:8:800:200C:417A", v)
			},
		},
		{
			name: "disable format validation",
			s:    schematest.New("string", schematest.WithFormat("date")),
			d:    "",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "", v)
			},

			skipValidationFormat: true,
		},
		{
			name: "const error",
			s:    schematest.New("string", schematest.WithConst("foo")),
			d:    "bar",
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/const: value 'bar' does not match const 'foo'")
			},
		},
		{
			name: "const",
			s:    schematest.New("string", schematest.WithConst("foo")),
			d:    "foo",
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := parser.Parser{Schema: tc.s, SkipValidationFormatKeyword: tc.skipValidationFormat}
			v, err := p.Parse(tc.d)
			tc.test(t, v, err)
		})
	}
}
