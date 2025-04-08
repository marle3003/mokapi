package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
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
				require.Equal(t, map[string]interface{}{"name": "Dash Cool Fitness Tracker"}, v)
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
				require.Equal(t, map[string]interface{}{"description": "Daily regularly you for recline her choir insufficient poised tribe. Mustering will might annually backwards abroad finger say."}, v)
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
				require.Equal(t, map[string]interface{}{"category": "musical instruments"}, v)
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
				require.Equal(t, map[string]interface{}{"material": "stainless"}, v)
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
