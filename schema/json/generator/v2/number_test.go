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
			name: "number with min and max",
			req: &Request{
				Schema: schematest.New("number",
					schematest.WithMinimum(0),
					schematest.WithMaximum(10),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 6.095916352063622, v)
			},
		},
		{
			name: "number with min, max and multiplyOf",
			req: &Request{
				Schema: schematest.New("number",
					schematest.WithMinimum(0),
					schematest.WithMaximum(10),
					schematest.WithMultipleOf(2.1),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 4.2, v)
			},
		},
		{
			name: "number with multiplyOf",
			req: &Request{
				Schema: schematest.New("number",
					schematest.WithMultipleOf(2.1),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 8242.5, v)
			},
		},
		{
			name: "integer with min and max",
			req: &Request{
				Schema: schematest.New("integer",
					schematest.WithMinimum(0),
					schematest.WithMaximum(10),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 8, v)
			},
		},
		{
			name: "integer with min, max and multiplyOf",
			req: &Request{
				Schema: schematest.New("integer",
					schematest.WithMinimum(0),
					schematest.WithMaximum(10),
					schematest.WithMultipleOf(3),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 3, v)
			},
		},
		{
			name: "integer with multiplyOf",
			req: &Request{
				Schema: schematest.New("integer",
					schematest.WithMultipleOf(3),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, -9223372036854775806, v)
				require.Equal(t, -3074457345618258602, -9223372036854775806/3)
			},
		},
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

func TestYear(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "year no schema",
			req: &Request{
				Path: []string{"year"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 1926, v)
			},
		},
		{
			name: "year",
			req: &Request{
				Path:   []string{"year"},
				Schema: schematest.New("integer"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(1926), v)
			},
		},
		{
			name: "year min",
			req: &Request{
				Path:   []string{"year"},
				Schema: schematest.New("integer", schematest.WithMinimum(1990)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(2196), v)
			},
		},
		{
			name: "year min max",
			req: &Request{
				Path: []string{"year"},
				Schema: schematest.New("integer",
					schematest.WithMinimum(1990),
					schematest.WithMaximum(2049),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(2016), v)
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

func TestQuantity(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "quantity",
			req: &Request{
				Path:   []string{"quantity"},
				Schema: schematest.New("integer"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(79), v)
			},
		},
		{
			name: "quantity min max",
			req: &Request{
				Path: []string{"quantity"},
				Schema: schematest.New("integer",
					schematest.WithMinimum(0),
					schematest.WithMaximum(50),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, int64(23), v)
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
