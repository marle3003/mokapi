package openapi_test

import (
	"fmt"
	"math"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/test"
	"testing"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		name   string
		i      interface{}
		schema *openapi.Schema
		err    error
	}{
		{
			"enum",
			4,
			&openapi.Schema{Type: "integer", Enum: []interface{}{1, 2, 3, 4}},
			nil,
		},
		{
			"enum not matching",
			5,
			&openapi.Schema{Type: "integer", Enum: []interface{}{1, 2, 3, 4}},
			fmt.Errorf("value does not match any enum value"),
		},
		{
			"enum object",
			map[string]interface{}{"name": "foo", "age": 12},
			&openapi.Schema{Type: "integer", Enum: []interface{}{
				map[string]interface{}{"name": "bar", "age": 21},
				map[string]interface{}{"name": "foo", "age": 12},
			},
			},
			nil,
		},
		{
			"string",
			"foobar",
			&openapi.Schema{Type: "string"},
			nil,
		},
		{
			"not string",
			12,
			&openapi.Schema{Type: "string"},
			fmt.Errorf("expected string got int"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := openapi.Validate(d.i, d.schema)
			test.Equals(t, d.err, err)
		})
	}
}

func TestValidate_String(t *testing.T) {
	cases := []struct {
		name   string
		i      interface{}
		schema *openapi.Schema
		err    error
	}{
		{
			"nil",
			nil,
			&openapi.Schema{},
			nil,
		},
		{
			"string",
			"gbRMaRxHkiJBPta",
			&openapi.Schema{Type: "string"},
			nil,
		},
		{
			"by pattern",
			"013-64-5994",
			&openapi.Schema{Type: "string", Pattern: "^\\d{3}-\\d{2}-\\d{4}$"},
			nil,
		},
		{
			"not pattern",
			"013-64-59943",
			&openapi.Schema{Type: "string", Pattern: "^\\d{3}-\\d{2}-\\d{4}$"},
			fmt.Errorf("value does not match pattern"),
		},
		{
			"date",
			"1908-12-07",
			&openapi.Schema{Type: "string", Format: "date"},
			nil,
		},
		{
			"not date",
			"1908-12-7",
			&openapi.Schema{Type: "string", Format: "date"},
			fmt.Errorf("string is not a date RFC3339"),
		},
		{
			"date-time",
			"1908-12-07T04:14:25Z",
			&openapi.Schema{Type: "string", Format: "date-time"},
			nil,
		},
		{
			"not date-time",
			"1908-12-07 T04:14:25Z",
			&openapi.Schema{Type: "string", Format: "date-time"},
			fmt.Errorf("string is not a date-time RFC3339"),
		},
		{
			"password",
			"H|$9lb{J<+S;",
			&openapi.Schema{Type: "string", Format: "password"},
			nil,
		},
		{
			"email",
			"markusmoen@pagac.net",
			&openapi.Schema{Type: "string", Format: "email"},
			nil,
		},
		{
			"not email",
			"markusmoen@@pagac.net",
			&openapi.Schema{Type: "string", Format: "email"},
			fmt.Errorf("string is not an email address"),
		},
		{
			"uuid",
			"590c1440-9888-45b0-bd51-a817ee07c3f2",
			&openapi.Schema{Type: "string", Format: "uuid"},
			nil,
		},
		{
			"not uuid",
			"590c1440-9888-45b0-bd51-a817ee07c3f2a",
			&openapi.Schema{Type: "string", Format: "uuid"},
			fmt.Errorf("string is not an uuid"),
		},
		{
			"ipv4",
			"152.23.53.100",
			&openapi.Schema{Type: "string", Format: "ipv4"},
			nil,
		},
		{
			"not ipv4",
			"152.23.53.100.",
			&openapi.Schema{Type: "string", Format: "ipv4"},
			fmt.Errorf("string is not an ipv4"),
		},
		{
			"ipv6",
			"8898:ee17:bc35:9064:5866:d019:3b95:7857",
			&openapi.Schema{Type: "string", Format: "ipv6"},
			nil,
		},
		{
			"not ipv6",
			"-8898:ee17:bc35:9064:5866:d019:3b95:7857",
			&openapi.Schema{Type: "string", Format: "ipv6"},
			fmt.Errorf("string is not an ipv6"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := openapi.Validate(d.i, d.schema)
			test.Equals(t, d.err, err)
		})
	}
}

