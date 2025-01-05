package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schematest"
	"testing"
)

func TestConditional(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{

		{
			name: "if-then",
			req: &Request{
				Path: Path{
					&PathElement{
						Name: "if",
						Schema: schematest.NewRef("object",
							schematest.WithProperty("foo", schematest.New("string")),
							schematest.WithIf(schematest.NewRef("object",
								schematest.WithProperty("foo", schematest.New("string", schematest.WithConst("FqwCrwMfkOjojx"))),
							)),
							schematest.WithThen(schematest.NewRef("object",
								schematest.WithProperty("bar", schematest.New("string")),
							)),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "FqwCrwMfkOjojx", "bar": "Sza"}, v)
			},
		},
		{
			name: "if-else",
			req: &Request{
				Path: Path{
					&PathElement{
						Name: "if",
						Schema: schematest.NewRef("object",
							schematest.WithProperty("foo", schematest.New("string")),
							schematest.WithIf(schematest.NewRef("object",
								schematest.WithProperty("foo", schematest.New("string", schematest.WithConst(""))),
							)),
							schematest.WithElse(schematest.NewRef("object",
								schematest.WithProperty("bar", schematest.New("string")),
							)),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"foo": "FqwCrwMfkOjojx", "bar": "Sza"}, v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}
