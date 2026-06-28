package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddress(t *testing.T) {
	testcases := []struct {
		name    string
		request *Request
		test    func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "city",
			request: &Request{
				Path: []string{"city"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "San Jose", v)
			},
		},
		{
			name: "city",
			request: &Request{
				Path:   []string{"city"},
				Schema: schematest.New("integer"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(3282), v)
			},
		},
		{
			name: "city array",
			request: &Request{
				Path: []string{"cities"}, Schema: schematest.New("array", schematest.WithMinItems(1))},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"Long Beach"}, v)
			},
		},
		{
			name: "zip",
			request: &Request{
				Path: []string{"zip"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "3282", v)
			},
		},
		{
			name: "zip with pattern should use pattern",
			request: &Request{
				Path:   []string{"zip"},
				Schema: schematest.New("string", schematest.WithPattern("[0-9]{4}")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "8029", v)
			},
		},
		{
			name: "zipCode",
			request: &Request{
				Path: []string{"zipCode"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "3282", v)
			},
		},
		{
			name: "postcode any type",
			request: &Request{
				Path: []string{"postcode"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "3282", v)
			},
		},
		{
			name: "postcode integer",
			request: &Request{
				Path:   []string{"postcode"},
				Schema: schematest.New("integer"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(3282), v)
			},
		},
		{
			name: "postcode number",
			request: &Request{
				Path:   []string{"postcode"},
				Schema: schematest.New("number"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, float64(3282), v)
			},
		},
		{
			name: "postcode string min and max",
			request: &Request{
				Path:   []string{"postcode"},
				Schema: schematest.New("string", schematest.WithMinLength(5), schematest.WithMaxLength(5)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "73282", v)
			},
		},
		{
			name: "zip with min & max",
			request: &Request{
				Path: []string{"zip"},
				Schema: schematest.New("integer",
					schematest.WithMinimum(1000),
					schematest.WithMaximum(9999),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(6328), v)
			},
		},
		{
			name:    "postcodes",
			request: &Request{Path: []string{"postcodes"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{}, v)
			},
		},
		{
			name: "postcodes with min & max",
			request: &Request{
				Path: []string{"postcodes"},
				Schema: schematest.New("array",
					schematest.WithMinItems(1),
					schematest.WithItems(
						"integer",
						schematest.WithMinimum(1000),
						schematest.WithMaximum(9999),
					)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{int64(3282)}, v)
			},
		},
		{
			name: "postal_code",
			request: &Request{
				Path: []string{"postal_code"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "3282", v)
			},
		},
		{
			name: "longitude",
			request: &Request{
				Path: []string{"longitude"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 90.354201, v)
			},
		},
		{
			name: "latitude",
			request: &Request{
				Path: []string{"latitude"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 45.1771, v)
			},
		},
		{
			name: "coAddress",
			request: &Request{
				Path:   []string{"coAddress"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Aria Hernandez", v)
			},
		},
		{
			name: "street",
			request: &Request{
				Path:   []string{"street"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "2824 North Walkborough", v)
			},
		},
		{
			name: "address - country",
			request: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("country", schematest.New("string", schematest.WithMaxLength(2))),
					schematest.WithRequired("country"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"country": "BJ"}, v)
			},
		},
		{
			name: "open address",
			request: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("line1", schematest.New("string")),
					schematest.WithProperty("line2", schematest.New("string")),
					schematest.WithProperty("line3", schematest.New("string")),
					schematest.WithProperty("country", schematest.New("string")),
					schematest.WithRequired("line1", "line2", "line3", "country"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"line1":   "Aria Hernandez",
					"line2":   "41936 Cornerton",
					"line3":   "Louisville/Jefferson KS 54911",
					"country": "Lesotho",
				}, v)
			},
		},
		{
			name: "houseNumber",
			request: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("houseNumber", schematest.New("string")),
					schematest.WithRequired("houseNumber"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"houseNumber": "10"}, v)
			},
		},
		{
			name: "houseNumber",
			request: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("houseNumber", schematest.New("integer", schematest.WithMinimum(110))),
					schematest.WithRequired("houseNumber"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"houseNumber": int64(32824)}, v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			Seed(1234567)

			v, err := New(tc.request)
			tc.test(t, v, err)
		})
	}
}
