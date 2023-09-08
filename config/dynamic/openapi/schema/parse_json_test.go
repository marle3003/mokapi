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
				require.EqualError(t, err, `could not parse {name: bar, age: 12}, too many properties for object, expected any of schema type=object properties=[name]`)
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
				require.EqualError(t, err, `could not parse {foo: 12, bar: true}, expected one of schema type=object properties=[foo], schema type=object properties=[bar]`)
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
				require.EqualError(t, err, `could not parse {foo: 12}, it is not valid for only one schema, expected one of schema type=object properties=[foo], schema type=object properties=[foo]`)
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
			"not exclusive min",
			"12",
			&schema.Schema{Type: "integer", Minimum: toFloatP(12), ExclusiveMinimum: toBoolP(true)},
			0,
			fmt.Errorf("12 is lower or equal as the required minimum 12, expected schema type=integer minimum=12 exclusiveMinimum"),
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
			&schema.Schema{Type: "integer", Maximum: toFloatP(11)},
			0,
			fmt.Errorf("12 is greater as the required maximum 11, expected schema type=integer maximum=11"),
		},
		{
			"not exclusive max",
			"12",
			&schema.Schema{Type: "integer", Maximum: toFloatP(12), ExclusiveMaximum: toBoolP(true)},
			0,
			fmt.Errorf("12 is greater or equal as the required maximum 12, expected schema type=integer maximum=12 exclusiveMaximum"),
		},
		{
			"not in enum",
			"12",
			&schema.Schema{Type: "integer", Enum: []interface{}{1, 2, 3}},
			0,
			fmt.Errorf("value 12 does not match one in the enum [1, 2, 3]"),
		},
		{
			"in enum",
			"2",
			&schema.Schema{Type: "integer", Enum: []interface{}{1, 2, 3}},
			2,
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
			"type not defined",
			"3.612",
			&schema.Schema{},
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
			"not min float",
			"3.612",
			&schema.Schema{Type: "number", Format: "float", Minimum: toFloatP(3.7)},
			0,
			fmt.Errorf("3.612 is lower as the required minimum 3.7, expected schema type=number format=float minimum=3.7"),
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
		{
			"not max float",
			"3.612",
			&schema.Schema{Type: "number", Format: "float", Maximum: toFloatP(3.6)},
			0,
			fmt.Errorf("3.612 is greater as the required maximum 3.6, expected schema type=number format=float maximum=3.6"),
		},
		{
			"not max exclusive",
			"3.6",
			&schema.Schema{Type: "number", Maximum: toFloatP(3.6), ExclusiveMaximum: toBoolP(true)},
			0,
			fmt.Errorf("3.6 is greater or equal as the required maximum 3.6, expected schema type=number maximum=3.6 exclusiveMaximum"),
		},
		{
			"not in enum",
			"3.6",
			&schema.Schema{Type: "number", Enum: []interface{}{3, 4, 5.5}},
			0,
			fmt.Errorf("value 3.6 does not match one in the enum [3, 4, 5.5]"),
		},
		{
			"in enum",
			"3.6",
			&schema.Schema{Type: "number", Enum: []interface{}{3.6, 4, 5.5}},
			3.6,
			nil,
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
			"type not defined",
			`{"foo":"bar"}`,
			&schema.Schema{},
			func(t *testing.T, v interface{}, _ error) {
				require.Equal(t, &struct {
					Foo string `json:"foo"`
				}{Foo: "bar"}, v)
			},
		},
		{
			"example type not defined",
			`{"type":"object","properties":{"id":{"format":"int64","type":"integer"}}}`,
			&schema.Schema{},
			func(t *testing.T, v interface{}, _ error) {
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
				require.EqualError(t, err, `missing required field age on {name: foo}, expected schema type=object properties=[name, age] required=[name age]`)
			},
		},
		{
			"not in enum",
			`{"name": "foo"}`,
			schematest.New("object",
				schematest.WithRequired("name"),
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithEnum([]interface{}{map[string]interface{}{"name": "bar"}}),
			),
			func(t *testing.T, _ interface{}, err error) {
				require.EqualError(t, err, `value '{name: foo}' does not match one in the enum [{name: bar}]`)
			},
		},
		{
			"in enum",
			`{"name": "foo"}`,
			schematest.New("object",
				schematest.WithRequired("name"),
				schematest.WithProperty("name", schematest.New("string")),
				schematest.WithEnum([]interface{}{map[string]interface{}{"name": "foo"}}),
			),
			func(t *testing.T, _ interface{}, err error) {
				require.NoError(t, err)
			},
		},
		{
			"not minProperties",
			`{"name": "foo"}`,
			schematest.New("object",
				schematest.WithMinProperties(2),
			),
			func(t *testing.T, _ interface{}, err error) {
				require.EqualError(t, err, `validation error minProperties on {name: foo}, expected schema type=object minProperties=2 free-form=true`)
			},
		},
		{
			"not maxProperties",
			`{"name": "foo", "age": 12}`,
			schematest.New("object",
				schematest.WithMaxProperties(1),
			),
			func(t *testing.T, _ interface{}, err error) {
				require.Truef(t, strings.HasSuffix(err.Error(), "expected schema type=object maxProperties=1 free-form=true"), "but error is %v", err)
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
			"not min items",
			`["foo", "bar"]`,
			&schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				MinItems: toIntP(3),
			},
			func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "validation error minItems on [foo, bar], expected schema type=array minItems=3")
			},
		},
		{
			"min items",
			`["foo", "bar"]`,
			&schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				MinItems: toIntP(2),
			},
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, i)
			},
		},
		{
			"not max items",
			`["foo", "bar"]`,
			&schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				MaxItems: toIntP(1),
			},
			func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "validation error maxItems on [foo, bar], expected schema type=array maxItems=1")
			},
		},
		{
			"max items",
			`["foo", "bar"]`,
			&schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				MaxItems: toIntP(2),
			},
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, i)
			},
		},
		{
			"not in enum",
			`["foo", "bar"]`,
			&schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				Enum: []interface{}{[]string{"foo", "test"}},
			},
			func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "value [foo, bar] does not match one in the enum [[foo, test]]")
			},
		},
		{
			"not with correct order in enum",
			`["foo", "bar"]`,
			&schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				Enum: []interface{}{[]string{"bar", "foo"}},
			},
			func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "value [foo, bar] does not match one in the enum [[bar, foo]]")
			},
		},
		{
			"in enum",
			`["foo", "bar"]`,
			&schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				Enum: []interface{}{[]string{"foo", "test"}, []string{"foo", "bar"}},
			},
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"foo", "bar"}, i)
			},
		},
		{
			"not uniqueItems",
			`["foo", "foo"]`,
			&schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				UniqueItems: true,
			},
			func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "value [foo, foo] must contain unique items, expected schema type=array unique-items")
			},
		},
		{
			"uniqueItems",
			`["foo", "bar"]`,
			&schema.Schema{
				Type: "array",
				Items: &schema.Ref{
					Value: &schema.Schema{Type: "string"},
				},
				UniqueItems: true,
			},
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
