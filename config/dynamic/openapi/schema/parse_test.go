package schema_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/media"
	"testing"
)

func TestParseString(t *testing.T) {
	testcases := []struct {
		name string
		f    func(t *testing.T)
	}{
		{
			"int",
			func(t *testing.T) {
				i, err := schema.ParseString("42", &schema.Ref{Value: &schema.Schema{Type: "integer"}})
				require.NoError(t, err)
				require.Equal(t, int64(42), i)
			},
		},
		{
			"int64",
			func(t *testing.T) {
				i, err := schema.ParseString("42", &schema.Ref{Value: &schema.Schema{Type: "integer", Format: "int64"}})
				require.NoError(t, err)
				require.Equal(t, int64(42), i)
			},
		},
		{
			"int32",
			func(t *testing.T) {
				i, err := schema.ParseString("42", &schema.Ref{Value: &schema.Schema{Type: "integer", Format: "int32"}})
				require.NoError(t, err)
				require.Equal(t, int64(42), i)
			},
		},
		{
			"int32 max overflow",
			func(t *testing.T) {
				n := int64(math.MaxInt32) + 1
				_, err := schema.ParseString(fmt.Sprintf("%v", n), &schema.Ref{Value: &schema.Schema{Type: "integer", Format: "int32"}})
				require.EqualError(t, err, "could not parse '2147483648', represents a number either less than int32 min value or greater max value, expected schema type=integer format=int32")
			},
		},
		{
			"int32 min overflow",
			func(t *testing.T) {
				n := int64(math.MinInt32) - 1
				_, err := schema.ParseString(fmt.Sprintf("%v", n), &schema.Ref{Value: &schema.Schema{Type: "integer", Format: "int32"}})
				require.EqualError(t, err, "could not parse '-2147483649', represents a number either less than int32 min value or greater max value, expected schema type=integer format=int32")
			},
		},
		{
			"int but float",
			func(t *testing.T) {
				_, err := schema.ParseString("3.141", &schema.Ref{Value: &schema.Schema{Type: "integer"}})
				require.EqualError(t, err, "could not parse '3.141' as int, expected schema type=integer")
			},
		},
		{
			"int32 but float",
			func(t *testing.T) {
				_, err := schema.ParseString("3.141", &schema.Ref{Value: &schema.Schema{Type: "integer", Format: "int32"}})
				require.EqualError(t, err, "could not parse '3.141' as int, expected schema type=integer format=int32")
			},
		},
		{
			"float default",
			func(t *testing.T) {
				i, err := schema.ParseString("3.141", &schema.Ref{Value: &schema.Schema{Type: "number"}})
				require.NoError(t, err)
				require.Equal(t, 3.141, i)
			},
		},
		{
			"double",
			func(t *testing.T) {
				i, err := schema.ParseString("3.141", &schema.Ref{Value: &schema.Schema{Type: "number", Format: "double"}})
				require.NoError(t, err)
				require.Equal(t, 3.141, i)
			},
		},
		{
			"float",
			func(t *testing.T) {
				i, err := schema.ParseString("3.141", &schema.Ref{Value: &schema.Schema{Type: "number", Format: "float"}})
				require.NoError(t, err)
				require.Equal(t, 3.141, i)
			},
		},
	}
	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.f(t)
		})
	}
}

func TestObject_SpecialNames(t *testing.T) {
	testcases := []struct {
		name   string
		s      string
		schema *schema.Schema
		f      func(t *testing.T, i interface{}, err error)
	}{
		{
			"-",
			`{"ship-date":"2022-01-01"}`,
			schematest.New("object", schematest.WithProperty("ship-date", schematest.New("string"))),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Shipdate string `json:"ship-date"`
				}{Shipdate: "2022-01-01"}, i)
			},
		},
		{
			"- conflict",
			`{"ship-date":"2022-01-01","shipdate":"2023-01-01"}`,
			schematest.New("object",
				schematest.WithProperty("ship-date", schematest.New("string")),
				schematest.WithProperty("shipdate", schematest.New("string"))),
			func(t *testing.T, i interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, &struct {
					Ship_date string `json:"ship-date"`
					Shipdate  string `json:"shipdate"`
				}{Ship_date: "2022-01-01", Shipdate: "2023-01-01"}, i)
			},
		},
		{
			"- conflict",
			`{"ship-date":"2022-01-01","ship_date":"2023-01-01"}`,
			schematest.New("object",
				schematest.WithProperty("ship-date", schematest.New("string")),
				schematest.WithProperty("shipdate", schematest.New("string"))),
			func(t *testing.T, i interface{}, err error) {
				require.EqualError(t, err, "duplicate field name Ship_date")
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
