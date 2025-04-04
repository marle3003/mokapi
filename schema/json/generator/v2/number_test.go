package v2

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestNumber(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "partyNumber",
			req: &Request{
				Path:   []string{"partyNumber"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "80291093648", v)
			},
		},
		{
			name: "partyNumbers",
			req: &Request{
				Path: []string{"partyNumbers"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"22910936489", "71180573501"}, v)
			},
		},
		{
			name: "partyNumbers as array",
			req: &Request{
				Path:   []string{"partyNumbers"},
				Schema: schematest.New("array", schematest.WithItems("string")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"22910936489", "71180573501"}, v)
			},
		},
		{
			name: "partyNumber as array",
			req: &Request{
				Path:   []string{"partyNumber"},
				Schema: schematest.New("array", schematest.WithItems("string")),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"22910936489", "71180573501"}, v)
			},
		},
		{
			name: "employeeNumber with min=max",
			req: &Request{
				Path: []string{"id"},
				Schema: schematest.New("string",
					schematest.WithMinLength(8),
					schematest.WithMaxLength(8),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Len(t, v, 8)
				require.Equal(t, "80291093", v)
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
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}
