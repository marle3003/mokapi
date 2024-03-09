package generator

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/json/schema"
	"testing"
)

func TestPet(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "pet name",
			req:  &Request{Names: []string{"pet", "name"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Betty", v)
			},
		},
		{
			name: "pets name",
			req:  &Request{Names: []string{"pets", "name"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"Fyodor Dogstoevsky", "Woofgang Puck"}, v)
			},
		},
		{
			name: "pets name",
			req: &Request{
				Names:  []string{"pets", "name"},
				Schema: &schema.Schema{Type: []string{"array"}, MinItems: toIntP(4)},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"Fyodor Dogstoevsky", "Woofgang Puck", "Chompers", "Khaleesi"}, v)
			},
		},
		{
			name: "pet category",
			req:  &Request{Names: []string{"pet", "category"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "bat", v)
			},
		},
		{
			name: "pet category object",
			req:  &Request{Names: []string{"pet", "category", "name"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "bat", v)
			},
		},
		{
			name: "pet categories",
			req:  &Request{Names: []string{"pet", "categories"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"elk", "fish"}, v)
			},
		},
		{
			name: "pet category with schema",
			req: &Request{
				Names:  []string{"pet"},
				Schema: mustParse(`{"type": "object", "properties": {"category": {"type": "object", "properties": {"name": {"type": "string"},"id": {"type": "integer"}} }}}`),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"category": map[string]interface{}{"id": 5571200078501983580, "name": "bat"}}, v)
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
