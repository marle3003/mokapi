package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schematest"
	"testing"
)

func TestCommerce(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "creditcard",
			req: &Request{
				Path: Path{
					&PathElement{
						Name: "creditcard",
						Schema: schematest.NewRef("object",
							schematest.WithProperty("type", nil),
							schematest.WithProperty("number", nil),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"number": "6504859109364908", "type": "American Express"}, v)
			},
		},
		{
			name: "creditcardtype",
			req: &Request{
				Path: Path{
					&PathElement{
						Name: "creditcardtype",
						Schema: schematest.NewRef("string",
							schematest.WithMaxLength(4),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "F", v)
			},
		},
		{
			name: "person with address, username and creditcard",
			req: &Request{
				Path: Path{
					&PathElement{
						Name: "person",
						Schema: schematest.NewRef("object",
							schematest.WithProperty("username", schematest.New("string")),
							schematest.WithProperty("address", nil),
							schematest.WithProperty("creditcard", nil),
						),
					},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"address": map[string]interface{}{
						"address":   "64893 South Harborshaven, St. Paul, South Dakota 73501",
						"city":      "St. Paul",
						"country":   "Martinique",
						"latitude":  -27.033468,
						"longitude": -154.420938,
						"state":     "South Dakota",
						"street":    "64893 South Harborshaven",
						"zip":       "73501",
					},
					"username": "Lockman7291",
					"creditcard": map[string]interface{}{
						"type":   "Diners Club",
						"number": "2298465535548962",
						"cvv":    "367",
						"exp":    "01/31",
					},
				}, v)
			},
		},
		{
			name: "department name",
			req: &Request{
				Path: Path{
					&PathElement{Name: "department"},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Talent Acquisition", v)
			},
		},
		{
			name: "company name",
			req: &Request{
				Path: Path{
					&PathElement{Name: "company"},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Crestway Enterprises", v)
			},
		},
		{
			name: "job title",
			req: &Request{
				Path: Path{
					&PathElement{Name: "jobTitle"},
				},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Marketing Manager", v)
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
