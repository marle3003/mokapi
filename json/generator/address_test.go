package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/json/schema"
	"mokapi/json/schematest"
	"testing"
)

func TestAddress(t *testing.T) {
	testcases := []struct {
		name    string
		request *Request
		test    func(t *testing.T, v interface{}, err error)
	}{
		{
			name:    "city",
			request: &Request{Names: []string{"city"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "New Orleans", v)
			},
		},
		{
			name:    "city array",
			request: &Request{Names: []string{"cities"}, Schema: &schema.Schema{Type: []string{"array"}}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"Plano", "New York City"}, v)
			},
		},
		{
			name:    "zip",
			request: &Request{Names: []string{"zip"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "7291", v)
			},
		},
		{
			name:    "postcode",
			request: &Request{Names: []string{"postcode"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "7291", v)
			},
		},
		{
			name: "postcode",
			request: &Request{
				Names:  []string{"postcode"},
				Schema: schematest.New("string", schematest.WithMinLength(5), schematest.WithMaxLength(5)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "80291", v)
			},
		},
		{
			name: "zip with min & max",
			request: &Request{
				Names: []string{"postcode"},
				Schema: schematest.New("integer",
					schematest.WithMinimum(1000),
					schematest.WithMaximum(9999),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(8029), v)
			},
		},
		{
			name:    "postcodes",
			request: &Request{Names: []string{"postcodes"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"29109", "6489"}, v)
			},
		},
		{
			name: "zips with min & max",
			request: &Request{
				Names: []string{"postcodes"},
				Schema: schematest.New("array", schematest.WithItems(
					"integer",
					schematest.WithMinimum(1000),
					schematest.WithMaximum(9999),
				)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(7291), int64(9364)}, v)
			},
		},
		{
			name:    "longitude",
			request: &Request{Names: []string{"longitude"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 39.452988, v)
			},
		},
		{
			name:    "latitude",
			request: &Request{Names: []string{"latitude"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 19.726494, v)
			},
		},
		{
			name: "coAddress",
			request: &Request{
				Names:  []string{"coAddress"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Shanelle Wehner", v)
			},
		},
		{
			name: "country",
			request: &Request{
				Names:  []string{"country"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Slovenia", v)
			},
		},
		{
			name: "address - country",
			request: &Request{
				Names:  []string{"address", "country"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "SI", v)
			},
		},
		{
			name: "open address",
			request: &Request{
				Names: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("line1", schematest.New("string")),
					schematest.WithProperty("line2", schematest.New("string")),
					schematest.WithProperty("line3", schematest.New("string")),
					schematest.WithProperty("country", schematest.New("string")),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"line1":   "Shanelle Wehner",
					"line2":   "1093 Lockstown",
					"line3":   "Newark VT 41180",
					"country": "FJ",
				}, v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.request)
			tc.test(t, v, err)
		})
	}
}
