package schema_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/media"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	data := []struct {
		s      string
		schema *schema.Schema
		e      interface{}
	}{
		{
			`""`,
			nil,
			"",
		},
		{
			`{"foo": 12}`,
			nil,
			&struct {
				Foo float64
			}{Foo: 12},
		},
		{
			`[1, 2, 3, 4]`,
			nil,
			[]interface{}{float64(1), float64(2), float64(3), float64(4)},
		},
		{
			`{"foo": 12}`,
			schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
			&struct {
				Foo int64 `json:"foo"`
			}{Foo: 12},
		},
		{
			`{"foo": "bar"}`,
			schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
			&struct {
				Foo string `json:"foo"`
			}{Foo: "bar"},
		},
		{
			`{"foo": "2021-01-20"}`,
			schematest.New("object",
				schematest.WithProperty("foo",
					schematest.New("string", schematest.WithFormat("date")))),
			&struct {
				Foo string `json:"foo"`
			}{Foo: "2021-01-20"},
		},
		{
			`{"foo": ["a", "b", "c"]}`,
			schematest.New("object",
				schematest.WithProperty("foo",
					schematest.New("array", schematest.WithItems(schematest.New("string"))))),
			&struct {
				Foo []string `json:"foo"`
			}{Foo: []string{"a", "b", "c"}},
		},
		{
			`{"test": 12, "test2": true}`,
			schematest.New("object",
				schematest.Any(
					schematest.New("object", schematest.WithProperty("test", schematest.New("integer"))),
					schematest.New("object", schematest.WithProperty("test2", schematest.New("boolean"))),
				)),
			&struct {
				Test  int64 `json:"test"`
				Test2 bool  `json:"test2"`
			}{Test: int64(12), Test2: true},
		},
		{
			`"hello world"`,
			schematest.New("object",
				schematest.Any(
					schematest.New("object", schematest.WithProperty("test", schematest.New("integer"))),
					schematest.New("string"),
				)),
			"hello world",
		},
	}

	for _, d := range data {
		t.Run(d.s, func(t *testing.T) {
			i, err := schema.Parse([]byte(d.s), media.ParseContentType("application/json"), &schema.Ref{Value: d.schema})
			require.NoError(t, err)
			require.Equal(t, d.e, i)
		})
	}
}

func TestAny(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		schema *schema.Schema
		exp    interface{}
		err    error
	}{
		{
			"any",
			"12",
			schematest.New("",
				schematest.Any(
					schematest.New("string"),
					schematest.New("integer"))),
			int64(12),
			nil,
		},
		{
			"not match any",
			"12.6",
			schematest.New("",
				schematest.Any(
					schematest.New("string"),
					schematest.New("integer"))),
			nil,
			fmt.Errorf("value 12.6 does not match any of expected schema"),
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
			&struct {
				Foo string `json:"foo"`
			}{Foo: "bar"},
			nil,
		},
		{
			"too many properties",
			`{"name": "bar", "age": 12}`,
			schematest.New("",
				schematest.Any(
					schematest.New("object",
						schematest.WithProperty("name", schematest.New("string"))))),
			nil,
			fmt.Errorf("too many properties for object"),
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
			nil,
			fmt.Errorf("expected required property age"),
		},
		{
			"marge",
			`{"name": "bar", "age": 12}`,
			schematest.New("",
				schematest.Any(
					schematest.New("object",
						schematest.WithProperty("name", schematest.New("string"))),
					schematest.New("object",
						schematest.WithProperty("age", schematest.New("integer")),
						schematest.WithRequired("age"),
					))),
			&struct {
				Name string `json:"name"`
				Age  int64  `json:"age"`
			}{Name: "bar", Age: int64(12)},
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

func TestParseOneOf(t *testing.T) {
	data := []struct {
		s      string
		schema *schema.Schema
		e      interface{}
		err    error
	}{
		{
			`{"foo": true}`,
			schematest.New("", schematest.OneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("boolean"))),
			)),
			&struct {
				Foo bool `json:"foo"`
			}{Foo: true},
			nil,
		},
		{
			`{"foo": 12, "bar": true}`,
			schematest.New("", schematest.OneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("bar", schematest.New("boolean"))),
			)),
			nil,
			fmt.Errorf("value does not match any of oneof schema"),
		},
		{
			`{"foo": 12}`,
			schematest.New("", schematest.OneOf(
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("integer"))),
				schematest.New("object",
					schematest.WithProperty("foo", schematest.New("number"))),
			)),
			nil,
			fmt.Errorf("oneOf: given data is valid against more as one schema"),
		},
		{
			`"hello world"`,
			schematest.New("", schematest.OneOf(
				schematest.New("integer"),
				schematest.New("string"),
			)),
			"hello world",
			nil,
		},
	}

	for _, d := range data {
		t.Run(d.s, func(t *testing.T) {
			i, err := schema.Parse([]byte(d.s), media.ParseContentType("application/json"), &schema.Ref{Value: d.schema})
			require.Equal(t, d.err, err)
			require.Equal(t, d.e, i)
		})
	}
}

