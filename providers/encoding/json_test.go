package encoding

import (
	"fmt"
	"math"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/models/media"
	"mokapi/test"
	"reflect"
	"testing"
)

func toFloatP(f float64) *float64 { return &f }

func TestParse(t *testing.T) {
	data := []struct {
		s      string
		schema *openapi.Schema
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
				foo int
			}{foo: 12},
		},
		{
			`[1, 2, 3, 4]`,
			nil,
			[]interface{}{float64(1), float64(2), float64(3), float64(4)},
		},
		{
			`{"foo": 12}`,
			openapitest.NewSchema("object", openapitest.WithProperty("foo", openapitest.NewSchema("integer"))),
			&struct {
				foo int
			}{foo: 12},
		},
		{
			`{"foo": "bar"}`,
			openapitest.NewSchema("object", openapitest.WithProperty("foo", openapitest.NewSchema("string"))),
			&struct {
				foo string
			}{foo: "bar"},
		},
		{
			`{"foo": "2021-01-20"}`,
			openapitest.NewSchema("object",
				openapitest.WithProperty("foo",
					openapitest.NewSchema("string", openapitest.WithFormat("date")))),
			&struct {
				foo string
			}{foo: "2021-01-20"},
		},
		{
			`{"foo": ["a", "b", "c"]}`,
			openapitest.NewSchema("object",
				openapitest.WithProperty("foo",
					openapitest.NewSchema("array", openapitest.WithItems(openapitest.NewSchema("string"))))),
			&struct {
				foo []string
			}{foo: []string{"a", "b", "c"}},
		},
		{
			`{"test": 12, "test2": true}`,
			openapitest.NewSchema("object",
				openapitest.Any(
					openapitest.NewSchema("object", openapitest.WithProperty("test", openapitest.NewSchema("integer"))),
					openapitest.NewSchema("object", openapitest.WithProperty("test2", openapitest.NewSchema("boolean"))),
				)),
			&struct {
				test  int64
				test2 bool
			}{test: int64(12), test2: true},
		},
		{
			`"hello world"`,
			openapitest.NewSchema("object",
				openapitest.Any(
					openapitest.NewSchema("object", openapitest.WithProperty("test", openapitest.NewSchema("integer"))),
					openapitest.NewSchema("string"),
				)),
			"hello world",
		},
	}

	for _, d := range data {
		t.Run(d.s, func(t *testing.T) {
			i, err := Parse([]byte(d.s), media.ParseContentType("application/json"), &openapi.SchemaRef{Value: d.schema})
			test.Ok(t, err)
			test.Equals(t, d.e, i)
		})
	}
}

func TestAny(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		schema *openapi.Schema
		exp    interface{}
		err    error
	}{
		{
			"any",
			"12",
			openapitest.NewSchema("",
				openapitest.Any(
					openapitest.NewSchema("string"),
					openapitest.NewSchema("integer"))),
			int64(12),
			nil,
		},
		{
			"not match any",
			"12.6",
			openapitest.NewSchema("",
				openapitest.Any(
					openapitest.NewSchema("string"),
					openapitest.NewSchema("integer"))),
			nil,
			fmt.Errorf("value 12.6 does not match any of expected schema"),
		},
		{
			"any object",
			`{"foo": "bar"}`,
			openapitest.NewSchema("",
				openapitest.Any(
					openapitest.NewSchema("object",
						openapitest.WithProperty("foo", openapitest.NewSchema("integer"))),
					openapitest.NewSchema("object",
						openapitest.WithProperty("foo", openapitest.NewSchema("string"))))),
			&struct {
				foo string
			}{foo: "bar"},
			nil,
		},
		{
			"too many properties",
			`{"name": "bar", "age": 12}`,
			openapitest.NewSchema("",
				openapitest.Any(
					openapitest.NewSchema("object",
						openapitest.WithProperty("name", openapitest.NewSchema("string"))))),
			nil,
			fmt.Errorf("too many properties for object"),
		},
		{
			"missing required property",
			`{"name": "bar"}`,
			openapitest.NewSchema("",
				openapitest.Any(
					openapitest.NewSchema("object",
						openapitest.WithProperty("name", openapitest.NewSchema("string"))),
					openapitest.NewSchema("object",
						openapitest.WithProperty("age", openapitest.NewSchema("integer")),
						openapitest.WithRequired("age"),
					))),
			nil,
			fmt.Errorf("expected required property age"),
		},
		{
			"marge",
			`{"name": "bar", "age": 12}`,
			openapitest.NewSchema("",
				openapitest.Any(
					openapitest.NewSchema("object",
						openapitest.WithProperty("name", openapitest.NewSchema("string"))),
					openapitest.NewSchema("object",
						openapitest.WithProperty("age", openapitest.NewSchema("integer")),
						openapitest.WithRequired("age"),
					))),
			&struct {
				name string
				age  int64
			}{name: "bar", age: int64(12)},
			nil,
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			i, err := Parse([]byte(d.s), media.ParseContentType("application/json"), &openapi.SchemaRef{Value: d.schema})
			if d.err != nil {
				test.Equals(t, d.err, err)
			} else {
				test.Ok(t, err)
				test.Equals(t, d.exp, i)
			}
		})
	}
}

