package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"testing"
)

func TestNew(t *testing.T) {
	testcases := []struct {
		name    string
		request *Request
		test    func(t *testing.T, v interface{}, err error)
	}{
		//{
		//	name:    "simple string",
		//	request: &Request{Schema: &schema.Schema{Type: []string{"string"}}},
		//	test: func(t *testing.T, v interface{}, err error) {
		//		require.NoError(t, err)
		//		require.Equal(t, "FQwCRWMFkOjoJX", v)
		//	},
		//},
		//{
		//	name:    "city",
		//	request: &Request{Names: []string{"city"}},
		//	test: func(t *testing.T, v interface{}, err error) {
		//		require.NoError(t, err)
		//		require.Equal(t, "New Orleans", v)
		//	},
		//},
		//{
		//	name:    "city array",
		//	request: &Request{Names: []string{"cities"}, Schema: &schema.Schema{Type: []string{"array"}}},
		//	test: func(t *testing.T, v interface{}, err error) {
		//		require.NoError(t, err)
		//		require.Equal(t, []interface{}{"Plano", "New York City"}, v)
		//	},
		//},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.request)
			tc.test(t, v, err)
		})
	}
}
