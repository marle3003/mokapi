package generator

import (
	"mokapi/config/dynamic"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
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
				Path: []string{"pet"},
				Schema: schematest.New("object",
					schematest.WithProperty("name", nil),
					schematest.WithRequired("name"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Betty"}, v)
			},
		},
		{
			name: "pet-name as string",
			req: &Request{
				Path: []string{"pet"},
				Schema: schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithRequired("name"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Betty"}, v)
			},
		},
		{
			name: "pets-name",
			req: &Request{
				Path: []string{"pet"},
				Schema: schematest.New("array",
					schematest.WithMinItems(1),
					schematest.WithItemsNew(
						&schema.Schema{Reference: dynamic.Reference[*schema.Schema]{Ref: "#/components/schemas/Pet"}, Type: schema.Types{"string"}},
					)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"Fyodor Dogstoevsky"}, v)
			},
		},
		{
			name: "pets-name within object",
			req: &Request{
				Path: []string{"pets"},
				Schema: schematest.New("array",
					schematest.WithMinItems(1),
					schematest.WithItems("object",
						schematest.WithProperty("name", nil),
						schematest.WithRequired("name"),
					)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					map[string]interface{}{"name": "Fyodor Dogstoevsky"},
				}, v)
			},
		},
		{
			name: "pet-category",
			req: &Request{
				Path: []string{"pet"},
				Schema: schematest.New("object",
					schematest.WithProperty("category", nil),
					schematest.WithRequired("category"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"category": "cat"}, v)
			},
		},
		{
			name: "pet-category-name",
			req: &Request{
				Path: []string{"pet"},
				Schema: schematest.New("object",
					schematest.WithProperty("category", schematest.New("object",
						schematest.WithProperty("name", nil),
						schematest.WithRequired("name"),
					),
					),
					schematest.WithRequired("category"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"category": map[string]interface{}{"name": "cat"}}, v)
			},
		},
		{
			name: "pet-categories",
			req: &Request{
				Path: []string{"pet"},
				Schema: schematest.New("array",
					schematest.WithMinItems(1),
					schematest.WithItemsNew(
						&schema.Schema{Reference: dynamic.Reference[*schema.Schema]{Ref: "#/components/schemas/Category"}, Type: schema.Types{"string"}},
					)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"guinea pig"}, v)
			},
		},
		{
			name: "pet-category with schema",
			req: &Request{
				Path: []string{"pet"},
				Schema: schematest.New("object",
					schematest.WithProperty("category", schematest.New("object",
						schematest.WithProperty("name", schematest.New("string")),
						schematest.WithProperty("id", schematest.New("integer")),
						schematest.WithRequired("name", "id"),
					)),
					schematest.WithRequired("category"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"category": map[string]interface{}{"id": int64(36202), "name": "cat"}}, v)
			},
		},
		{
			name: "pet categories in two sub objects",
			req: &Request{
				Path: []string{"pet"},
				Schema: schematest.New("object",
					schematest.WithProperty("category", schematest.New("object",
						schematest.WithProperty("name", schematest.New("string")),
						schematest.WithProperty("id", schematest.New("integer")),
						schematest.WithRequired("name", "id"),
					)),
					schematest.WithProperty("petDetails", schematest.New("object",
						schematest.WithProperty("category", schematest.New("object",
							schematest.WithProperty("name", schematest.New("string")),
							schematest.WithProperty("id", schematest.New("integer")),
							schematest.WithRequired("name", "id"),
						)), schematest.WithRequired("category"),
					),
					), schematest.WithRequired("petDetails", "category"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"category": map[string]interface{}{
						"id":   int64(36202),
						"name": "cat",
					},
					"petDetails": map[string]interface{}{
						"category": map[string]interface{}{
							"id":   int64(36202),
							"name": "cat",
						},
					},
				}, v)
			},
		},
		{
			name: "pet and owner",
			req: &Request{
				Path: []string{"pet"},
				Schema: schematest.New("object",
					schematest.WithProperty("name", nil),
					schematest.WithProperty("owner", schematest.New("object",
						schematest.WithProperty("name", schematest.New("string")),
						schematest.WithRequired("name"),
					)),
					schematest.WithRequired("name", "owner"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Betty", "owner": map[string]interface{}{"name": "Emily Nelson"}}, v)
			},
		},
		{
			name: "pet with category",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithId("/pet"),
					schematest.WithProperty("name", nil),
					schematest.WithProperty("category", schematest.New("object",
						schematest.WithProperty("name", schematest.New("string")),
						schematest.WithRequired("name"),
					)),
					schematest.WithRequired("name", "category"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"category": map[string]interface{}{"name": "guinea pig"}, "name": "Betty"}, v)
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

func TestPetStore(t *testing.T) {
	var pet = schematest.New("object",
		schematest.WithProperty("id", schematest.New("integer")),
		schematest.WithProperty("category", schematest.New("object",
			schematest.WithProperty("id", schematest.New("integer")),
			schematest.WithProperty("name", schematest.New("string")),
			schematest.WithRequired("id", "name"),
		)),
		schematest.WithProperty("photoUrls", schematest.New("array", schematest.WithItems("string"))),
		schematest.WithProperty("tags", schematest.New("object",
			schematest.WithProperty("id", schematest.New("integer")),
			schematest.WithProperty("name", schematest.New("string")),
			schematest.WithRequired("id", "name"),
		)),
		schematest.WithProperty("status", schematest.New("string", schematest.WithEnum([]interface{}{"available", "pending", "sold"}))),
		schematest.WithRequired("id", "category", "photoUrls", "tags", "status"),
	)

	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "pet",
			req: &Request{
				Path:   []string{"pet"},
				Schema: pet,
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"category": map[string]interface{}{"id": int64(36202), "name": "rabbit"},
					"id":       int64(9900),
					"photoUrls": []interface{}{
						"http://www.financialvalue-added.com/end-to-end/envisioneer",
						"http://www.internationalnext-generation.name/scale",
						"https://www.brandrich.name/extend/implement/innovative/enterprise",
						"https://www.operationse-services.org/matrix/portals/vortals/e-markets",
						"https://www.executivefront-end.info/innovative"},
					"status": "available",
					"tags":   map[string]interface{}{"id": int64(18481), "name": "Echo"}},
					v)
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
