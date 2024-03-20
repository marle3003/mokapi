package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/json/schematest"
	"testing"
)

func TestCurrency(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "currency",
			req: &Request{
				Path: Path{
					&PathElement{Name: "currency", Schema: schematest.NewRef("string")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "BYR", v)
			},
		},
		{
			name: "price",
			req: &Request{
				Path: Path{
					&PathElement{Name: "price", Schema: schematest.NewRef("number")},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 60959.16, v)
			},
		},
		{
			name: "price with max=99",
			req: &Request{
				Path: Path{
					&PathElement{Name: "price", Schema: schematest.NewRef("number", schematest.WithMaximum(99))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 60.34, v)
			},
		},
		{
			name: "price object",
			req: &Request{
				Path: Path{
					&PathElement{Name: "price", Schema: schematest.NewRef("object",
						schematest.WithProperty("value", schematest.New("integer")),
						schematest.WithProperty("currency", schematest.New("string")),
						schematest.WithProperty("name", schematest.New("string")),
					)},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"currency": "MAD",
					"name":     "Velvet",
					"value":    60959.16,
				}, v)
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