func TestValidate_Integer(t *testing.T) {
	cases := []struct {
		name   string
		i      interface{}
		schema *openapi.Schema
		err    error
	}{
		{
			"int32",
			int32(12),
			&openapi.Schema{Type: "integer", Format: "int32"},
			nil,
		},
		{
			"not int32",
			int64(math.MaxInt64),
			&openapi.Schema{Type: "integer", Format: "int32"},
			fmt.Errorf("integer is not int32"),
		},
		{
			"min",
			int64(12),
			&openapi.Schema{Type: "integer", Minimum: toFloatP(5)},
			nil,
		},
		{
			"not min",
			int64(12),
			&openapi.Schema{Type: "integer", Minimum: toFloatP(13)},
			fmt.Errorf("value is lower as defined minimum 13"),
		},
		{
			"max",
			int64(12),
			&openapi.Schema{Type: "integer", Maximum: toFloatP(13)},
			nil,
		},
		{
			"not max",
			int64(12),
			&openapi.Schema{Type: "integer", Maximum: toFloatP(5)},
			fmt.Errorf("value is greater as defined maximum 5"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := openapi.Validate(d.i, d.schema)
			test.Equals(t, d.err, err)
		})
	}
}

func TestValidate_Number(t *testing.T) {
	cases := []struct {
		name   string
		i      interface{}
		schema *openapi.Schema
		err    error
	}{
		{
			"float32",
			3.612,
			&openapi.Schema{Type: "number", Format: "float32"},
			nil,
		},
		{
			"not float32",
			float64(math.MaxFloat64),
			&openapi.Schema{Type: "number", Format: "float32"},
			fmt.Errorf("number is not float32"),
		},
		{
			"min",
			float64(3.612),
			&openapi.Schema{Type: "number", Minimum: toFloatP(3.6)},
			nil,
		},
		{
			"not min",
			3.612,
			&openapi.Schema{Type: "number", Minimum: toFloatP(3.7)},
			fmt.Errorf("value is lower as defined minimum 3.7"),
		},
		{
			"max",
			3.612,
			&openapi.Schema{Type: "number", Maximum: toFloatP(3.7)},
			nil,
		},
		{
			"not max",
			3.612,
			&openapi.Schema{Type: "number", Maximum: toFloatP(3.6)},
			fmt.Errorf("value is greater as defined maximum 3.6"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := openapi.Validate(d.i, d.schema)
			test.Equals(t, d.err, err)
		})
	}
}

func TestValidate_Object(t *testing.T) {
	cases := []struct {
		name   string
		i      interface{}
		schema *openapi.Schema
		err    error
	}{
		{
			"empty",
			struct{}{},
			&openapi.Schema{Type: "object"},
			nil,
		},
		{
			"simple",
			struct {
				Name string
				Age  int
			}{Name: "foo", Age: 12},
			openapitest.NewSchema("object",
				openapitest.WithProperty("name", openapitest.NewSchema("string")),
				openapitest.WithProperty("age", openapitest.NewSchema("integer")),
			),
			nil,
		},
		{
			"missing but not required",
			struct {
				Name string
			}{Name: "foo"},
			openapitest.NewSchema("object",
				openapitest.WithProperty("name", openapitest.NewSchema("string")),
				openapitest.WithProperty("age", openapitest.NewSchema("integer")),
			),
			nil,
		},
		{
			"missing but required",
			struct {
				Name string
			}{Name: "foo"},
			openapitest.NewSchema("object",
				openapitest.WithRequired("name", "age"),
				openapitest.WithProperty("name", openapitest.NewSchema("string")),
				openapitest.WithProperty("age", openapitest.NewSchema("integer")),
			),
			fmt.Errorf("expected required property age"),
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := openapi.Validate(d.i, d.schema)
			test.Equals(t, d.err, err)
		})
	}
}

func TestValidate_Array(t *testing.T) {
	cases := []struct {
		name   string
		i      interface{}
		schema *openapi.Schema
		err    error
	}{
		{
			"empty",
			[]interface{}{},
			&openapi.Schema{Type: "array"},
			nil,
		},
		{
			"simple",
			[]interface{}{1, 2, 3, 4},
			openapitest.NewSchema("array",
				openapitest.WithItems(openapitest.NewSchema("integer")),
			),
			nil,
		},
		{
			"invalid item",
			[]interface{}{1, 2, 3, 4.3},
			openapitest.NewSchema("array",
				openapitest.WithItems(openapitest.NewSchema("integer")),
			),
			fmt.Errorf("expected integer got float64"),
		},
		{
			"unique item",
			[]interface{}{1, 2, 3, 3},
			openapitest.NewSchema("array",
				openapitest.WithUniqueItems(),
				openapitest.WithItems(openapitest.NewSchema("integer")),
			),
			fmt.Errorf("array requires unique items"),
		},
		{
			"array of objects",
			[]interface{}{struct {
				name string
			}{name: "foo"}},
			openapitest.NewSchema("array",
				openapitest.WithItems(openapitest.NewSchema("object",
					openapitest.WithProperty("name", openapitest.NewSchema("string")),
					openapitest.WithProperty("age", openapitest.NewSchema("integer")),
				)),
			),
			nil,
		},
	}

	t.Parallel()
	for _, c := range cases {
		d := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := openapi.Validate(d.i, d.schema)
			test.Equals(t, d.err, err)
		})
	}
}
