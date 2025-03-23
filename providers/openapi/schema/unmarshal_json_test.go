package schema_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math"
	"mokapi/media"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	jsonSchema "mokapi/schema/json/schema"
	"testing"
)

func TestRef_Unmarshal_Json(t *testing.T) {
	testcases := []struct {
		name   string
		data   string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name:   "empty with no schema",
			data:   `""`,
			schema: nil,
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "", i)
			},
		},
		{
			name:   "empty with schema but no type",
			data:   `""`,
			schema: &schema.Schema{},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "", i)
			},
		},
		{
			name:   "object and no schema",
			data:   `{"foo": 12}`,
			schema: nil,
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": float64(12)}, i)
			},
		},
		{
			name:   "array and no schema",
			data:   `[1, 2, 3, 4]`,
			schema: nil,
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{float64(1), float64(2), float64(3), float64(4)}, i)
			},
		},
		{
			name:   "object with integer ",
			data:   `{"foo": 12}`,
			schema: schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": int64(12)}, i)
			},
		},
		{
			name:   "object with string",
			data:   `{"foo": "bar"}`,
			schema: schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, i)
			},
		},
		{
			name: "object with string format",
			data: `{"foo": "2021-01-20"}`,
			schema: schematest.New("object",
				schematest.WithProperty("foo",
					schematest.New("string", schematest.WithFormat("date")))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "2021-01-20"}, i)
			},
		},
		{
			name: "object containing array",
			data: `{"foo": ["a", "b", "c"]}`,
			schema: schematest.New("object",
				schematest.WithProperty("foo",
					schematest.New("array", schematest.WithItems("string")))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": []interface{}{"a", "b", "c"}}, i)
			},
		},
		{
			name:   "null",
			data:   `{ "foo": null }`,
			schema: schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/foo/type: invalid type, expected string but got null")
			},
		},
		{
			name:   "accept strings or numbers, value is string",
			schema: schematest.New("string", schematest.And("number")),
			data:   `"foo"`,
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", i)
			},
		},
		{
			name:   "accept strings or numbers, value is number",
			schema: schematest.New("string", schematest.And("number")),
			data:   `12`,
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(12), i)
			},
		},
		{
			name:   "accept strings or numbers, value is boolean",
			schema: schematest.New("string", schematest.And("number")),
			data:   `true`,
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/type: invalid type, expected number but got boolean")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mt := &openapi.MediaType{Schema: tc.schema}
			i, err := mt.Parse([]byte(tc.data), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}

func TestRef_Unmarshal_Json_Any(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name: "any",
			s:    "12",
			schema: schematest.New("",
				schematest.Any(
					schematest.New("string"),
					schematest.New("integer"))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), i)
			},
		},
		{
			name: "not match any",
			s:    "12.6",
			schema: schematest.New("",
				schematest.Any(
					schematest.New("string"),
					schematest.New("integer"))),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/anyOf: does not match any schemas")
			},
		},
		{
			name: "any object",
			s:    `{"foo": "bar"}`,
			schema: schematest.NewAny(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("string")))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, i)
			},
		},
		{
			name: "missing required property should not error",
			s:    `{"name": "bar"}`,
			schema: schematest.NewAny(
				schematest.New("object",
					schematest.WithProperty("name", schematest.New("string"))),
				schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("age", schematest.New("integer")),
					schematest.WithRequired("age"),
				)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "bar"}, i)
			},
		},
		{
			name: "merge",
			s:    `{"name": "bar", "age": 12}`,
			schema: schematest.NewAny(
				schematest.New("object",
					schematest.WithProperty("name", schematest.New("string"))),
				schematest.New("object",
					schematest.WithProperty("age", schematest.New("integer")),
					schematest.WithRequired("age"),
				)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "bar", "age": int64(12)}, i)
			},
		},
		{
			name: "anyOf: object containing both properties",
			s:    `{"test": 12, "test2": true}`,
			schema: schematest.New("object",
				schematest.Any(
					schematest.New("object", schematest.WithProperty("test", schematest.New("integer"))),
					schematest.New("object", schematest.WithProperty("test2", schematest.New("boolean"))),
				)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"test": int64(12), "test2": true}, i)
			},
		},
		{
			name: "anyOf",
			s:    `"hello world"`,
			schema: schematest.NewAny(
				schematest.New("object", schematest.WithProperty("test", schematest.New("integer"))),
				schematest.New("string"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "hello world", i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mt := &openapi.MediaType{Schema: tc.schema}
			i, err := mt.Parse([]byte(tc.s), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}

func TestRef_Unmarshal_Json_String(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
		err    error
	}{
		{
			name:   "not string",
			s:      `12`,
			schema: schematest.New("string"),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/type: invalid type, expected string but got number")
			},
		},
		{
			name:   "string",
			s:      `"gbRMaRxHkiJBPta"`,
			schema: schematest.New("string"),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "gbRMaRxHkiJBPta", i)
			},
		},
		{
			name:   "type not defined",
			s:      `"gbRMaRxHkiJBPta"`,
			schema: &schema.Schema{},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "gbRMaRxHkiJBPta", i)
			},
		},
		{
			name:   "by pattern",
			s:      `"013-64-5994"`,
			schema: schematest.New("string", schematest.WithPattern("^\\d{3}-\\d{2}-\\d{4}$")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "013-64-5994", i)
			},
		},
		{
			name:   "not pattern",
			s:      `"013-64-59943"`,
			schema: schematest.New("string", schematest.WithPattern("^\\d{3}-\\d{2}-\\d{4}$")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/pattern: string '013-64-59943' does not match regex pattern '^\\d{3}-\\d{2}-\\d{4}$'")
			},
		},
		{
			name:   "date",
			s:      `"1908-12-07"`,
			schema: schematest.New("string", schematest.WithFormat("date")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "1908-12-07", i)
			},
		},
		{
			name:   "not date",
			s:      `"1908-12-7"`,
			schema: schematest.New("string", schematest.WithFormat("date")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string '1908-12-7' does not match format 'date'")
			},
		},
		{
			name:   "date-time",
			s:      `"1908-12-07T04:14:25Z"`,
			schema: schematest.New("string", schematest.WithFormat("date-time")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "1908-12-07T04:14:25Z", i)
			},
		},
		{
			name:   "not date-time",
			s:      `"1908-12-07 T04:14:25Z"`,
			schema: schematest.New("string", schematest.WithFormat("date-time")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string '1908-12-07 T04:14:25Z' does not match format 'date-time'")
			},
		},
		{
			name:   "password",
			s:      `"H|$9lb{J<+S;"`,
			schema: schematest.New("string", schematest.WithFormat("password")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "H|$9lb{J<+S;", i)
			},
		},
		{
			name:   "email",
			s:      `"markusmoen@pagac.net"`,
			schema: schematest.New("string", schematest.WithFormat("email")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "markusmoen@pagac.net", i)
			},
		},
		{
			name:   "not email",
			s:      `"markusmoen@@pagac.net"`,
			schema: schematest.New("string", schematest.WithFormat("email")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string 'markusmoen@@pagac.net' does not match format 'email'")
			},
		},
		{
			name:   "uuid",
			s:      `"590c1440-9888-45b0-bd51-a817ee07c3f2"`,
			schema: schematest.New("string", schematest.WithFormat("uuid")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "590c1440-9888-45b0-bd51-a817ee07c3f2", i)
			},
		},
		{
			name:   "not uuid",
			s:      `"590c1440-9888-45b0-bd51-a817ee07c3f2a"`,
			schema: schematest.New("string", schematest.WithFormat("uuid")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string '590c1440-9888-45b0-bd51-a817ee07c3f2a' does not match format 'uuid'")
			},
		},
		{
			name:   "ipv4",
			s:      `"152.23.53.100"`,
			schema: schematest.New("string", schematest.WithFormat("ipv4")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "152.23.53.100", i)
			},
		},
		{
			name:   "not ipv4",
			s:      `"152.23.53.100."`,
			schema: schematest.New("string", schematest.WithFormat("ipv4")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string '152.23.53.100.' does not match format 'ipv4'")
			},
		},
		{
			name:   "ipv6",
			s:      `"8898:ee17:bc35:9064:5866:d019:3b95:7857"`,
			schema: schematest.New("string", schematest.WithFormat("ipv6")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "8898:ee17:bc35:9064:5866:d019:3b95:7857", i)
			},
		},
		{
			name:   "not ipv6",
			s:      `"-8898:ee17:bc35:9064:5866:d019:3b95:7857"`,
			schema: schematest.New("string", schematest.WithFormat("ipv6")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: string '-8898:ee17:bc35:9064:5866:d019:3b95:7857' does not match format 'ipv6'")
			},
		},
		{
			name:   "not minLength",
			s:      `"foo"`,
			schema: schematest.New("string", schematest.WithMinLength(4)),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/minLength: string 'foo' is less than minimum of 4")
			},
		},
		{
			name:   "minLength",
			s:      `"foo"`,
			schema: schematest.New("string", schematest.WithMinLength(3)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", i)
			},
		},
		{
			name:   "not maxLength",
			s:      `"foo"`,
			schema: schematest.New("string", schematest.WithMaxLength(2)),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/maxLength: string 'foo' exceeds maximum of 2")
			},
		},
		{
			name:   "maxLength",
			s:      `"foo"`,
			schema: schematest.New("string", schematest.WithMaxLength(3)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", i)
			},
		},
		{
			name:   "enum",
			s:      `"foo"`,
			schema: schematest.New("string", schematest.WithEnumValues("foo")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", i)
			},
		},
		{
			name:   "not in enum",
			s:      `"foo"`,
			schema: schematest.New("string", schematest.WithEnumValues("bar")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/enum: value 'foo' does not match one in the enumeration [bar]", i)
			},
		},
		{
			name:   "nullable string",
			s:      `null`,
			schema: schematest.New("string", schematest.IsNullable(true)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Nil(t, i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mt := openapi.MediaType{Schema: tc.schema}
			i, err := mt.Parse([]byte(tc.s), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}

func TestRef_Unmarshal_Json_OneOf(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name: "valid oneOf",
			s:    `{"foo": true}`,
			schema: schematest.NewOneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("boolean"))),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": true}, i)
			},
		},
		{
			name: "no match",
			s:    `{"foo": "bar", "bar": 12}`,
			schema: schematest.NewOneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("bar", schematest.New("boolean"))),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/oneOf: valid against no schemas\n\t- #/oneOf/1/bar/type: invalid type, expected boolean but got number")
			},
		},
		{
			name: "two match",
			s:    `{"foo": 12}`,
			schema: schematest.NewOneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("number"))),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/oneOf: valid against more than one schema: valid schema indexes: 0, 1")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.s, func(t *testing.T) {
			t.Parallel()
			mt := &openapi.MediaType{Schema: tc.schema}
			i, err := mt.Parse([]byte(tc.s), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}

