package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
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
					"firstname": "Zoey",
					"lastname":  "Nguyen",
					"gender":    "female",
					"email":     "zoey.nguyen@internationaldeliver.info",
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
					"firstname": "Zoey",
					"gender":    "female",
					"lastname":  "Nguyen",
					"email":     "zoey.nguyen@internationaldeliver.info",
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
				require.Equal(t, map[string]interface{}{"name": "Zoey Nguyen"}, v)
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
					"firstname": "Zoey",
					"lastname":  "Nguyen",
					"name":      "Zoey Nguyen",
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
					"firstname": "Gabriel",
					"lastname":  "Clark",
					"sex":       "male",
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
						"email": "anthony.clark@legacyb2b.net",
						"phone": "+61350186146"},
					"email":     "anthony.clark@dynamiccommunities.io",
					"firstname": "Anthony",
					"gender":    "male",
					"lastname":  "Clark",
					"phone":     "+61810936489301",
					"sex":       "male",
					"username":  "aclark",
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
					map[string]any{"name": "Gabriel Adams"},
					map[string]any{"name": "Ella Torres"},
					map[string]any{"name": "Penelope Jackson"},
					map[string]any{"name": "Michael Carter"},
					map[string]any{"name": "Jackson Green"},
					map[string]any{"name": "Liam Gonzalez"},
				},
					v)
			},
		},
		{
			name: "persons as any",
			req: &Request{
				Path: []string{"persons"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					map[string]interface{}{
						"email":     "gabriel.adams@futurecultivate.biz",
						"firstname": "Gabriel",
						"gender":    "male",
						"lastname":  "Adams",
					},
					map[string]interface{}{
						"email":     "penelope.jackson@directembrace.biz",
						"firstname": "Penelope",
						"gender":    "female",
						"lastname":  "Jackson",
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
				require.Equal(t, map[string]interface{}{"email": "porterkiehn@gerhold.name", "phone": "+28829109"}, v)
			},
		},
		{
			name: "phone any schema",
			req: &Request{
				Path: []string{"phone"},
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, "+28829109", v)
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
				require.Equal(t, "+28829109", v)
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
				require.Equal(t, "+28829109", v)
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
				require.Equal(t, false, v)
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
				require.Equal(t, "znguyen", v)
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
				require.Equal(t, map[string]interface{}{"firstname": "Zoey", "lastname": "Nguyen"}, v)
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
					map[string]any{"firstname": "Gabriel", "title": "Dr."},
					map[string]any{"firstname": "Ella", "title": "Miss"},
					map[string]any{"firstname": "Penelope", "title": "Ms."},
					map[string]any{"firstname": "Michael", "title": "Rev."},
					map[string]any{"firstname": "Jackson", "title": "Mx."},
					map[string]any{"firstname": "Liam", "title": "Prof."},
					map[string]any{"firstname": "Thomas", "title": "Mr."},
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
					map[string]any{"alias": "H. Robinson", "firstname": "Hudson", "lastname": "Robinson", "sex": "male", "username": "hrobinson"},
					map[string]any{"alias": "J. Nelson", "firstname": "Jayden", "lastname": "Nelson", "sex": "male", "username": "jnelson"},
					map[string]any{"alias": "A. Walker", "firstname": "Aria", "lastname": "Walker", "sex": "female", "username": "awalker"},
					map[string]any{"alias": "M. Johnson", "firstname": "Maverick", "lastname": "Johnson", "sex": "male", "username": "mjohnson"},
					map[string]any{"alias": "E. Mitchell", "firstname": "Emilia", "lastname": "Mitchell", "sex": "female", "username": "emitchell"},
					map[string]any{"alias": "E. Lopez", "firstname": "Elizabeth", "lastname": "Lopez", "sex": "female", "username": "elopez"},
					map[string]any{"alias": "I. Martinez", "firstname": "Isabella", "lastname": "Martinez", "sex": "female", "username": "imartinez"},
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
				require.Equal(t, map[string]interface{}{"firstname": "Zoey", "fullName": "Zoey Nguyen", "name": "Zoey Nguyen"}, v)
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
				require.Equal(t, map[string]interface{}{"firstName": "Zoey", "name": "Zoey Nguyen"}, v)
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
				require.Equal(t, map[string]interface{}{"firstName2": "Ann", "name": "Stella Ann Adams"}, v)
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
				require.Equal(t, []interface{}{"Z. Nguyen", "Z. N."}, v)
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
