package schema_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/media"
	"reflect"
	"strings"
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
			schema: schematest.New(""),
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
				require.Equal(t, &struct {
					Foo float64 `json:"foo"`
				}{Foo: 12}, i)
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
				require.Equal(t, &struct {
					Foo int64 `json:"foo"`
				}{Foo: int64(12)}, i)
			},
		},
		{
			name:   "object with string",
			data:   `{"foo": "bar"}`,
			schema: schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Foo string `json:"foo"`
				}{Foo: "bar"}, i)
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
				require.Equal(t, &struct {
					Foo string `json:"foo"`
				}{Foo: "2021-01-20"}, i)
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
				require.Equal(t, &struct {
					Foo []string `json:"foo"`
				}{Foo: []string{"a", "b", "c"}}, i)
			},
		},
		{
			name: "object containing two properties",
			data: `{"test": 12, "test2": true}`,
			schema: schematest.New("object",
				schematest.Any(
					schematest.New("object", schematest.WithProperty("test", schematest.New("integer"))),
					schematest.New("object", schematest.WithProperty("test2", schematest.New("boolean"))),
				)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Test  int64 `json:"test"`
					Test2 bool  `json:"test2"`
				}{Test: 12, Test2: true}, i)
			},
		},
		{
			name: "anyOf",
			data: `"hello world"`,
			schema: schematest.New("object",
				schematest.Any(
					schematest.New("object", schematest.WithProperty("test", schematest.New("integer"))),
					schematest.New("string"),
				)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "hello world", i)
			},
		},
		{
			name:   "null",
			data:   `{ "foo": null }`,
			schema: schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "parse foo failed: could not parse <nil> as string, expected schema type=string")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := &schema.Ref{Value: tc.schema}
			i, err := r.Unmarshal([]byte(tc.data), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}

