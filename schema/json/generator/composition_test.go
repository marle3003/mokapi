package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
	"testing"
)

func TestComposition(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "anyOf empty",
			req: &Request{
				Path: Path{
					&PathElement{Schema: &schema.Ref{Value: schematest.NewAny()}},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(-3652171958352792229), v)
			},
		},
		{
			name: "anyOf string or number",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: &schema.Ref{
							Value: schematest.NewAny(
								schematest.New("string", schematest.WithMaxLength(5)),
								schematest.New("number", schematest.WithMinimum(0)),
							),
						},
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "l", v)
			},
		},
		{
			name: "allOf",
			req: &Request{
				Path: Path{
					&PathElement{
						Schema: &schema.Ref{
							Value: schematest.NewAllOf(
								schematest.New("object",
									schematest.WithProperty("foo", schematest.New("string"))),
								schematest.New("object",
									schematest.WithProperty("bar", schematest.New("number"))),
							),
						},
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"bar": 5.359460873460473e+307, "foo": "FqwCrwMfkOjojx"}, v)
			},
		},
		{
			name: "oneOf",
			req: &Request{
				Path: Path{
					&PathElement{
						Name: "price",
						Schema: &schema.Ref{
							Value: schematest.NewOneOf(
								schematest.New("number", schematest.WithMultipleOf(5)),
								schematest.New("number", schematest.WithMultipleOf(3)),
							),
						},
					},
				},
			},

			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(280), v)
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
