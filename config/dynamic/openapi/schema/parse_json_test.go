package schema_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/media"
	"mokapi/sortedmap"
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	testcases := []struct {
		s      string
		schema *schema.Schema
		f      func(t *testing.T, i interface{}, err error)
	}{
		{
			`""`,
			nil,
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "", i)
			},
		},
		{
			`""`,
			schematest.New(""),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "", i)
			},
		},
		{
			`{"foo": 12}`,
			nil,
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Foo float64 `json:"foo"`
				}{Foo: 12}, i)
			},
		},
		{
			`[1, 2, 3, 4]`,
			nil,
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{float64(1), float64(2), float64(3), float64(4)}, i)
			},
		},
		{
			`{"foo": 12}`,
			schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Foo int64 `json:"foo"`
				}{Foo: int64(12)}, i)
			},
		},
		{
			`{"foo": "bar"}`,
			schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Foo string `json:"foo"`
				}{Foo: "bar"}, i)
			},
		},
		{
			`{"foo": "2021-01-20"}`,
			schematest.New("object",
				schematest.WithProperty("foo",
					schematest.New("string", schematest.WithFormat("date")))),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Foo string `json:"foo"`
				}{Foo: "2021-01-20"}, i)
			},
		},
		{
			`{"foo": ["a", "b", "c"]}`,
			schematest.New("object",
				schematest.WithProperty("foo",
					schematest.New("array", schematest.WithItems(schematest.New("string"))))),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Foo []string `json:"foo"`
				}{Foo: []string{"a", "b", "c"}}, i)
			},
		},
		{
			`{"test": 12, "test2": true}`,
			schematest.New("object",
				schematest.Any(
					schematest.New("object", schematest.WithProperty("test", schematest.New("integer"))),
					schematest.New("object", schematest.WithProperty("test2", schematest.New("boolean"))),
				)),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Test  int64 `json:"test"`
					Test2 bool  `json:"test2"`
				}{Test: 12, Test2: true}, i)
			},
		},
		{
			`"hello world"`,
			schematest.New("object",
				schematest.Any(
					schematest.New("object", schematest.WithProperty("test", schematest.New("integer"))),
					schematest.New("string"),
				)),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "hello world", i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.s, func(t *testing.T) {
			t.Parallel()
			i, err := schema.Parse([]byte(tc.s), media.ParseContentType("application/json"), &schema.Ref{Value: tc.schema})
			tc.f(t, i, err)
		})
	}
}

func TestAny(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		f      func(t *testing.T, i interface{}, err error)
	}{
		{
			"any",
			"12",
			schematest.New("",
				schematest.Any(
					schematest.New("string"),
					schematest.New("integer"))),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(12), i)
			},
		},
		{
			"not match any",
			"12.6",
			schematest.New("",
				schematest.Any(
					schematest.New("string"),
					schematest.New("integer"))),
			func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "could not parse 12.6, expected any of schema type=string, schema type=integer")
			},
		},
		{
			"any object",
			`{"foo": "bar"}`,
			schematest.New("",
				schematest.Any(
					schematest.New("object",
						schematest.WithProperty("foo", schematest.New("integer"))),
					schematest.New("object",
						schematest.WithProperty("foo", schematest.New("string"))))),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Foo string `json:"foo"`
				}{Foo: "bar"}, i)
			},
		},
		{
			"too many properties",
			`{"name": "bar", "age": 12}`,
			schematest.New("",
				schematest.Any(
					schematest.New("object",
						schematest.WithProperty("name", schematest.New("string"))))),
			func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, `could not parse {"age":12,"name":"bar"}, too many properties for object, expected any of schema type=object properties=[name]`)
			},
		},
		{
			"missing required property",
			`{"name": "bar"}`,
			schematest.New("",
				schematest.Any(
					schematest.New("object",
						schematest.WithProperty("name", schematest.New("string"))),
					schematest.New("object",
						schematest.WithProperty("age", schematest.New("integer")),
						schematest.WithRequired("age"),
					))),
			func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "missing required property age, expected any of schema type=object properties=[name], schema type=object properties=[age] required=[age]")
			},
		},
		{
			"merge",
			`{"name": "bar", "age": 12}`,
			schematest.New("",
				schematest.Any(
					schematest.New("object",
						schematest.WithProperty("name", schematest.New("string"))),
					schematest.New("object",
						schematest.WithProperty("age", schematest.New("integer")),
						schematest.WithRequired("age"),
					))),
			func(t *testing.T, i interface{}, err error) {
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
			i, err := schema.Parse([]byte(tc.s), media.ParseContentType("application/json"), &schema.Ref{Value: tc.schema})
			tc.f(t, i, err)
		})
	}
}

