package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestId(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "id",
			req: &Request{
				Path:   []string{"id"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "ce702a60-0f08-4819-bcd4-0907c044ad5c", v)
			},
		},
		{
			name: "id string with max",
			req: &Request{
				Path:   []string{"id"},
				Schema: schematest.New("string", schematest.WithMaxLength(30)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "132824193600225549115440653199", v)
			},
		},
		{
			name: "id string with min",
			req: &Request{
				Path:   []string{"id"},
				Schema: schematest.New("string", schematest.WithMinLength(4)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "ce702a60-0f08-4819-bcd4-0907c044ad5c", v)
			},
		},
		{
			name: "id string with min & max",
			req: &Request{
				Path: []string{"id"},
				Schema: schematest.New("string",
					schematest.WithMinLength(4),
					schematest.WithMaxLength(10),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "3282", v)
			},
		},
		{
			name: "id",
			req: &Request{
				Path: []string{"id"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 9900, v)
			},
		},
		{
			name: "id with max",
			req: &Request{
				Path:   []string{"id"},
				Schema: schematest.New("integer", schematest.WithMaximum(10000)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(990), v)
			},
		},
		{
			name: "id with min & max",
			req: &Request{
				Path: []string{"id"},
				Schema: schematest.New("integer",
					schematest.WithMinimum(10),
					schematest.WithMaximum(20),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(11), v)
			},
		},
		{
			name: "ids with schema array",
			req: &Request{
				Path:   []string{"ids"},
				Schema: schematest.New("array", schematest.WithMinItems(1)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{36202}, v)
			},
		},
		{
			name: "ids",
			req: &Request{
				Path:   []string{"ids"},
				Schema: schematest.NewTypes(nil, schematest.WithMinItems(1)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{36202}, v)
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
