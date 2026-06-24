package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestCurrency(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "currency",
			req: &Request{
				Path:   []string{"currency"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "BIF", v)
			},
		},
		{
			name: "currency object no properties",
			req: &Request{
				Path:   []string{"currency"},
				Schema: schematest.New("object"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"code": "BIF", "name": "Burundi Franc"}, v)
			},
		},
		{
			name: "currency object no properties",
			req: &Request{
				Path: []string{"currency"},
				Schema: schematest.New("object",
					schematest.WithProperty("code", nil),
					schematest.WithProperty("name", nil),
					schematest.WithRequired("code", "name"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"code": "BIF", "name": "Burundi Franc"}, v)
			},
		},
		{
			name: "price",
			req: &Request{
				Path:   []string{"price"},
				Schema: schematest.New("number"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 750983.89, v)
			},
		},
		{
			name: "price with max=99",
			req: &Request{
				Path:   []string{"price"},
				Schema: schematest.New("number", schematest.WithMaximum(99)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, 74.34, v)
			},
		},
		{
			name: "price object",
			req: &Request{
				Path: []string{"price"},
				Schema: schematest.New("object",
					schematest.WithProperty("value", schematest.New("integer")),
					schematest.WithProperty("currency", schematest.New("string")),
					schematest.WithProperty("currencyName", schematest.New("string")),
					schematest.WithRequired("value", "currency", "currencyName"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"currency":     "HTG",
					"currencyName": "Haiti Gourde",
					"value":        int64(98999),
				}, v)
			},
		},
		{
			name: "price object using amount",
			req: &Request{
				Path: []string{"price"},
				Schema: schematest.New("object",
					schematest.WithProperty("amount", schematest.New("number")),
					schematest.WithProperty("currency", schematest.New("string")),
					schematest.WithProperty("currencyName", schematest.New("string")),
					schematest.WithRequired("amount", "currency", "currencyName"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"currency":     "HTG",
					"currencyName": "Haiti Gourde",
					"amount":       750983.89,
				}, v)
			},
		},
		{
			name: "credit_card",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("credit_card", schematest.New("number")),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"credit_card": 4.28241936002255e+14,
				}, v)
			},
		},
		{
			name: "creditCard",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("creditCard", schematest.New("number")),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"creditCard": 4.28241936002255e+14,
				}, v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)
			Seed(1234567)

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}