func TestRef_Unmarshal_Json_AllOf(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name: "valid",
			s:    `{"foo": 12, "bar": true}`,
			schema: schematest.New("object",
				schematest.AllOf(
					schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
					schematest.New("object", schematest.WithProperty("bar", schematest.New("boolean"))),
				)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": int64(12), "bar": true}, i)
			},
		},
		{
			name: "missing required property",
			s:    `{"foo": 12}`,
			schema: schematest.New("object",
				schematest.AllOf(
					schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
					schematest.New("object",
						schematest.WithRequired("bar"),
						schematest.WithProperty("bar", schematest.New("boolean"))),
				)),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/allOf: does not match all schema\n\t- #/allOf/1/required: required properties are missing: bar")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.s, func(t *testing.T) {
			t.Parallel()

			mt := &openapi.MediaType{Schema: tc.schema}
			i, err := mt.Parse([]byte(tc.s), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}

func TestRef_Unmarshal_Json_Integer(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		exp    int
		err    error
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name:   "int32",
			s:      "12",
			schema: schematest.New("integer", schematest.WithFormat("int32")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), i)
			},
		},
		{
			name:   "not int",
			s:      "3.61",
			schema: schematest.New("integer", schematest.WithFormat("int32")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/type: invalid type, expected integer but got number")
			},
		},
		{
			name:   "not int32",
			s:      fmt.Sprintf("%v", math.MaxInt64),
			schema: schematest.New("integer", schematest.WithFormat("int32")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: integer '9.223372036854776e+18' does not match format 'int32'")
			},
		},
		{
			name:   "min",
			s:      "12",
			schema: schematest.New("integer", schematest.WithMinimum(5)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), i)
			},
		},
		{
			name:   "minimum=12 and value=12",
			s:      "12",
			schema: schematest.New("integer", schematest.WithMinimum(13)),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/minimum: integer 12 is less than minimum value of 13")
			},
		},
		{
			name:   "minimum=12, exclusiveMinimum=true and value=12",
			s:      "12",
			schema: &schema.Schema{SubSchema: &schema.SubSchema{Type: jsonSchema.Types{"integer"}, Minimum: toFloatP(12), ExclusiveMinimum: jsonSchema.NewUnionTypeB[float64, bool](true)}},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/minimum: integer 12 equals minimum value of 12 and exclusive minimum is true")
			},
		},
		{
			name:   "minimum=12, exclusiveMinimum=false and value=12",
			s:      "12",
			schema: &schema.Schema{SubSchema: &schema.SubSchema{Type: jsonSchema.Types{"integer"}, Minimum: toFloatP(12), ExclusiveMinimum: jsonSchema.NewUnionTypeB[float64, bool](false)}},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), i)
			},
		},
		{
			name:   "exclusiveMinimum=12 and value=12",
			s:      "12",
			schema: &schema.Schema{SubSchema: &schema.SubSchema{Type: jsonSchema.Types{"integer"}, ExclusiveMinimum: jsonSchema.NewUnionTypeA[float64, bool](12)}},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/exclusiveMinimum: integer 12 equals minimum value of 12")
			},
		},
		{
			name:   "exclusiveMinimum=true but no minimum is set",
			s:      "12",
			schema: &schema.Schema{SubSchema: &schema.SubSchema{Type: jsonSchema.Types{"integer"}, ExclusiveMinimum: jsonSchema.NewUnionTypeB[float64, bool](true)}},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), i)
			},
		},
		{
			name:   "max",
			s:      "12",
			schema: schematest.New("integer", schematest.WithMaximum(13)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), i)
			},
		},
		{
			name:   "not max",
			s:      "12",
			schema: schematest.New("integer", schematest.WithMaximum(11)),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/maximum: integer 12 exceeds maximum value of 11")
			},
		},
		{
			name:   "not exclusive max",
			s:      "12",
			schema: &schema.Schema{SubSchema: &schema.SubSchema{Type: jsonSchema.Types{"integer"}, Maximum: toFloatP(12), ExclusiveMaximum: jsonSchema.NewUnionTypeB[float64, bool](true)}},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/maximum: integer 12 equals maximum value of 12 and exclusive maximum is true")
			},
		},
		{
			name:   "not in enum",
			s:      "12",
			schema: schematest.New("integer", schematest.WithEnumValues(1, 2, 3)),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/enum: value '12' does not match one in the enumeration [1, 2, 3]")
			},
		},
		{
			name:   "in enum",
			s:      "2",
			schema: schematest.New("integer", schematest.WithEnumValues(1, 2, 3)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(2), i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mt := &openapi.MediaType{Schema: tc.schema}
			i, err := mt.Parse([]byte(tc.s), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}

func TestParse_Number(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
		exp    float64
		err    error
	}{
		{
			name:   "float32",
			s:      "3.612",
			schema: schematest.New("number", schematest.WithFormat("float32")),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 3.612, i)
			},
		},
		{
			name:   "type not defined",
			s:      "3.612",
			schema: &schema.Schema{},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 3.612, i)
			},
		},
		{
			name:   "float out of range",
			s:      fmt.Sprintf("%v", math.MaxFloat64),
			schema: schematest.New("number", schematest.WithFormat("float")),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: number '1.7976931348623157e+308' does not match format 'float'")
			},
		},
		{
			name:   "min",
			s:      "3.612",
			schema: schematest.New("number", schematest.WithMinimum(3.6)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 3.612, i)
			},
		},
		{
			name:   "not min",
			s:      "3.612",
			schema: schematest.New("number", schematest.WithMinimum(3.7)),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/minimum: number 3.612 is less than minimum value of 3.7")
			},
		},
		{
			name:   "not min float",
			s:      "3.612",
			schema: schematest.New("number", schematest.WithFormat("float"), schematest.WithMinimum(3.7)),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: number '3.612' does not match format 'float'")
			},
		},
		{
			name:   "max",
			s:      "3.612",
			schema: schematest.New("number", schematest.WithMaximum(3.7)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 3.612, i)
			},
		},
		{
			name:   "not max",
			s:      "3.612",
			schema: schematest.New("number", schematest.WithMaximum(3.6)),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/maximum: number 3.612 exceeds maximum value of 3.6")
			},
		},
		{
			name:   "not max float",
			s:      "3.612",
			schema: schematest.New("number", schematest.WithFormat("float"), schematest.WithMaximum(3.6)),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/format: number '3.612' does not match format 'float'")
			},
		},
		{
			name:   "not max exclusive",
			s:      "3.6",
			schema: &schema.Schema{SubSchema: &schema.SubSchema{Type: jsonSchema.Types{"number"}, Maximum: toFloatP(3.6), ExclusiveMaximum: jsonSchema.NewUnionTypeB[float64, bool](true)}},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/maximum: number 3.6 equals maximum value of 3.6 and exclusive maximum is true")
			},
		},
		{
			name:   "exclusiveMaximum=true but no maximum value is set",
			s:      "3.6",
			schema: &schema.Schema{SubSchema: &schema.SubSchema{Type: jsonSchema.Types{"number"}, ExclusiveMaximum: jsonSchema.NewUnionTypeB[float64, bool](true)}},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 3.6, i)
			},
		},
		{
			name:   "not in enum",
			s:      "3.6",
			schema: &schema.Schema{SubSchema: &schema.SubSchema{Type: jsonSchema.Types{"number"}, Enum: []interface{}{3, 4, 5.5}}},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/enum: value '3.6' does not match one in the enumeration [3, 4, 5.5]")
			},
		},
		{
			name:   "in enum",
			s:      "3.6",
			schema: schematest.New("number", schematest.WithEnumValues(3.6, 4, 5.5)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 3.6, i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mt := &openapi.MediaType{Schema: tc.schema}
			i, err := mt.Parse([]byte(tc.s), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}

func TestRef_Unmarshal_Json_Object(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		test   func(t *testing.T, v interface{}, err error)
	}{
		{
			name:   "empty",
			s:      "{}",
			schema: schematest.New("object"),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{}, v)
			},
		},
		{
			name:   "type not defined",
			s:      `{"foo":"bar"}`,
			schema: &schema.Schema{},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "bar"}, v)
			},
		},
		{
			name:   "example type not defined",
			s:      `{"type":"object","properties":{"id":{"format":"int64","type":"integer"}}}`,
			schema: &schema.Schema{},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"format": "int64",
							"type":   "integer",
						},
					},
				}, v)
			},
		},
		{
			name: "simple",
			s:    `{"name": "foo", "age": 12}`,
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("age", schematest.New("integer")),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "foo", "age": int64(12)}, v)
			},
		},
		{
			name: "missing but not required",
			s:    `{"name": "foo"}`,
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("age", schematest.New("integer")),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "foo"}, v)
			},
		},
		{
			name: "missing but required",
			s:    `{"name": "foo"}`,
			schema: schematest.New("object",
				schematest.WithRequired("name", "age"),
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("age", schematest.New("integer")),
			),
			test: func(t *testing.T, _ interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/required: required properties are missing: age")
			},
		},
		{
			name: "property is not valid",
			s:    `{"name": "abc"}`,
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string", schematest.WithMinLength(6))),
				schematest.WithProperty("age", schematest.New("integer")),
			),
			test: func(t *testing.T, _ interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/name/minLength: string 'abc' is less than minimum of 6")
			},
		},
		{
			name: "not in enum",
			s:    `{"name": "foo"}`,
			schema: schematest.New("object",
				schematest.WithRequired("name"),
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithEnum([]interface{}{map[string]interface{}{"name": "bar"}}),
			),
			test: func(t *testing.T, _ interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/enum: value '{name: foo}' does not match one in the enumeration [{name: bar}]")
			},
		},
		{
			name: "in enum",
			s:    `{"name": "foo"}`,
			schema: schematest.New("object",
				schematest.WithRequired("name"),
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithEnum([]interface{}{map[string]interface{}{"name": "foo"}}),
			),
			test: func(t *testing.T, _ interface{}, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "in enum nested",
			s:    `{"name": "foo", "colors": ["red", "green", "blue"] }`,
			schema: schematest.New("object",
				schematest.WithRequired("name"),
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("colors", schematest.New("array",
					schematest.WithItems("string"))),
				schematest.WithEnum([]interface{}{map[string]interface{}{"name": "foo", "colors": []string{"red", "green", "blue"}}}),
			),
			test: func(t *testing.T, _ interface{}, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "with additional property",
			s:    `{"name": "foo", "age": 12}`,
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "foo", "age": float64(12)}, v)
			},
		},
		{
			name: "with additional property but not allowed",
			s:    `{"name": "foo", "age": 12}`,
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithFreeForm(false),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/additionalProperties: property 'age' not defined and the schema does not allow additional properties")
			},
		},
		{
			name: "not minProperties",
			s:    `{"name": "foo"}`,
			schema: schematest.New("object",
				schematest.WithMinProperties(2),
			),
			test: func(t *testing.T, _ interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/minProperties: property count 1 is less than minimum count of 2")
			},
		},
		{
			name: "not maxProperties",
			s:    `{"name": "foo", "age": 12}`,
			schema: schematest.New("object",
				schematest.WithMaxProperties(1),
			),
			test: func(t *testing.T, _ interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/maxProperties: property count 2 exceeds maximum count of 1")
			},
		},
		{
			name: "min-max properties, free-form order of properties not defined",
			s:    `{"name": "foo", "age": 12}`,
			schema: schematest.New("object",
				schematest.WithMinProperties(1),
				schematest.WithMaxProperties(2),
			),
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "foo", "age": float64(12)}, v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mt := &openapi.MediaType{Schema: tc.schema}
			i, err := mt.Parse([]byte(tc.s), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}