func TestParseAllOf(t *testing.T) {
	data := []struct {
		s      string
		schema *schema.Schema
		e      interface{}
	}{
		{
			`{"foo": 12, "bar": true}`,
			schematest.New("object",
				schematest.Any(
					schematest.New("object", schematest.WithProperty("foo", schematest.New("integer"))),
					schematest.New("object", schematest.WithProperty("bar", schematest.New("boolean"))),
				)),
			&struct {
				Foo int64 `json:"foo"`
				Bar bool  `json:"bar"`
			}{Foo: int64(12), Bar: true},
		},
	}

	for _, d := range data {
		t.Run(d.s, func(t *testing.T) {
			i, err := schema.Parse([]byte(d.s), media.ParseContentType("application/json"), &schema.Ref{Value: d.schema})
			require.NoError(t, err)
			require.Equal(t, d.e, i)
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
			fmt.Errorf("expected schema type=integer format=int32, got 3.61"),
		},
		{
			"not int32",
			fmt.Sprintf("%v", math.MaxInt64),
			&schema.Schema{Type: "integer", Format: "int32"},
			0,
			fmt.Errorf("integer is not int32"),
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
			fmt.Errorf("12 is lower as the expected minimum 13"),
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
			fmt.Errorf("12 is greater as the expected maximum 5"),
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
			fmt.Errorf("expected schema type=number format=float, got 1.7976931348623157e+308"),
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
			fmt.Errorf("3.612 is lower as the expected minimum 3.7"),
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
			fmt.Errorf("3.612 is greater as the expected maximum 3.6"),
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
			fmt.Errorf("value does not match pattern"),
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
			fmt.Errorf("string is not a date RFC3339"),
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
			fmt.Errorf("string is not a date-time RFC3339"),
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
			fmt.Errorf("string is not an email address"),
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
			fmt.Errorf("string is not an uuid"),
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
			fmt.Errorf("string is not an ipv4"),
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
			fmt.Errorf("string is not an ipv6"),
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
				require.EqualError(t, err, "expected schema type=object required=[name age], got {name=foo}")
			},
		},
		{
			"not minProperties",
			`{"name": "foo"}`,
			schematest.New("object",
				schematest.WithMinProperties(2),
			),
			func(t *testing.T, _ interface{}, err error) {
				require.EqualError(t, err, "expected schema type=object minProperties=2, got {name=foo}")
			},
		},
		{
			"not maxProperties",
			`{"name": "foo", "age": 12}`,
			schematest.New("object",
				schematest.WithMaxProperties(1),
			),
			func(t *testing.T, _ interface{}, err error) {
				switch err.Error() {
				case "expected schema type=object maxProperties=1, got {age=12, name=foo}":
				case "expected schema type=object maxProperties=1, got {name=foo, age=12}":
					return
				default:
					require.Error(t, err)
					require.Failf(t, "got wrong error message", err.Error())
				}
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
				if !reflect.DeepEqual(&struct {
					Name string
					Age  float64
				}{Name: "foo", Age: 12}, v) && !reflect.DeepEqual(&struct {
					Age  float64
					Name string
				}{Name: "foo", Age: 12}, v) {
					require.Failf(t, "actual object is not an expected one", fmt.Sprintf("%v", v))
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
				v := reflect.ValueOf(i)
				require.Equal(t, 1, v.Len())
				require.Equal(t, &struct {
					Name string `json:"name"`
					Age  int64  `json:"age"`
				}{Name: "foo", Age: 12}, v.Index(0).Interface())
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
			fmt.Errorf("expected bool but got float64"),
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
