package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
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
				require.Equal(t, "44f4ae5d-233e-4f89-ae02-126591065f49", v)
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
				require.Equal(t, "802910936489301180573501861460", v)
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
				require.Equal(t, "44f4ae5d-233e-4f89-ae02-126591065f49", v)
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
				require.Equal(t, "7291093", v)
			},
		},
		{
			name: "id",
			req: &Request{
				Path: []string{"id"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 37727, v)
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
				require.Equal(t, int64(7727), v)
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
				require.Equal(t, int64(18), v)
			},
		},
		{
			name: "ids with schema array",
			req: &Request{
				Path:   []string{"ids"},
				Schema: schematest.New("array"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{83580, 80588}, v)
			},
		},
		{
			name: "ids",
			req: &Request{
				Path: []string{"ids"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{83580, 80588}, v)
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
