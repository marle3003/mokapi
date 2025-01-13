package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestComposition_AllOf(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "with one integer schema",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: schematest.NewAllOf(
							schematest.New("integer"),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-3600881594791838082), v)
			},
		},
		{
			name: "with two integer schema, second is more precise",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: schematest.NewAllOf(
							schematest.New("integer"),
							schematest.New("integer", schematest.WithMinimum(0), schematest.WithMaximum(10)),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(6), v)
			},
		},
		{
			name: "with two integer schema, first is more precise",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: schematest.NewAllOf(
							schematest.New("integer", schematest.WithMinimum(0), schematest.WithMaximum(10)),
							schematest.New("integer"),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(6), v)
			},
		},
		{
			name: "first is any, second is integer",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: schematest.NewAllOf(
							schematest.NewAny(),
							schematest.New("integer"),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-3600881594791838082), v)
			},
		},
		{
			name: "one object",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: schematest.NewAllOf(
							schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "FqwCrwMfkOjojx"}, v)
			},
		},
		{
			name: "two object",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: schematest.NewAllOf(
							schematest.New("object", schematest.WithProperty("foo", schematest.New("string"))),
							schematest.New("object", schematest.WithProperty("bar", schematest.New("string"))),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"bar": "Sza", "foo": "FqwCrwMfkOjojx"}, v)
			},
		},
		{
			name: "two object with required properties",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: schematest.NewAllOf(
							schematest.New("object",
								schematest.WithProperty("foo", schematest.New("string")),
								schematest.WithRequired("foo"),
							),
							schematest.New("object",
								schematest.WithProperty("bar", schematest.New("string")),
								schematest.WithRequired("bar"),
							),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"bar": "Sza", "foo": "FqwCrwMfkOjojx"}, v)
			},
		},
		{
			name: "two object, first is integer and object",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: schematest.NewAllOf(
							schematest.NewTypes([]string{"integer", "object"},
								schematest.WithProperty("foo", schematest.New("string")),
								schematest.WithRequired("foo"),
							),
							schematest.New("object",
								schematest.WithProperty("bar", schematest.New("string")),
								schematest.WithRequired("bar"),
							),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"bar": "Sza", "foo": "FqwCrwMfkOjojx"}, v)
			},
		},
		{
			name: "no shared type",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: schematest.NewAllOf(
							schematest.NewTypes([]string{"integer", "string"}),
							schematest.NewTypes([]string{"number", "boolean"}),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "generate random data for schema failed: all of schema type=[integer, string], schema type=[number, boolean]: no shared types found")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)
			Seed(1234567)

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}
