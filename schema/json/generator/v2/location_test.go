package v2

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestLocation(t *testing.T) {
	testcases := []struct {
		name    string
		request *Request
		test    func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "country",
			request: &Request{
				Path:   []string{"country"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Slovenia", v)
			},
		},
		{
			name: "country with pattern [A-Z]{2}",
			request: &Request{
				Path: []string{"country"},
				Schema: schematest.New("string",
					schematest.WithPattern("[A-Z]{2}")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "SI", v)
			},
		},
		{
			name: "country with pattern [a-z]{2}",
			request: &Request{
				Path: []string{"country"},
				Schema: schematest.New("string",
					schematest.WithPattern("[a-z]{2}")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "si", v)
			},
		},
		{
			name: "country with pattern [a-zA-Z]{4}",
			request: &Request{
				Path: []string{"country"},
				Schema: schematest.New("string",
					schematest.WithPattern("^[a-zA-Z]{4}$")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "cQOp", v)
			},
		},
		{
			name: "country maxlength 15",
			request: &Request{
				Path:   []string{"country"},
				Schema: schematest.New("string", schematest.WithMaxLength(15)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Slovenia", v)
			},
		},
		{
			name: "countryName",
			request: &Request{
				Path:   []string{"countryName"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Slovenia", v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)
			Seed(1234567)

			v, err := New(tc.request)
			tc.test(t, v, err)
		})
	}
}
