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
				require.Equal(t, int64(64863), v)
			},
		},
		{
			name:    "postcode",
			request: &Request{Names: []string{"postcode"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(64863), v)
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
				require.Equal(t, int64(6486), v)
			},
		},
		{
			name:    "postcodes",
			request: &Request{Names: []string{"postcodes"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(64362), int64(60304)}, v)
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
				require.Equal(t, []interface{}{int64(6436), int64(6030)}, v)
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
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.request)
			tc.test(t, v, err)
		})
	}
}
