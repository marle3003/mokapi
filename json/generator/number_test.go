package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/json/schema"
	"mokapi/json/schematest"
	"testing"
)

func toFloat64P(f float64) *float64 { return &f }
func toIntP(i int) *int             { return &i }

func TestNumber(t *testing.T) {
	testcases := []struct {
		name    string
		request *Request
		test    func(t *testing.T, v interface{}, err error)
	}{
		{
			name:    "id",
			request: &Request{Names: []string{"id"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 5622490442062937727, v)
			},
		},
		{
			name: "id with max",
			request: &Request{Names: []string{"id"}, Schema: &schema.Schema{
				Type:    []string{"integer"},
				Maximum: toFloat64P(10000),
			}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 7727, v)
			},
		},
		{
			name: "id with min & max",
			request: &Request{Names: []string{"id"}, Schema: &schema.Schema{
				Type:    []string{"integer"},
				Minimum: toFloat64P(10),
				Maximum: toFloat64P(20),
			}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 18, v)
			},
		},
		{
			name:    "price",
			request: &Request{Names: []string{"price"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 60959.16, v)
			},
		},
		{
			name: "price with max",
			request: &Request{
				Names:  []string{"price"},
				Schema: schematest.New("number", schematest.WithMaximum(100)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 60.95, v)
			},
		},
		{
			name: "year",
			request: &Request{
				Names:  []string{"year"},
				Schema: schematest.New("integer"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 1926, v)
			},
		},
		{
			name: "year min",
			request: &Request{
				Names:  []string{"year"},
				Schema: schematest.New("integer", schematest.WithMinimum(1990)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 2196, v)
			},
		},
		{
			name: "year min max",
			request: &Request{
				Names:  []string{"year"},
				Schema: schematest.New("integer", schematest.WithMinimum(1990), schematest.WithMaximum(2049)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 2016, v)
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