func TestParseOneOf(t *testing.T) {
	data := []struct {
		s      string
		schema *openapi.Schema
		e      interface{}
		err    error
	}{
		{
			`{"foo": true}`,
			openapitest.NewSchema("", openapitest.OneOf(
				openapitest.NewSchema("object",
					openapitest.WithProperty("foo", openapitest.NewSchema("integer"))),
				openapitest.NewSchema("object",
					openapitest.WithProperty("foo", openapitest.NewSchema("boolean"))),
			)),
			&struct {
				foo bool
			}{foo: true},
			nil,
		},
		{
			`{"foo": 12, "bar": true}`,
			openapitest.NewSchema("", openapitest.OneOf(
				openapitest.NewSchema("object",
					openapitest.WithProperty("foo", openapitest.NewSchema("integer"))),
				openapitest.NewSchema("object",
					openapitest.WithProperty("bar", openapitest.NewSchema("boolean"))),
			)),
			nil,
			fmt.Errorf("value does not match any of oneof schema"),
		},
		{
			`{"foo": 12}`,
			openapitest.NewSchema("", openapitest.OneOf(
				openapitest.NewSchema("object",
					openapitest.WithProperty("foo", openapitest.NewSchema("integer"))),
				openapitest.NewSchema("object",
					openapitest.WithProperty("foo", openapitest.NewSchema("number"))),
			)),
			nil,
			fmt.Errorf("oneOf: given data is valid against more as one schema"),
		},
		{
			`"hello world"`,
			openapitest.NewSchema("", openapitest.OneOf(
				openapitest.NewSchema("integer"),
				openapitest.NewSchema("string"),
			)),
			"hello world",
			nil,
		},
	}

	for _, d := range data {
		t.Run(d.s, func(t *testing.T) {
			i, err := Parse([]byte(d.s), media.ParseContentType("application/json"), &openapi.SchemaRef{Value: d.schema})
			test.Equals(t, d.err, err)
			test.Equals(t, d.e, i)
		})
	}
}

func TestParseAllOf(t *testing.T) {
	data := []struct {
		s      string
		schema *openapi.Schema
		e      interface{}
	}{
		{
			`{"foo": 12, "bar": true}`,
			openapitest.NewSchema("object",
				openapitest.Any(
					openapitest.NewSchema("object", openapitest.WithProperty("foo", openapitest.NewSchema("integer"))),
					openapitest.NewSchema("object", openapitest.WithProperty("bar", openapitest.NewSchema("boolean"))),
				)),
			&struct {
				foo int64
				bar bool
			}{foo: int64(12), bar: true},
		},
	}

	for _, d := range data {
		t.Run(d.s, func(t *testing.T) {
			i, err := Parse([]byte(d.s), media.ParseContentType("application/json"), &openapi.SchemaRef{Value: d.schema})
			test.Ok(t, err)
			test.Equals(t, d.e, i)
		})
	}
}

func TestParse_Integer(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		schema *openapi.Schema
		exp    int
		err    error
	}{
		{
			"int32",
			"12",
			&openapi.Schema{Type: "integer", Format: "int32"},
			12,
			nil,
		},
		{
			"not int",
			"3.61",
			&openapi.Schema{Type: "integer", Format: "int32"},
			0,
			fmt.Errorf("expected integer but got floating number"),
		},
		{
			"not int32",
			fmt.Sprintf("%v", math.MaxInt64),
			&openapi.Schema{Type: "integer", Format: "int32"},
			0,
			fmt.Errorf("integer is not int32"),
		},
		{
			"min",
			"12",
			&openapi.Schema{Type: "integer", Minimum: toFloatP(5)},
			12,
			nil,
		},
		{
			"not min",
			"12",
			&openapi.Schema{Type: "integer", Minimum: toFloatP(13)},
			0,
			fmt.Errorf("12 is lower as the expected minimum 13"),
		},
		{
			"max",
			"12",
			&openapi.Schema{Type: "integer", Maximum: toFloatP(13)},
			12,
			nil,
		},
		{
			"not max",
			"12",
			&openapi.Schema{Type: "integer", Maximum: toFloatP(5)},
			0,
			fmt.Errorf("12 is greater as the expected maximum 5"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			i, err := Parse([]byte(d.s), media.ParseContentType("application/json"), &openapi.SchemaRef{Value: d.schema})
			if d.err != nil {
				test.Equals(t, d.err, err)
			} else {
				test.Ok(t, err)
				test.Equals(t, int64(d.exp), i)
			}
		})
	}
}

