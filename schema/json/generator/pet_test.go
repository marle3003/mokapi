package generator

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestPet(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "pet-name",
			req: &Request{
				Path: Path{
					&PathElement{Name: "pet", Schema: schematest.New("object",
						schematest.WithProperty("name", nil),
					)},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Betty"}, v)
			},
		},
		{
			name: "pet-name as string",
			req: &Request{
				Path: Path{
					&PathElement{Name: "pet", Schema: schematest.New("object",
						schematest.WithProperty("name", schematest.New("string")),
					)},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Betty"}, v)
			},
		},
		{
			name: "pets-name",
			req: &Request{
				Path: Path{
					&PathElement{Name: "pets", Schema: schematest.New("array", schematest.WithItemsNew(
						&schema.Schema{Ref: "#/components/schemas/Pet", Type: schema.Types{"string"}},
					))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"Fyodor Dogstoevsky", "Woofgang Puck"}, v)
			},
		},
		{
			name: "pets-name within object",
			req: &Request{
				Path: Path{
					&PathElement{Name: "pets", Schema: schematest.New("array", schematest.WithItems(
						"object", schematest.WithProperty("name", nil),
					))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					map[string]interface{}{"name": "Fyodor Dogstoevsky"},
					map[string]interface{}{"name": "Woofgang Puck"},
				}, v)
			},
		},
		{
			name: "pet-category",
			req: &Request{
				Path: Path{
					&PathElement{Name: "pet", Schema: schematest.New("object", schematest.WithProperty("category", nil))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"category": "canary"}, v)
			},
		},
		{
			name: "pet-category-name",
			req: &Request{
				Path: Path{
					&PathElement{
						Name: "pet", Schema: schematest.New("object",
							schematest.WithProperty("category", schematest.New("object",
								schematest.WithProperty("name", nil)),
							),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"category": map[string]interface{}{"name": "canary"}}, v)
			},
		},
		{
			name: "pet-categories",
			req: &Request{
				Path: Path{
					&PathElement{Name: "pet", Schema: schematest.New("array", schematest.WithItemsNew(
						&schema.Schema{Ref: "#/components/schemas/Category", Type: schema.Types{"string"}},
					))},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"ferret", "rabbit"}, v)
			},
		},
		{
			name: "pet-category with schema",
			req: &Request{
				Path: Path{
					&PathElement{
						Name: "pet",
						Schema: schematest.New("object",
							schematest.WithProperty("category", schematest.New("object",
								schematest.WithProperty("name", schematest.New("string")),
								schematest.WithProperty("id", schematest.New("integer")),
							)),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"category": map[string]interface{}{"id": 83580, "name": "canary"}}, v)
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

func mustParse(b string) *schema.Schema {
	var s *schema.Schema
	err := json.Unmarshal([]byte(b), &s)
	if err != nil {
		panic(err)
	}
	return s
}
