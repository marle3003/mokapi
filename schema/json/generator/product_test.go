package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestProduct(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "name",
			req: &Request{
				Path: []string{"product"},
				Schema: schematest.New("object",
					schematest.WithProperty("name", nil),
					schematest.WithRequired("name"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Designed Can Opener"}, v)
			},
		},
		{
			name: "use path /products/name",
			req: &Request{
				Path:   []string{"products", "name"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Designed Can Opener", v)
			},
		},
		{
			name: "description",
			req: &Request{
				Path: []string{"product"},
				Schema: schematest.New("object",
					schematest.WithProperty("description", nil),
					schematest.WithRequired("description"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"description": "Experience the comfort of the Burlap Window beast, made from enamel with a enthusiastic design. It's ideal for indoor activities and loved by collectors and cooks."}, v)
			},
		},
		{
			name: "category",
			req: &Request{
				Path: []string{"product"},
				Schema: schematest.New("object",
					schematest.WithProperty("category", nil),
					schematest.WithRequired("category"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"category": "bathroom accessories"}, v)
			},
		},
		{
			name: "material",
			req: &Request{
				Path: []string{"product"},
				Schema: schematest.New("object",
					schematest.WithProperty("material", nil),
					schematest.WithRequired("material"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"material": "burlap"}, v)
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