func TestParse_Number(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		schema *openapi.Schema
		exp    float64
		err    error
	}{
		{
			"float32",
			"3.612",
			&openapi.Schema{Type: "number", Format: "float32"},
			3.612,
			nil,
		},
		{
			"not float32",
			fmt.Sprintf("%v", math.MaxFloat64),
			&openapi.Schema{Type: "number", Format: "float32"},
			0,
			fmt.Errorf("number is not float32"),
		},
		{
			"min",
			"3.612",
			&openapi.Schema{Type: "number", Minimum: toFloatP(3.6)},
			3.612,
			nil,
		},
		{
			"not min",
			"3.612",
			&openapi.Schema{Type: "number", Minimum: toFloatP(3.7)},
			0,
			fmt.Errorf("3.612 is lower as the expected minimum 3.7"),
		},
		{
			"max",
			"3.612",
			&openapi.Schema{Type: "number", Maximum: toFloatP(3.7)},
			3.612,
			nil,
		},
		{
			"not max",
			"3.612",
			&openapi.Schema{Type: "number", Maximum: toFloatP(3.6)},
			0,
			fmt.Errorf("3.612 is greater as the expected maximum 3.6"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			i, err := Parse([]byte(d.s), media.ParseContentType("application/json"), &openapi.SchemaRef{Value: d.schema})
			if d.err != nil {
				test.Equals(t, d.err, err)
			} else {
				test.Ok(t, err)
				test.Equals(t, d.exp, i)
			}
		})
	}
}

func TestParse_String(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		schema *openapi.Schema
		err    error
	}{
		{
			"string",
			`"gbRMaRxHkiJBPta"`,
			&openapi.Schema{Type: "string"},
			nil,
		},
		{
			"by pattern",
			`"013-64-5994"`,
			&openapi.Schema{Type: "string", Pattern: "^\\d{3}-\\d{2}-\\d{4}$"},
			nil,
		},
		{
			"not pattern",
			`"013-64-59943"`,
			&openapi.Schema{Type: "string", Pattern: "^\\d{3}-\\d{2}-\\d{4}$"},
			fmt.Errorf("value does not match pattern"),
		},
		{
			"date",
			`"1908-12-07"`,
			&openapi.Schema{Type: "string", Format: "date"},
			nil,
		},
		{
			"not date",
			`"1908-12-7"`,
			&openapi.Schema{Type: "string", Format: "date"},
			fmt.Errorf("string is not a date RFC3339"),
		},
		{
			"date-time",
			`"1908-12-07T04:14:25Z"`,
			&openapi.Schema{Type: "string", Format: "date-time"},
			nil,
		},
		{
			"not date-time",
			`"1908-12-07 T04:14:25Z"`,
			&openapi.Schema{Type: "string", Format: "date-time"},
			fmt.Errorf("string is not a date-time RFC3339"),
		},
		{
			"password",
			`"H|$9lb{J<+S;"`,
			&openapi.Schema{Type: "string", Format: "password"},
			nil,
		},
		{
			"email",
			`"markusmoen@pagac.net"`,
			&openapi.Schema{Type: "string", Format: "email"},
			nil,
		},
		{
			"not email",
			`"markusmoen@@pagac.net"`,
			&openapi.Schema{Type: "string", Format: "email"},
			fmt.Errorf("string is not an email address"),
		},
		{
			"uuid",
			`"590c1440-9888-45b0-bd51-a817ee07c3f2"`,
			&openapi.Schema{Type: "string", Format: "uuid"},
			nil,
		},
		{
			"not uuid",
			`"590c1440-9888-45b0-bd51-a817ee07c3f2a"`,
			&openapi.Schema{Type: "string", Format: "uuid"},
			fmt.Errorf("string is not an uuid"),
		},
		{
			"ipv4",
			`"152.23.53.100"`,
			&openapi.Schema{Type: "string", Format: "ipv4"},
			nil,
		},
		{
			"not ipv4",
			`"152.23.53.100."`,
			&openapi.Schema{Type: "string", Format: "ipv4"},
			fmt.Errorf("string is not an ipv4"),
		},
		{
			"ipv6",
			`"8898:ee17:bc35:9064:5866:d019:3b95:7857"`,
			&openapi.Schema{Type: "string", Format: "ipv6"},
			nil,
		},
		{
			"not ipv6",
			`"-8898:ee17:bc35:9064:5866:d019:3b95:7857"`,
			&openapi.Schema{Type: "string", Format: "ipv6"},
			fmt.Errorf("string is not an ipv6"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			i, err := Parse([]byte(d.s), media.ParseContentType("application/json"), &openapi.SchemaRef{Value: d.schema})
			test.Equals(t, d.err, err)
			test.Equals(t, d.s[1:len(d.s)-1], i)
		})
	}
}