func TestParseOneOf(t *testing.T) {
	testcases := []struct {
		s      string
		schema *schema.Schema
		f      func(t *testing.T, i interface{}, err error)
	}{
		{
			`{"foo": true}`,
			schematest.New("", schematest.OneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("boolean"))),
			)),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Foo bool `json:"foo"`
				}{Foo: true}, i)
			},
		},
		{
			`{"foo": 12, "bar": true}`,
			schematest.New("", schematest.OneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("bar", schematest.New("boolean"))),
			)),
			func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, `could not parse {"bar":true,"foo":12}, expected one of schema type=object properties=[foo], schema type=object properties=[bar]`)
			},
		},
		{
			`{"foo": 12}`,
			schematest.New("", schematest.OneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("number"))),
			)),
			func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, `could not parse {"foo":12}, it is not valid for only one schema, expected one of schema type=object properties=[foo], schema type=object properties=[foo]`)
			},
		},
		{
			`"hello world"`,
			schematest.New("", schematest.OneOf(
				schematest.New("integer"),
				schematest.New("string"),
			)),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "hello world", i)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.s, func(t *testing.T) {
			t.Parallel()
			i, err := schema.Parse([]byte(tc.s), media.ParseContentType("application/json"), &schema.Ref{Value: tc.schema})
			tc.f(t, i, err)
		})
	}
}

func TestParseAllOf(t *testing.T) {
	testcases := []struct {
		s      string
		schema *schema.Schema
		f      func(t *testing.T, i interface{}, err error)
	}{
		{
			`{"foo": 12, "bar": true}`,
			schematest.New("object",
				schematest.AllOf(
					schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
					schematest.New("object", schematest.WithProperty("bar", schematest.New("boolean"))),
				)),
			func(t *testing.T, i interface{}, err error) {
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
			i, err := schema.Parse([]byte(tc.s), media.ParseContentType("application/json"), &schema.Ref{Value: tc.schema})
			tc.f(t, i, err)
		})
	}
}

func TestParse_Integer(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		schema *schema.Schema
		exp    int
		err    error
	}{
		{
			"int32",
			"12",
			&schema.Schema{Type: "integer", Format: "int32"},
			12,
			nil,
		},
		{
			"not int",
			"3.61",
			&schema.Schema{Type: "integer", Format: "int32"},
			0,
			fmt.Errorf("could not parse 3.61 as integer, expected schema type=integer format=int32"),
		},
		{
			"not int32",
			fmt.Sprintf("%v", math.MaxInt64),
			&schema.Schema{Type: "integer", Format: "int32"},
			0,
			fmt.Errorf("could not parse '9.223372036854776e+18', represents a number either less than int32 min value or greater max value, expected schema type=integer format=int32"),
		},
		{
			"min",
			"12",
			&schema.Schema{Type: "integer", Minimum: toFloatP(5)},
			12,
			nil,
		},
		{
			"not min",
			"12",
			&schema.Schema{Type: "integer", Minimum: toFloatP(13)},
			0,
			fmt.Errorf("12 is lower as the required minimum 13, expected schema type=integer minimum=13"),
		},
		{
			"max",
			"12",
			&schema.Schema{Type: "integer", Maximum: toFloatP(13)},
			12,
			nil,
		},
		{
			"not max",
			"12",
			&schema.Schema{Type: "integer", Maximum: toFloatP(5)},
			0,
			fmt.Errorf("12 is greater as the required maximum 5, expected schema type=integer maximum=5"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			i, err := schema.Parse([]byte(d.s), media.ParseContentType("application/json"), &schema.Ref{Value: d.schema})
			if d.err != nil {
				require.Equal(t, d.err, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, int64(d.exp), i)
			}
		})
	}
}