func TestRef_Unmarshal_Json_Array(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name: "empty",
			s:    "[]",
			schema: schematest.New("array",
				schematest.WithItems("string"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Len(t, i, 0)
			},
		},
		{
			name: "string array",
			s:    `["foo", "bar"]`,
			schema: schematest.New("array",
				schematest.WithItems("string"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", "bar"}, i)
			},
		},
		{
			name: "not min items",
			s:    `["foo", "bar"]`,
			schema: schematest.New("array", schematest.WithMinItems(3),
				schematest.WithItems("string"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/minItems: item count 2 is less than minimum count of 3")
			},
		},
		{
			name: "min items",
			s:    `["foo", "bar"]`,
			schema: schematest.New("array", schematest.WithMinItems(2),
				schematest.WithItems("string"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", "bar"}, i)
			},
		},
		{
			name: "not max items",
			s:    `["foo", "bar"]`,
			schema: schematest.New("array", schematest.WithMaxItems(1),
				schematest.WithItems("string"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/maxItems: item count 2 exceeds maximum count of 1")
			},
		},
		{
			name: "max items",
			s:    `["foo", "bar"]`,
			schema: schematest.New("array", schematest.WithMaxItems(2),
				schematest.WithItems("string"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", "bar"}, i)
			},
		},
		{
			name: "not in enum",
			s:    `["foo", "bar"]`,
			schema: schematest.New("array", schematest.WithEnumValues([]interface{}{"foo", "test"}),
				schematest.WithItems("string"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/enum: value '[foo, bar]' does not match one in the enumeration [[foo, test]]")
			},
		},
		{
			name: "not with correct order in enum",
			s:    `["foo", "bar"]`,
			schema: schematest.New("array", schematest.WithEnumValues([]interface{}{"bar", "foo"}),
				schematest.WithItems("string"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/enum: value '[foo, bar]' does not match one in the enumeration [[bar, foo]]")
			},
		},
		{
			name: "in enum",
			s:    `["foo", "bar"]`,
			schema: schematest.New("array", schematest.WithEnumValues([]interface{}{"foo", "test"}, []interface{}{"foo", "bar"}),
				schematest.WithItems("string"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", "bar"}, i)
			},
		},
		{
			name: "not uniqueItems",
			s:    `["foo", "foo"]`,
			schema: schematest.New("array", schematest.WithUniqueItems(),
				schematest.WithItems("string"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/uniqueItems: non-unique array item at index 1")
			},
		},
		{
			name: "uniqueItems",
			s:    `["foo", "bar"]`,
			schema: schematest.New("array", schematest.WithUniqueItems(),
				schematest.WithItems("string"),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"foo", "bar"}, i)
			},
		},
		{
			name: "object array",
			s:    `[{"name": "foo", "age": 12}]`,
			schema: schematest.New("array",
				schematest.WithItems(
					"object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("age", schematest.New("integer")),
				),
			),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{map[string]interface{}{"name": "foo", "age": int64(12)}}, i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mt := &openapi.MediaType{Schema: tc.schema}
			i, err := mt.Parse([]byte(tc.s), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}

func TestRef_Unmarshal_Json_Bool(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
		exp    bool
		err    error
	}{
		{
			name:   "true",
			s:      `true`,
			schema: schematest.New("boolean"),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, true, i)
			},
		},
		{
			name:   "int to bool not possible",
			s:      `1`,
			schema: schematest.New("boolean"),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "error count 1:\n- #/type: invalid type, expected boolean but got number")
			},
		},
		{
			name:   "false",
			s:      `false`,
			schema: schematest.New("boolean"),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, false, i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mt := &openapi.MediaType{Schema: tc.schema}
			i, err := mt.Parse([]byte(tc.s), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}

func TestRef_Unmarshal_Json_Errors(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		test   func(t *testing.T, err error)
	}{
		{
			name:   "empty json",
			s:      ``,
			schema: &schema.Schema{},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "invalid json format: unexpected end of JSON input")
			},
		},
		{
			name:   "invalid json",
			s:      `bar`,
			schema: schematest.New("string"),
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "invalid json format: invalid character 'b' looking for beginning of value")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mt := &openapi.MediaType{Schema: tc.schema}
			_, err := mt.Parse([]byte(tc.s), media.ParseContentType("application/json"))
			tc.test(t, err)
		})
	}
}

func TestRef_Unmarshal_Json_Dictionary(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
		exp    interface{}
		err    error
	}{
		{
			name:   "string:string",
			s:      `{"en": "English", "de": "German"}`,
			schema: schematest.New("object", schematest.WithAdditionalProperties(schematest.New("string"))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"en": "English", "de": "German"}, i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mt := &openapi.MediaType{Schema: tc.schema}
			_, err := mt.Parse([]byte(tc.s), media.ParseContentType("application/json"))
			require.Equal(t, tc.err, err)
		})
	}
}

func TestRef_Unmarshal_Json_SpecialNames(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		test   func(t *testing.T, i interface{}, err error)
	}{
		{
			name:   "kebab case",
			s:      `{"ship-date":"2022-01-01"}`,
			schema: schematest.New("object", schematest.WithProperty("ship-date", schematest.New("string"))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"ship-date": "2022-01-01"}, i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mt := &openapi.MediaType{Schema: tc.schema}
			i, err := mt.Parse([]byte(tc.s), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}
