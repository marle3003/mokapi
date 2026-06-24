package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestPerson(t *testing.T) {
	// tests depends on current year so without this, all tests will break in next year
	isDateString := func(t *testing.T, s any) {
		_, err := time.Parse("2006-01-02", s.(string))
		require.NoError(t, err)
	}

	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "person any",
			req: &Request{
				Path: []string{"person"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"firstname": "Aria",
					"lastname":  "Hernandez",
					"gender":    "female",
					"email":     "aria.hernandez@salesmorph.name",
				}, v)
			},
		},
		{
			name: "person object without properties",
			req: &Request{
				Path:   []string{"person"},
				Schema: schematest.New("object"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"firstname": "Aria",
					"gender":    "female",
					"lastname":  "Hernandez",
					"email":     "aria.hernandez@salesmorph.name",
				}, v)
			},
		},
		{
			name: "person name",
			req: &Request{
				Path: []string{"person"},
				Schema: schematest.New("object",
					schematest.WithProperty("name", nil),
					schematest.WithRequired("name"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Aria Hernandez"}, v)
			},
		},
		{
			name: "person dependent fields no gender field",
			req: &Request{
				Path: []string{"person"},
				Schema: schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithProperty("firstname", schematest.New("string")),
					schematest.WithProperty("lastname", schematest.New("string")),
					schematest.WithRequired("name", "firstname", "lastname"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"firstname": "Aria",
					"lastname":  "Hernandez",
					"name":      "Aria Hernandez",
				}, v)
			},
		},
		{
			name: "person dependent fields with gender field",
			req: &Request{
				Path: []string{"person"},
				Schema: schematest.New("object",
					schematest.WithProperty("firstname", schematest.New("string")),
					schematest.WithProperty("lastname", schematest.New("string")),
					schematest.WithProperty("sex", schematest.New("string")),
					schematest.WithRequired("firstname", "lastname", "sex"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"firstname": "Emily",
					"lastname":  "Jones",
					"sex":       "female",
				}, v)
			},
		},
		{
			name: "person with schema",
			req: &Request{
				Path: []string{"person"},
				Schema: schematest.New("object",
					schematest.WithProperty("firstname", schematest.New("string")),
					schematest.WithProperty("lastname", schematest.New("string")),
					schematest.WithProperty("gender", schematest.New("string")),
					schematest.WithProperty("sex", schematest.New("string")),
					schematest.WithProperty("email", schematest.New("string", schematest.WithFormat("email"))),
					schematest.WithProperty("phone", schematest.New("string")),
					schematest.WithProperty("username", schematest.New("string")),
					schematest.WithProperty("contact", nil),
					schematest.WithRequired("firstname", "lastname", "gender", "sex", "email", "phone", "username", "contact"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"contact": map[string]interface{}{
						"email": "riley.jones@vicee-markets.com",
						"phone": "+15144065319"},
					"email":     "riley.jones@groupschemas.com",
					"firstname": "Riley",
					"gender":    "female",
					"lastname":  "Jones",
					"phone":     "+20524193600225",
					"sex":       "female",
					"username":  "rjones",
				}, v)
			},
		},
		{
			name: "persons as array",
			req: &Request{
				Path: []string{"persons"},
				Schema: schematest.New("array",
					schematest.WithMinItems(4),
					schematest.WithItems("object",
						schematest.WithProperty("name", nil),
						schematest.WithRequired("name"),
					),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					map[string]any{"name": "Emily Nelson"},
					map[string]any{"name": "Aiden Garcia"},
					map[string]any{"name": "Sebastian Wright"},
					map[string]any{"name": "Sophia Wilson"},
				},
					v)
			},
		},
		{
			name: "persons as any",
			req: &Request{
				Path:   []string{"persons"},
				Schema: schematest.NewTypes(nil, schematest.WithMinItems(1)),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					map[string]any{
						"email":     "emily.nelson@directend-to-end.com",
						"firstname": "Emily",
						"gender":    "female",
						"lastname":  "Nelson",
					},
				}, v)
			},
		},
		{
			name: "contact any",
			req: &Request{
				Path: []string{"contact"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"email": "kbryant57@berge.biz", "phone": "+992824193"}, v)
			},
		},
		{
			name: "phone any schema",
			req: &Request{
				Path: []string{"phone"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "+992824193", v)
			},
		},
		{
			name: "phone schema string",
			req: &Request{
				Path:   []string{"phone"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "+992824193", v)
			},
		},
		{
			name: "phone schema string",
			req: &Request{
				Path:   []string{"notificationPhoneNumber"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "+992824193", v)
			},
		},
		{
			name: "phone but expect object",
			req: &Request{
				Path:   []string{"phone"},
				Schema: schematest.New("boolean"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, true, v)
			},
		},
		{
			name: "windowsUserName",
			req: &Request{
				Path:   []string{"windowsUserName"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "ahernandez", v)
			},
		},
		{
			name: "person data without person in parent name",
			req: &Request{
				Path: []string{"individual"},
				Schema: schematest.New("object",
					schematest.WithProperty("firstname", schematest.New("string")),
					schematest.WithProperty("lastname", schematest.New("string")),
					schematest.WithRequired("firstname", "lastname"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"firstname": "Aria", "lastname": "Hernandez"}, v)
			},
		},
		{
			name: "birthday",
			req: &Request{
				Path:   []string{"person", "birthday"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "birthDate",
			req: &Request{
				Path:   []string{"person", "birthDate"},
				Schema: schematest.New("string"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				isDateString(t, v)
			},
		},
		{
			name: "title depends on firstname",
			req: &Request{
				Path: []string{"person"},
				Schema: schematest.New("array",
					schematest.WithMinItems(5),
					schematest.WithItems("object",
						schematest.WithProperty("firstname", schematest.New("string")),
						schematest.WithProperty("title", schematest.New("string")),
						schematest.WithRequired("firstname", "title"),
					),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					map[string]any{"firstname": "Emily", "title": "Mrs."},
					map[string]any{"firstname": "Aiden", "title": "Mx."},
					map[string]any{"firstname": "Sebastian", "title": "Rev."},
					map[string]any{"firstname": "Sophia", "title": "Dr."},
					map[string]any{"firstname": "John", "title": "Prof."},
				}, v)
			},
		},
		{
			name: "username, firstname, lastname and sex",
			req: &Request{
				Path: []string{"person"},
				Schema: schematest.New("array",
					schematest.WithMinItems(5),
					schematest.WithItems("object",
						schematest.WithProperty("firstname", schematest.New("string")),
						schematest.WithProperty("lastname", schematest.New("string")),
						schematest.WithProperty("sex", schematest.New("string")),
						schematest.WithProperty("username", schematest.New("string")),
						schematest.WithProperty("alias", schematest.New("string")),
						schematest.WithRequired("firstname", "lastname", "sex", "username", "alias"),
					),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{
					map[string]interface{}{"alias": "L. Jackson", "firstname": "Leah", "lastname": "Jackson", "sex": "female", "username": "ljackson"},
					map[string]interface{}{"alias": "M. Hernandez", "firstname": "Mia", "lastname": "Hernandez", "sex": "female", "username": "mhernandez"},
					map[string]interface{}{"alias": "V. Carter", "firstname": "Violet", "lastname": "Carter", "sex": "female", "username": "vcarter"},
					map[string]interface{}{"alias": "C. Jones", "firstname": "Camila", "lastname": "Jones", "sex": "female", "username": "cjones"},
					map[string]interface{}{"alias": "W. Anderson", "firstname": "Wyatt", "lastname": "Anderson", "sex": "male", "username": "wanderson"},
				}, v)
			},
		},
		{
			name: "person fullname",
			req: &Request{
				Path: []string{"person"},
				Schema: schematest.New("object",
					schematest.WithProperty("fullName", nil),
					schematest.WithProperty("firstname", schematest.New("string")),
					schematest.WithRequired("name"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"firstname": "Aria", "fullName": "Aria Hernandez", "name": "Aria Hernandez"}, v)
			},
		},

		{
			name: "detect person domain",
			req: &Request{
				Path: []string{"test"},
				Schema: schematest.New("object",
					schematest.WithProperty("name", nil),
					schematest.WithProperty("firstName", nil),
					schematest.WithRequired("name", "firstName"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"firstName": "Aria", "name": "Aria Hernandez"}, v)
			},
		},
		{
			name: "middle name as firstName2",
			req: &Request{
				Path: []string{"person"},
				Schema: schematest.New("object",
					schematest.WithProperty("firstName2", nil),
					schematest.WithProperty("name", nil),
					schematest.WithRequired("name", "firstName2"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"firstName2": "Drew", "name": "Emily Drew Nelson"}, v)
			},
		},
		{
			name: "personAliases",
			req: &Request{
				Path: []string{"personAliases"},
				Schema: schematest.New("array",
					schematest.WithItems("string"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{"A. Hernandez", "A. H."}, v)
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