func TestParse_Number(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		schema *schema.Schema
		exp    float64
		err    error
	}{
		{
			"float32",
			"3.612",
			&schema.Schema{Type: "number", Format: "float32"},
			3.612,
			nil,
		},
		{
			"not float",
			fmt.Sprintf("%v", math.MaxFloat64),
			&schema.Schema{Type: "number", Format: "float"},
			0,
			fmt.Errorf("could not parse 1.7976931348623157e+308 as float, expected schema type=number format=float"),
		},
		{
			"min",
			"3.612",
			&schema.Schema{Type: "number", Minimum: toFloatP(3.6)},
			3.612,
			nil,
		},
		{
			"not min",
			"3.612",
			&schema.Schema{Type: "number", Minimum: toFloatP(3.7)},
			0,
			fmt.Errorf("3.612 is lower as the required minimum 3.7, expected schema type=number minimum=3.7"),
		},
		{
			"max",
			"3.612",
			&schema.Schema{Type: "number", Maximum: toFloatP(3.7)},
			3.612,
			nil,
		},
		{
			"not max",
			"3.612",
			&schema.Schema{Type: "number", Maximum: toFloatP(3.6)},
			0,
			fmt.Errorf("3.612 is greater as the required maximum 3.6, expected schema type=number maximum=3.6"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			i, err := schema.Parse([]byte(d.s), media.ParseContentType("application/json"), &schema.Ref{Value: d.schema})
			if d.err != nil {
				require.Equal(t, d.err, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, d.exp, i)
			}
		})
	}
}

func TestParse_String(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		schema *schema.Schema
		err    error
	}{
		{
			"string",
			`"gbRMaRxHkiJBPta"`,
			&schema.Schema{Type: "string"},
			nil,
		},
		{
			"by pattern",
			`"013-64-5994"`,
			&schema.Schema{Type: "string", Pattern: "^\\d{3}-\\d{2}-\\d{4}$"},
			nil,
		},
		{
			"not pattern",
			`"013-64-59943"`,
			&schema.Schema{Type: "string", Pattern: "^\\d{3}-\\d{2}-\\d{4}$"},
			fmt.Errorf("value '013-64-59943' does not match pattern, expected schema type=string pattern=^\\d{3}-\\d{2}-\\d{4}$"),
		},
		{
			"date",
			`"1908-12-07"`,
			&schema.Schema{Type: "string", Format: "date"},
			nil,
		},
		{
			"not date",
			`"1908-12-7"`,
			&schema.Schema{Type: "string", Format: "date"},
			fmt.Errorf("value '1908-12-7' is not a date RFC3339, expected schema type=string format=date"),
		},
		{
			"date-time",
			`"1908-12-07T04:14:25Z"`,
			&schema.Schema{Type: "string", Format: "date-time"},
			nil,
		},
		{
			"not date-time",
			`"1908-12-07 T04:14:25Z"`,
			&schema.Schema{Type: "string", Format: "date-time"},
			fmt.Errorf("value '1908-12-07 T04:14:25Z' is not a date-time RFC3339, expected schema type=string format=date-time"),
		},
		{
			"password",
			`"H|$9lb{J<+S;"`,
			&schema.Schema{Type: "string", Format: "password"},
			nil,
		},
		{
			"email",
			`"markusmoen@pagac.net"`,
			&schema.Schema{Type: "string", Format: "email"},
			nil,
		},
		{
			"not email",
			`"markusmoen@@pagac.net"`,
			&schema.Schema{Type: "string", Format: "email"},
			fmt.Errorf("value 'markusmoen@@pagac.net' is not an email address, expected schema type=string format=email"),
		},
		{
			"uuid",
			`"590c1440-9888-45b0-bd51-a817ee07c3f2"`,
			&schema.Schema{Type: "string", Format: "uuid"},
			nil,
		},
		{
			"not uuid",
			`"590c1440-9888-45b0-bd51-a817ee07c3f2a"`,
			&schema.Schema{Type: "string", Format: "uuid"},
			fmt.Errorf("value '590c1440-9888-45b0-bd51-a817ee07c3f2a' is not an uuid, expected schema type=string format=uuid"),
		},
		{
			"ipv4",
			`"152.23.53.100"`,
			&schema.Schema{Type: "string", Format: "ipv4"},
			nil,
		},
		{
			"not ipv4",
			`"152.23.53.100."`,
			&schema.Schema{Type: "string", Format: "ipv4"},
			fmt.Errorf("value '152.23.53.100.' is not an ipv4, expected schema type=string format=ipv4"),
		},
		{
			"ipv6",
			`"8898:ee17:bc35:9064:5866:d019:3b95:7857"`,
			&schema.Schema{Type: "string", Format: "ipv6"},
			nil,
		},
		{
			"not ipv6",
			`"-8898:ee17:bc35:9064:5866:d019:3b95:7857"`,
			&schema.Schema{Type: "string", Format: "ipv6"},
			fmt.Errorf("value '-8898:ee17:bc35:9064:5866:d019:3b95:7857' is not an ipv6, expected schema type=string format=ipv6"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			i, err := schema.Parse([]byte(d.s), media.ParseContentType("application/json"), &schema.Ref{Value: d.schema})
			require.Equal(t, d.err, err)
			require.Equal(t, d.s[1:len(d.s)-1], i)
		})
	}
}