func TestValidate_Object(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		schema *openapi.Schema
		exp    interface{}
		err    error
	}{
		{
			"empty",
			"{}",
			&openapi.Schema{Type: "object"},
			&struct{}{},
			nil,
		},
		{
			"simple",
			`{"name": "foo", "age": 12}`,
			openapitest.NewSchema("object",
				openapitest.WithProperty("name", openapitest.NewSchema("string")),
				openapitest.WithProperty("age", openapitest.NewSchema("integer")),
			),
			&struct {
				Name string
				Age  int64
			}{Name: "foo", Age: 12},
			nil,
		},
		{
			"missing but not required",
			`{"name": "foo"}`,
			openapitest.NewSchema("object",
				openapitest.WithProperty("name", openapitest.NewSchema("string")),
				openapitest.WithProperty("age", openapitest.NewSchema("integer")),
			),
			&struct {
				Name string
			}{Name: "foo"},
			nil,
		},
		{
			"missing but required",
			`{"name": "foo"}`,
			openapitest.NewSchema("object",
				openapitest.WithRequired("name", "age"),
				openapitest.WithProperty("name", openapitest.NewSchema("string")),
				openapitest.WithProperty("age", openapitest.NewSchema("integer")),
			),
			nil,
			fmt.Errorf("expected required property age"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			i, err := Parse([]byte(d.s), media.ParseContentType("application/json"), &openapi.SchemaRef{Value: d.schema})
			if d.err != nil {
				test.Equals(t, d.err, err)
			} else {
				test.Ok(t, err)
				test.Equals(t, d.exp, i)
			}
		})
	}
}

func TestValidate_Array(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		schema *openapi.Schema
		fn     func(t *testing.T, i interface{}, err error)
	}{
		{
			"empty",
			"[]",
			&openapi.Schema{Type: "array", Items: &openapi.SchemaRef{
				Value: &openapi.Schema{Type: "string"},
			}},
			func(t *testing.T, i interface{}, err error) {
				test.Ok(t, err)
				test.Equals(t, []string{}, i)
			},
		},
		{
			"string array",
			`["foo", "bar"]`,
			&openapi.Schema{Type: "array", Items: &openapi.SchemaRef{
				Value: &openapi.Schema{Type: "string"},
			}},
			func(t *testing.T, i interface{}, err error) {
				test.Ok(t, err)
				test.Equals(t, []string{"foo", "bar"}, i)
			},
		},
		{
			"object array",
			`[{"name": "foo", "age": 12}]`,
			openapitest.NewSchema("array", openapitest.WithItems(
				openapitest.NewSchema("object",
					openapitest.WithProperty("name", openapitest.NewSchema("string")),
					openapitest.WithProperty("age", openapitest.NewSchema("integer")),
				),
			),
			),
			func(t *testing.T, i interface{}, err error) {
				test.Ok(t, err)
				v := reflect.ValueOf(i)
				test.Equals(t, 1, v.Len())
				test.Equals(t, &struct {
					Name string
					Age  int64
				}{Name: "foo", Age: 12}, v.Index(0).Interface())
			},
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			i, err := Parse([]byte(d.s), media.ParseContentType("application/json"), &openapi.SchemaRef{Value: d.schema})
			d.fn(t, i, err)
		})
	}
}

func TestParse_Bool(t *testing.T) {
	cases := []struct {
		s      string
		schema *openapi.Schema
		exp    bool
		err    error
	}{
		{
			`true`,
			&openapi.Schema{Type: "boolean"},
			true,
			nil,
		},
		{
			`1`,
			&openapi.Schema{Type: "boolean"},
			false,
			fmt.Errorf("expected bool but got float64"),
		},
		{
			`false`,
			&openapi.Schema{Type: "boolean"},
			false,
			nil,
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.s, func(t *testing.T) {
			t.Parallel()
			i, err := Parse([]byte(d.s), media.ParseContentType("application/json"), &openapi.SchemaRef{Value: d.schema})
			test.Equals(t, d.err, err)
			test.Equals(t, d.exp, i)
		})
	}
}

func TestParse_Errors(t *testing.T) {
	cases := []struct {
		s      string
		schema *openapi.Schema
		err    error
	}{
		{
			``,
			&openapi.Schema{},
			fmt.Errorf("invalid json format: unexpected end of JSON input"),
		},
		{
			`bar`,
			&openapi.Schema{Type: "string"},
			fmt.Errorf("invalid json format: invalid character 'b' looking for beginning of value"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.s, func(t *testing.T) {
			t.Parallel()
			_, err := Parse([]byte(d.s), media.ParseContentType("application/json"), &openapi.SchemaRef{Value: d.schema})
			test.Equals(t, d.err, err)
		})
	}
}
