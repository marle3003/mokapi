package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestPerson(t *testing.T) {
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
				Path:   []string{"person"},
				Schema: schematest.New("object", schematest.WithProperty("name", nil)),
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
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"contact": map[string]interface{}{
						"email": "ethan.clark@legacyb2b.net",
						"phone": "+31986146",
					},
					"email":     "ethan.clark@productenvisioneer.io",
					"firstname": "Ethan",
					"gender":    "male",
					"lastname":  "Clark",
					"phone":     "+737793648930118",
					"sex":       "male",
					"username":  "eclark",
				}, v)
			},
		},
		{
			name: "persons as array",
			req: &Request{
				Path: []string{"persons"},
				Schema: schematest.New("array",
					schematest.WithMinItems(4),
					schematest.WithItems("object", schematest.WithProperty("name", nil)),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					map[string]interface{}{"name": "Gabriel Adams"},
					map[string]interface{}{"name": "Ella Torres"},
					map[string]interface{}{"name": "Penelope Jackson"},
					map[string]interface{}{"name": "Michael Carter"},
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
				require.Equal(t, "1970-08-26", v)
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
				require.Equal(t, "1970-08-26", v)
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