func TestValidate_Object(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		fn     func(t *testing.T, v interface{}, err error)
	}{
		{
			"empty",
			"{}",
			&schema.Schema{Type: "object"},
			func(t *testing.T, v interface{}, _ error) {
				require.Equal(t, &struct{}{}, v)
			},
		},
		{
			"simple",
			`{"name": "foo", "age": 12}`,
			schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("age", schematest.New("integer")),
			),
			func(t *testing.T, v interface{}, _ error) {
				require.Equal(t, &struct {
					Name string `json:"name"`
					Age  int64  `json:"age"`
				}{Name: "foo", Age: 12}, v)
			},
		},
		{
			"missing but not required",
			`{"name": "foo"}`,
			schematest.New("object",
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("age", schematest.New("integer")),
			),
			func(t *testing.T, v interface{}, _ error) {
				require.Equal(t, &struct {
					Name string `json:"name"`
				}{Name: "foo"}, v)
			},
		},
		{
			"missing but required",
			`{"name": "foo"}`,
			schematest.New("object",
				schematest.WithRequired("name", "age"),
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithProperty("age", schematest.New("integer")),
			),
			func(t *testing.T, _ interface{}, err error) {
				require.EqualError(t, err, `missing required field age on {"name":"foo"}, expected schema type=object properties=[name, age] required=[name age]`)
			},
		},
		{
			"not minProperties",
			`{"name": "foo"}`,
			schematest.New("object",
				schematest.WithMinProperties(2),
			),
			func(t *testing.T, _ interface{}, err error) {
				require.EqualError(t, err, `validation error minProperties on {"name":"foo"}, expected schema type=object minProperties=2`)
			},
		},
		{
			"not maxProperties",
			`{"name": "foo", "age": 12}`,
			schematest.New("object",
				schematest.WithMaxProperties(1),
			),
			func(t *testing.T, _ interface{}, err error) {
				require.Truef(t, strings.HasSuffix(err.Error(), "expected schema type=object maxProperties=1"), "but error is %v", err)
			},
		},
		{
			"min-max properties, free-form order of properties not defined",
			`{"name": "foo", "age": 12}`,
			schematest.New("object",
				schematest.WithMinProperties(1),
				schematest.WithMaxProperties(2),
			),
			func(t *testing.T, v interface{}, _ error) {
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
			i, err := schema.Parse([]byte(tc.s), media.ParseContentType("application/json"), &schema.Ref{Value: tc.schema})
			tc.fn(t, i, err)
		})
	}
}

