package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
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
				require.Equal(t, 7.509838941801037, v)
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
				require.Equal(t, float64(0), v)
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
				require.Equal(t, 18.900000000000002, v)
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
				require.Equal(t, int64(8), v)
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
				require.Equal(t, int64(0), v)
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
				require.Equal(t, int64(27), v)
				require.Equal(t, 3925, 11775/3)
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
				require.Equal(t, "13282419360", v)
			},
		},
		{
			name: "partyNumbers",
			req: &Request{
				Path:   []string{"partyNumbers"},
				Schema: schematest.NewTypes(nil, schematest.WithMinItems(1)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"32824193600"}, v)
			},
		},
		{
			name: "partyNumbers as array",
			req: &Request{
				Path:   []string{"partyNumbers"},
				Schema: schematest.New("array", schematest.WithItems("string"), schematest.WithMinItems(1)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"32824193600"}, v)
			},
		},
		{
			name: "partyNumber as array",
			req: &Request{
				Path:   []string{"partyNumber"},
				Schema: schematest.New("array", schematest.WithItems("string"), schematest.WithMinItems(1)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"32824193600"}, v)
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
				require.Equal(t, "53282419", v)
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
				require.Equal(t, int64(1929), v)
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
				require.Equal(t, int64(1929), v)
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
				require.Equal(t, int64(2010), v)
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
				require.Equal(t, int64(1995), v)
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
				require.Equal(t, int64(9), v)
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
				require.Equal(t, int64(5), v)
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
