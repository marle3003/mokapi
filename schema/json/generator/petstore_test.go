package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schematest"
	"testing"
)

var pet = schematest.NewRef("object",
	schematest.WithProperty("id", schematest.New("integer")),
	schematest.WithProperty("category", schematest.New("object",
		schematest.WithProperty("id", schematest.New("integer")),
		schematest.WithProperty("name", schematest.New("string")),
	)),
	schematest.WithProperty("photoUrls", schematest.New("array", schematest.WithItems("string"))),
	schematest.WithProperty("tags", schematest.New("object",
		schematest.WithProperty("id", schematest.New("integer")),
		schematest.WithProperty("name", schematest.New("string")),
	)),
	schematest.WithProperty("status", schematest.New("string", schematest.WithEnum([]interface{}{"available", "pending", "sold"}))),
)

func TestPetStore(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "pet",
			req: &Request{
				Path: Path{
					&PathElement{
						Name:   "pet",
						Schema: pet,
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"category":  map[string]interface{}{"id": 83580, "name": "fish"},
					"id":        37727,
					"photoUrls": []interface{}{"https://www.principalapplications.biz/cultivate/e-enable/integrated"},
					"status":    "sold",
					"tags":      map[string]interface{}{"id": 69949, "name": "Sol"}},
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