func TestValidate_Array(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		schema *schema.Schema
		fn     func(t *testing.T, i interface{}, err error)
	}{
		{
			"empty",
			"[]",
			&schema.Schema{Type: "array", Items: &schema.Ref{
				Value: &schema.Schema{Type: "string"},
			}},
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{}, i)
			},
		},
		{
			"string array",
			`["foo", "bar"]`,
			&schema.Schema{Type: "array", Items: &schema.Ref{
				Value: &schema.Schema{Type: "string"},
			}},
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, i)
			},
		},
		{
			"object array",
			`[{"name": "foo", "age": 12}]`,
			schematest.New("array", schematest.WithItems(
				schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("age", schematest.New("integer")),
				),
			),
			),
			func(t *testing.T, i interface{}, err error) {
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
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			i, err := schema.Parse([]byte(d.s), media.ParseContentType("application/json"), &schema.Ref{Value: d.schema})
			d.fn(t, i, err)
		})
	}
}

func TestParse_Bool(t *testing.T) {
	cases := []struct {
		s      string
		schema *schema.Schema
		exp    bool
		err    error
	}{
		{
			`true`,
			&schema.Schema{Type: "boolean"},
			true,
			nil,
		},
		{
			`1`,
			&schema.Schema{Type: "boolean"},
			false,
			fmt.Errorf("could not parse 1 as boolean, expected schema type=boolean"),
		},
		{
			`false`,
			&schema.Schema{Type: "boolean"},
			false,
			nil,
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.s, func(t *testing.T) {
			t.Parallel()
			i, err := schema.Parse([]byte(d.s), media.ParseContentType("application/json"), &schema.Ref{Value: d.schema})
			require.Equal(t, d.err, err)
			require.Equal(t, d.exp, i)
		})
	}
}

func TestParse_Errors(t *testing.T) {
	cases := []struct {
		s      string
		schema *schema.Schema
		err    error
	}{
		{
			``,
			&schema.Schema{},
			fmt.Errorf("invalid json format: unexpected end of JSON input"),
		},
		{
			`bar`,
			&schema.Schema{Type: "string"},
			fmt.Errorf("invalid json format: invalid character 'b' looking for beginning of value"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.s, func(t *testing.T) {
			t.Parallel()
			_, err := schema.Parse([]byte(d.s), media.ParseContentType("application/json"), &schema.Ref{Value: d.schema})
			require.Equal(t, d.err, err)
		})
	}
}

func TestValidate_Dictionary(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		schema *schema.Schema
		exp    interface{}
		err    error
	}{
		{
			"string:string",
			`{"en": "English", "de": "German"}`,
			schematest.New("object", schematest.WithAdditionalProperties(schematest.New("string"))),
			map[string]interface{}{"en": "English", "de": "German"},
			nil,
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			i, err := schema.Parse([]byte(d.s), media.ParseContentType("application/json"), &schema.Ref{Value: d.schema})
			if d.err != nil {
				require.Equal(t, d.err, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, d.exp, i)
			}
		})
	}
}

func TestMarshal_SortedMap(t *testing.T) {
	cases := []struct {
		name   string
		schema *schema.Schema
		f      func(t *testing.T, s *schema.Ref)
	}{
		{
			"string:string",
			schematest.New("object", schematest.WithProperty("number", schematest.New("number"))),
			func(t *testing.T, s *schema.Ref) {
				m := sortedmap.NewLinkedHashMap()
				m.Set("number", 12)

				b, err := s.Marshal(m, media.ParseContentType("application/json"))
				require.NoError(t, err)
				require.Equal(t, `{"number":12}`, string(b))
			},
		},
	}

	t.Parallel()
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t, &schema.Ref{Value: tc.schema})
		})
	}
}
