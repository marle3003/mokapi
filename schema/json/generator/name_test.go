package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestName(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "name",
			req:  &Request{Path: []string{"name"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Ink", v)
			},
		},
		{
			name: "name as string",
			req: &Request{
				Path:   []string{"name"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Ink", v)
			},
		},
		{
			name: "customerName",
			req: &Request{
				Path: []string{"customerName"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Ink", v)
			},
		},
		{
			name: "campaignName",
			req: &Request{
				Path:   []string{"campaignName"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Ink", v)
			},
		},
		{
			name: "name with max length 3",
			req: &Request{
				Path: []string{"name"},
				Schema: schematest.New("string",
					schematest.WithMaxLength(3),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Neo", v)
			},
		},
		{
			name: "name with min and max length 4",
			req: &Request{
				Path: []string{"name"},
				Schema: schematest.New("string",
					schematest.WithMinLength(4),
					schematest.WithMaxLength(4),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Apex", v)
			},
		},
		{
			name: "name with min length 4",
			req: &Request{
				Path: []string{"name"},
				Schema: schematest.New("string",
					schematest.WithMinLength(4),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "TerraCove", v)
			},
		},
		{
			name: "name with min and max length 5",
			req: &Request{
				Path: []string{"name"},
				Schema: schematest.New("string",
					schematest.WithMinLength(5),
					schematest.WithMaxLength(5),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Exalt", v)
			},
		},
		{
			name: "name with min and max length 6",
			req: &Request{
				Path: []string{"name"},
				Schema: schematest.New("string",
					schematest.WithMinLength(6),
					schematest.WithMaxLength(6),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Spirit", v)
			},
		},
		{
			name: "name with min length 7",
			req: &Request{
				Path: []string{"name"},
				Schema: schematest.New("string",
					schematest.WithMinLength(7),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "EchoValley", v)
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
