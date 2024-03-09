package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/json/schematest"
	"testing"
)

func TestUser(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "person any",
			req:  &Request{Names: []string{"person"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"email":     "arelyzemlak@muller.biz",
					"firstname": "Shanelle",
					"gender":    "male",
					"lastname":  "Wehner",
				}, v)
			},
		},
		{
			name: "person name",
			req:  &Request{Names: []string{"person", "name"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "Shanelle Wehner", v)
			},
		},
		{
			name: "person with schema",
			req: &Request{
				Names: []string{"person"},
				Schema: schematest.New("object",
					schematest.WithProperty("firstname", schematest.New("string")),
					schematest.WithProperty("lastname", schematest.New("string")),
					schematest.WithProperty("gender", schematest.New("string")),
					schematest.WithProperty("sex", schematest.New("string")),
					schematest.WithProperty("email", schematest.New("string", schematest.WithFormat("email"))),
					schematest.WithProperty("phone", schematest.New("string")),
					schematest.WithProperty("contact", nil),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"firstname": "Shanelle",
					"lastname":  "Wehner",
					"gender":    "male",
					"sex":       "male",
					"email":     "brookmuller@yundt.info",
					"phone":     "6489301180",
					"contact":   map[string]interface{}{"email": "darronbartell@lowe.net", "phone": "5735018614"},
				}, v)
			},
		},
		{
			name: "persons as array",
			req: &Request{
				Names: []string{"persons"},
				Schema: schematest.New("array",
					schematest.WithMinItems(4),
					schematest.WithItems("object", schematest.WithProperty("name", nil)),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					map[string]interface{}{"name": "Jennifer Cruickshank"},
					map[string]interface{}{"name": "Arely Zemlak"},
					map[string]interface{}{"name": "Chase Yundt"},
					map[string]interface{}{"name": "Porter Kiehn"}},
					v)
			},
		},
		{
			name: "persons as any",
			req: &Request{
				Names: []string{"persons"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					map[string]interface{}{
						"email":     "brookmuller@yundt.info",
						"firstname": "Jennifer",
						"gender":    "male",
						"lastname":  "Cruickshank"},
					map[string]interface{}{
						"email":     "modestowiza@farrell.name",
						"firstname": "Thad",
						"gender":    "male",
						"lastname":  "Gerhold",
					},
				}, v)
			},
		},
		{
			name: "contact any",
			req:  &Request{Names: []string{"contact"}},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"email": "nolankuhlman@wiza.name", "phone": "8029109364"}, v)
			},
		},
		{
			name: "creditcard",
			req: &Request{
				Names: []string{"creditcard"},
				Schema: schematest.New("object",
					schematest.WithProperty("type", nil),
					schematest.WithProperty("number", nil),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"number": "6504859109364908", "type": "American Express"}, v)
			},
		},
		{
			name: "person with address, username and creditcard",
			req: &Request{
				Names: []string{"person"},
				Schema: schematest.New("object",
					schematest.WithProperty("username", schematest.New("string")),
					schematest.WithProperty("address", nil),
					schematest.WithProperty("creditcard", nil),
				),
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
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}