func TestRef_Unmarshal_Json_String(t *testing.T) {
	maxLength2 := 2
	maxLength3 := 3

	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		err    error
	}{
		{
			name:   "not string",
			s:      `12`,
			schema: &schema.Schema{Type: "string"},
			err:    fmt.Errorf("could not parse 12 as string, expected schema type=string"),
		},
		{
			name:   "string",
			s:      `"gbRMaRxHkiJBPta"`,
			schema: &schema.Schema{Type: "string"},
			err:    nil,
		},
		{
			name:   "type not defined",
			s:      `"gbRMaRxHkiJBPta"`,
			schema: &schema.Schema{},
			err:    nil,
		},
		{
			name:   "by pattern",
			s:      `"013-64-5994"`,
			schema: &schema.Schema{Type: "string", Pattern: "^\\d{3}-\\d{2}-\\d{4}$"},
			err:    nil,
		},
		{
			name:   "not pattern",
			s:      `"013-64-59943"`,
			schema: &schema.Schema{Type: "string", Pattern: "^\\d{3}-\\d{2}-\\d{4}$"},
			err:    fmt.Errorf("value '013-64-59943' does not match pattern, expected schema type=string pattern=^\\d{3}-\\d{2}-\\d{4}$"),
		},
		{
			name:   "date",
			s:      `"1908-12-07"`,
			schema: &schema.Schema{Type: "string", Format: "date"},
			err:    nil,
		},
		{
			name:   "not date",
			s:      `"1908-12-7"`,
			schema: &schema.Schema{Type: "string", Format: "date"},
			err:    fmt.Errorf("value '1908-12-7' is not a date RFC3339, expected schema type=string format=date"),
		},
		{
			name:   "date-time",
			s:      `"1908-12-07T04:14:25Z"`,
			schema: &schema.Schema{Type: "string", Format: "date-time"},
			err:    nil,
		},
		{
			name:   "not date-time",
			s:      `"1908-12-07 T04:14:25Z"`,
			schema: &schema.Schema{Type: "string", Format: "date-time"},
			err:    fmt.Errorf("value '1908-12-07 T04:14:25Z' is not a date-time RFC3339, expected schema type=string format=date-time"),
		},
		{
			name:   "password",
			s:      `"H|$9lb{J<+S;"`,
			schema: &schema.Schema{Type: "string", Format: "password"},
			err:    nil,
		},
		{
			name:   "email",
			s:      `"markusmoen@pagac.net"`,
			schema: &schema.Schema{Type: "string", Format: "email"},
			err:    nil,
		},
		{
			name:   "not email",
			s:      `"markusmoen@@pagac.net"`,
			schema: &schema.Schema{Type: "string", Format: "email"},
			err:    fmt.Errorf("value 'markusmoen@@pagac.net' is not an email address, expected schema type=string format=email"),
		},
		{
			name:   "uuid",
			s:      `"590c1440-9888-45b0-bd51-a817ee07c3f2"`,
			schema: &schema.Schema{Type: "string", Format: "uuid"},
			err:    nil,
		},
		{
			name:   "not uuid",
			s:      `"590c1440-9888-45b0-bd51-a817ee07c3f2a"`,
			schema: &schema.Schema{Type: "string", Format: "uuid"},
			err:    fmt.Errorf("value '590c1440-9888-45b0-bd51-a817ee07c3f2a' is not an uuid, expected schema type=string format=uuid"),
		},
		{
			name:   "ipv4",
			s:      `"152.23.53.100"`,
			schema: &schema.Schema{Type: "string", Format: "ipv4"},
			err:    nil,
		},
		{
			name:   "not ipv4",
			s:      `"152.23.53.100."`,
			schema: &schema.Schema{Type: "string", Format: "ipv4"},
			err:    fmt.Errorf("value '152.23.53.100.' is not an ipv4, expected schema type=string format=ipv4"),
		},
		{
			name:   "ipv6",
			s:      `"8898:ee17:bc35:9064:5866:d019:3b95:7857"`,
			schema: &schema.Schema{Type: "string", Format: "ipv6"},
			err:    nil,
		},
		{
			name:   "not ipv6",
			s:      `"-8898:ee17:bc35:9064:5866:d019:3b95:7857"`,
			schema: &schema.Schema{Type: "string", Format: "ipv6"},
			err:    fmt.Errorf("value '-8898:ee17:bc35:9064:5866:d019:3b95:7857' is not an ipv6, expected schema type=string format=ipv6"),
		},
		{
			name:   "not minLength",
			s:      `"foo"`,
			schema: &schema.Schema{Type: "string", MinLength: toIntP(4)},
			err:    fmt.Errorf("value 'foo' does not meet min length of 4"),
		},
		{
			name:   "minLength",
			s:      `"foo"`,
			schema: &schema.Schema{Type: "string", MinLength: toIntP(3)},
			err:    nil,
		},
		{
			name:   "not maxLength",
			s:      `"foo"`,
			schema: &schema.Schema{Type: "string", MaxLength: &maxLength2},
			err:    fmt.Errorf("value 'foo' does not meet max length of 2"),
		},
		{
			name:   "maxLength",
			s:      `"foo"`,
			schema: &schema.Schema{Type: "string", MaxLength: &maxLength3},
			err:    nil,
		},
		{
			name:   "enum",
			s:      `"foo"`,
			schema: &schema.Schema{Type: "string", Enum: []interface{}{"foo"}},
			err:    nil,
		},
		{
			name:   "not in enum",
			s:      `"foo"`,
			schema: &schema.Schema{Type: "string", Enum: []interface{}{"bar"}},
			err:    fmt.Errorf("value foo does not match one in the enum [bar]"),
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &schema.Ref{Value: tc.schema}
			_, err := r.Unmarshal([]byte(tc.s), media.ParseContentType("application/json"))
			require.Equal(t, tc.err, err)
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
				require.EqualError(t, err, "could not parse 12.6, expected any of schema type=string, schema type=integer")
			},
		},
		{
			name: "any object",
			s:    `{"foo": "bar"}`,
			schema: schematest.New("",
				schematest.Any(
					schematest.New("object",
						schematest.WithProperty("foo", schematest.New("integer"))),
					schematest.New("object",
						schematest.WithProperty("foo", schematest.New("string"))))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Foo string `json:"foo"`
				}{Foo: "bar"}, i)
			},
		},
		{
			name: "too many properties",
			s:    `{"name": "bar", "age": 12}`,
			schema: schematest.New("",
				schematest.Any(
					schematest.New("object",
						schematest.WithProperty("name", schematest.New("string"))))),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, `could not parse {name: bar, age: 12}, too many properties for object, expected any of schema type=object properties=[name]`)
			},
		},
		{
			name: "missing required property",
			s:    `{"name": "bar"}`,
			schema: schematest.New("",
				schematest.Any(
					schematest.New("object",
						schematest.WithProperty("name", schematest.New("string"))),
					schematest.New("object",
						schematest.WithProperty("age", schematest.New("integer")),
						schematest.WithRequired("age"),
					))),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "missing required property age, expected any of schema type=object properties=[name], schema type=object properties=[age] required=[age]")
			},
		},
		{
			name: "merge",
			s:    `{"name": "bar", "age": 12}`,
			schema: schematest.New("",
				schematest.Any(
					schematest.New("object",
						schematest.WithProperty("name", schematest.New("string"))),
					schematest.New("object",
						schematest.WithProperty("age", schematest.New("integer")),
						schematest.WithRequired("age"),
					))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Name string `json:"name"`
					Age  int64  `json:"age"`
				}{Name: "bar", Age: 12}, i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := &schema.Ref{Value: tc.schema}
			i, err := r.Unmarshal([]byte(tc.s), media.ParseContentType("application/json"))
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
			schema: schematest.New("", schematest.OneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("boolean"))),
			)),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Foo bool `json:"foo"`
				}{Foo: true}, i)
			},
		},
		{
			name: "no match",
			s:    `{"foo": 12, "bar": true}`,
			schema: schematest.New("", schematest.OneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("bar", schematest.New("boolean"))),
			)),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, `could not parse {foo: 12, bar: true}, expected one of schema type=object properties=[foo], schema type=object properties=[bar]`)
			},
		},
		{
			name: "two match",
			s:    `{"foo": 12}`,
			schema: schematest.New("", schematest.OneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("number"))),
			)),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, `could not parse {foo: 12}, it is not valid for only one schema, expected one of schema type=object properties=[foo], schema type=object properties=[foo]`)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.s, func(t *testing.T) {
			t.Parallel()
			r := &schema.Ref{Value: tc.schema}
			i, err := r.Unmarshal([]byte(tc.s), media.ParseContentType("application/json"))
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
				require.Equal(t, &struct {
					Foo int64 `json:"foo"`
					Bar bool  `json:"bar"`
				}{Foo: 12, Bar: true}, i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.s, func(t *testing.T) {
			t.Parallel()

			r := &schema.Ref{Value: tc.schema}
			i, err := r.Unmarshal([]byte(tc.s), media.ParseContentType("application/json"))
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
			schema: &schema.Schema{Type: "integer", Format: "int32"},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), i)
			},
		},
		{
			name:   "not int",
			s:      "3.61",
			schema: &schema.Schema{Type: "integer", Format: "int32"},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "could not parse 3.61 as integer, expected schema type=integer format=int32")
			},
		},
		{
			name:   "not int32",
			s:      fmt.Sprintf("%v", math.MaxInt64),
			schema: &schema.Schema{Type: "integer", Format: "int32"},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "could not parse '9.223372036854776e+18', represents a number either less than int32 min value or greater max value, expected schema type=integer format=int32")
			},
		},
		{
			name:   "min",
			s:      "12",
			schema: &schema.Schema{Type: "integer", Minimum: toFloatP(5)},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), i)
			},
		},
		{
			name:   "not min",
			s:      "12",
			schema: &schema.Schema{Type: "integer", Minimum: toFloatP(13)},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "12 is lower as the required minimum 13, expected schema type=integer minimum=13")
			},
		},
		{
			name:   "not exclusive min",
			s:      "12",
			schema: &schema.Schema{Type: "integer", Minimum: toFloatP(12), ExclusiveMinimum: toBoolP(true)},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "12 is lower or equal as the required minimum 12, expected schema type=integer minimum=12 exclusiveMinimum")
			},
		},
		{
			name:   "max",
			s:      "12",
			schema: &schema.Schema{Type: "integer", Maximum: toFloatP(13)},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), i)
			},
		},
		{
			name:   "not max",
			s:      "12",
			schema: &schema.Schema{Type: "integer", Maximum: toFloatP(11)},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "12 is greater as the required maximum 11, expected schema type=integer maximum=11")
			},
		},
		{
			name:   "not exclusive max",
			s:      "12",
			schema: &schema.Schema{Type: "integer", Maximum: toFloatP(12), ExclusiveMaximum: toBoolP(true)},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "12 is greater or equal as the required maximum 12, expected schema type=integer maximum=12 exclusiveMaximum")
			},
		},
		{
			name:   "not in enum",
			s:      "12",
			schema: &schema.Schema{Type: "integer", Enum: []interface{}{1, 2, 3}},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "value 12 does not match one in the enum [1, 2, 3]")
			},
		},
		{
			name:   "in enum",
			s:      "2",
			schema: &schema.Schema{Type: "integer", Enum: []interface{}{1, 2, 3}},
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

			r := &schema.Ref{Value: tc.schema}
			i, err := r.Unmarshal([]byte(tc.s), media.ParseContentType("application/json"))
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
			schema: &schema.Schema{Type: "number", Format: "float32"},
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
			schema: &schema.Schema{Type: "number", Format: "float"},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "could not parse 1.7976931348623157e+308 as float, expected schema type=number format=float")
			},
		},
		{
			name:   "min",
			s:      "3.612",
			schema: &schema.Schema{Type: "number", Minimum: toFloatP(3.6)},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 3.612, i)
			},
		},
		{
			name:   "not min",
			s:      "3.612",
			schema: &schema.Schema{Type: "number", Minimum: toFloatP(3.7)},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "3.612 is lower as the required minimum 3.7, expected schema type=number minimum=3.7")
			},
		},
		{
			name:   "not min float",
			s:      "3.612",
			schema: &schema.Schema{Type: "number", Format: "float", Minimum: toFloatP(3.7)},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "3.612 is lower as the required minimum 3.7, expected schema type=number format=float minimum=3.7")
			},
		},
		{
			name:   "max",
			s:      "3.612",
			schema: &schema.Schema{Type: "number", Maximum: toFloatP(3.7)},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 3.612, i)
			},
		},
		{
			name:   "not max",
			s:      "3.612",
			schema: &schema.Schema{Type: "number", Maximum: toFloatP(3.6)},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "3.612 is greater as the required maximum 3.6, expected schema type=number maximum=3.6")
			},
		},
		{
			name:   "not max float",
			s:      "3.612",
			schema: &schema.Schema{Type: "number", Format: "float", Maximum: toFloatP(3.6)},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "3.612 is greater as the required maximum 3.6, expected schema type=number format=float maximum=3.6")
			},
		},
		{
			name:   "not max exclusive",
			s:      "3.6",
			schema: &schema.Schema{Type: "number", Maximum: toFloatP(3.6), ExclusiveMaximum: toBoolP(true)},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "3.6 is greater or equal as the required maximum 3.6, expected schema type=number maximum=3.6 exclusiveMaximum")
			},
		},
		{
			name:   "not in enum",
			s:      "3.6",
			schema: &schema.Schema{Type: "number", Enum: []interface{}{3, 4, 5.5}},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "value 3.6 does not match one in the enum [3, 4, 5.5]")
			},
		},
		{
			name:   "in enum",
			s:      "3.6",
			schema: &schema.Schema{Type: "number", Enum: []interface{}{3.6, 4, 5.5}},
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

			r := &schema.Ref{Value: tc.schema}
			i, err := r.Unmarshal([]byte(tc.s), media.ParseContentType("application/json"))
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
			schema: &schema.Schema{Type: "object"},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct{}{}, v)
			},
		},
		{
			name:   "type not defined",
			s:      `{"foo":"bar"}`,
			schema: &schema.Schema{},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Foo string `json:"foo"`
				}{Foo: "bar"}, v)
			},
		},
		{
			name:   "example type not defined",
			s:      `{"type":"object","properties":{"id":{"format":"int64","type":"integer"}}}`,
			schema: &schema.Schema{},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				exp := &struct {
					Type       string `json:"type"`
					Properties *struct {
						Id *struct {
							Format string `json:"format"`
							Type   string `json:"type"`
						} `json:"id"`
					} `json:"properties"`
				}{
					Type: "object",
					Properties: &struct {
						Id *struct {
							Format string `json:"format"`
							Type   string `json:"type"`
						} `json:"id"`
					}{Id: &struct {
						Format string `json:"format"`
						Type   string `json:"type"`
					}{Format: "int64", Type: "integer"}},
				}
				require.Equal(t, exp, v)
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
				require.Equal(t, &struct {
					Name string `json:"name"`
					Age  int64  `json:"age"`
				}{Name: "foo", Age: 12}, v)
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
				require.Equal(t, &struct {
					Name string `json:"name"`
				}{Name: "foo"}, v)
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
				require.EqualError(t, err, `missing required field 'age'`)
			},
		},
		{
			name: "property is not valid",
			s:    `{"name": null}`,
			schema: schematest.New("object",
				schematest.WithProperty("name", schematest.New("string", schematest.WithMinLength(6))),
				schematest.WithProperty("age", schematest.New("integer")),
			),
			test: func(t *testing.T, _ interface{}, err error) {
				require.EqualError(t, err, `parse name failed: could not parse <nil> as string, expected schema type=string minLength=6`)
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
				require.EqualError(t, err, `value {name: foo} does not match one in the enum [{name: bar}]`)
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
			name: "not minProperties",
			s:    `{"name": "foo"}`,
			schema: schematest.New("object",
				schematest.WithMinProperties(2),
			),
			test: func(t *testing.T, _ interface{}, err error) {
				require.EqualError(t, err, `validation error minProperties on {name: foo}, expected schema type=object minProperties=2`)
			},
		},
		{
			name: "not maxProperties",
			s:    `{"name": "foo", "age": 12}`,
			schema: schematest.New("object",
				schematest.WithMaxProperties(1),
			),
			test: func(t *testing.T, _ interface{}, err error) {
				require.Truef(t, strings.HasSuffix(err.Error(), "expected schema type=object maxProperties=1"), "but error is %v", err)
			},
		},
		{
			name: "min-max properties, free-form order of properties not defined",
			s:    `{"name": "foo", "age": 12}`,
			schema: schematest.New("object",
				schematest.WithMinProperties(1),
				schematest.WithMaxProperties(2),
			),
			test: func(t *testing.T, v interface{}, _ error) {
				if !reflect.DeepEqual(
					&struct {
						Name string  `json:"name"`
						Age  float64 `json:"age"`
					}{
						Name: "foo", Age: 12,
					}, v) &&
					!reflect.DeepEqual(&struct {
						Age  float64 `json:"age"`
						Name string  `json:"name"`
					}{
						Name: "foo", Age: 12,
					}, v) {
					require.Fail(t, "not equal in any order")
				}
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &schema.Ref{Value: tc.schema}
			i, err := r.Unmarshal([]byte(tc.s), media.ParseContentType("application/json"))
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
			schema: &schema.Schema{Type: "array", Items: &schema.Ref{
				Value: &schema.Schema{Type: "string"},
			}},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{}, i)
			},
		},
		{
			name: "string array",
			s:    `["foo", "bar"]`,
			schema: &schema.Schema{Type: "array", Items: &schema.Ref{
				Value: &schema.Schema{Type: "string"},
			}},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, i)
			},
		},
		{
			name: "not min items",
			s:    `["foo", "bar"]`,
			schema: &schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				MinItems: toIntP(3),
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "should NOT have less than 3 items")
			},
		},
		{
			name: "min items",
			s:    `["foo", "bar"]`,
			schema: &schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				MinItems: toIntP(2),
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, i)
			},
		},
		{
			name: "not max items",
			s:    `["foo", "bar"]`,
			schema: &schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				MaxItems: toIntP(1),
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "should NOT have more than 1 items")
			},
		},
		{
			name: "max items",
			s:    `["foo", "bar"]`,
			schema: &schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				MaxItems: toIntP(2),
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, i)
			},
		},
		{
			name: "not in enum",
			s:    `["foo", "bar"]`,
			schema: &schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				Enum: []interface{}{[]string{"foo", "test"}},
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "value [foo, bar] does not match one in the enum [[foo, test]]")
			},
		},
		{
			name: "not with correct order in enum",
			s:    `["foo", "bar"]`,
			schema: &schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				Enum: []interface{}{[]string{"bar", "foo"}},
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "value [foo, bar] does not match one in the enum [[bar, foo]]")
			},
		},
		{
			name: "in enum",
			s:    `["foo", "bar"]`,
			schema: &schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				Enum: []interface{}{[]string{"foo", "test"}, []string{"foo", "bar"}},
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, i)
			},
		},
		{
			name: "not uniqueItems",
			s:    `["foo", "foo"]`,
			schema: &schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				UniqueItems: true,
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "should NOT have duplicate items (foo)")
			},
		},
		{
			name: "uniqueItems",
			s:    `["foo", "bar"]`,
			schema: &schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				UniqueItems: true,
			},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, i)
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
				v := i.([]interface{})[0]
				require.Equal(t, &struct {
					Name string `json:"name"`
					Age  int64  `json:"age"`
				}{Name: "foo", Age: 12}, v)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &schema.Ref{Value: tc.schema}
			i, err := r.Unmarshal([]byte(tc.s), media.ParseContentType("application/json"))
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
			schema: &schema.Schema{Type: "boolean"},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, true, i)
			},
		},
		{
			name:   "int to bool not possible",
			s:      `1`,
			schema: &schema.Schema{Type: "boolean"},
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "could not parse 1 as boolean, expected schema type=boolean")
			},
		},
		{
			name:   "false",
			s:      `false`,
			schema: &schema.Schema{Type: "boolean"},
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, false, i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.s, func(t *testing.T) {
			t.Parallel()

			r := &schema.Ref{Value: tc.schema}
			i, err := r.Unmarshal([]byte(tc.s), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}

func TestRef_Unmarshal_Json_Errors(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		err    error
	}{
		{
			name:   "empty json",
			s:      ``,
			schema: &schema.Schema{},
			err:    fmt.Errorf("invalid json format: unexpected end of JSON input"),
		},
		{
			name:   "invalid json",
			s:      `bar`,
			schema: &schema.Schema{Type: "string"},
			err:    fmt.Errorf("invalid json format: invalid character 'b' looking for beginning of value"),
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &schema.Ref{Value: tc.schema}
			_, err := r.Unmarshal([]byte(tc.s), media.ParseContentType("application/json"))
			require.Equal(t, tc.err, err)
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

			r := &schema.Ref{Value: tc.schema}
			_, err := r.Unmarshal([]byte(tc.s), media.ParseContentType("application/json"))
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
				require.Equal(t, &struct {
					Ship_date string `json:"ship-date"` // can be accessed via obj['ship-date'] in javascript
				}{Ship_date: "2022-01-01"}, i)
			},
		},
		{
			name: "kebab case conflict",
			s:    `{"ship-date":"2022-01-01","shipdate":"2023-01-01"}`,
			schema: schematest.New("object",
				schematest.WithProperty("ship-date", schematest.New("string")),
				schematest.WithProperty("shipdate", schematest.New("string"))),
			test: func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Ship_date string `json:"ship-date"`
					Shipdate  string `json:"shipdate"`
				}{Ship_date: "2022-01-01", Shipdate: "2023-01-01"}, i)
			},
		},
		{
			name: "- conflict",
			s:    `{"ship-date":"2022-01-01","ship_date":"2023-01-01"}`,
			schema: schematest.New("object",
				schematest.WithProperty("ship-date", schematest.New("string")),
				schematest.WithProperty("shipdate", schematest.New("string"))),
			test: func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "duplicate field name Ship_date")
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &schema.Ref{Value: tc.schema}
			i, err := r.Unmarshal([]byte(tc.s), media.ParseContentType("application/json"))
			tc.test(t, i, err)
		})
	}
}
