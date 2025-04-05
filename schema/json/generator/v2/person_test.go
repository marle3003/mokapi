package v2

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
					"email":     "arelyzemlak@muller.biz",
					"firstname": "Shanelle",
					"gender":    "male",
					"lastname":  "Wehner",
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
					"email":     "arelyzemlak@muller.biz",
					"firstname": "Shanelle",
					"gender":    "male",
					"lastname":  "Wehner",
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
				require.Equal(t, map[string]interface{}{"name": "Shanelle Wehner"}, v)
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
					"lastname":  "Cruickshank",
					"name":      "Zoey Cruickshank",
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
					"lastname":  "Lockman",
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
					schematest.WithProperty("contact", nil),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"contact": map[string]interface{}{
						"email": "ethan.lockman@legacyb2b.net",
						"phone": "+31986146",
					},
					"email":     "ethan.lockman@productenvisioneer.io",
					"firstname": "Ethan",
					"gender":    "male",
					"lastname":  "Lockman",
					"phone":     "+737793648930118",
					"sex":       "male",
				}, v)
			},
		},
		//{
		//	name: "persons as array",
		//	req: &Request{
		//		Names: []string{"persons"},
		//		Schema: schematest.New("array",
		//			schematest.WithMinItems(4),
		//			schematest.WithItems("object", schematest.WithProperty("name", nil)),
		//		),
		//	},
		//	test: func(t *testing.T, v interface{}, err error) {
		//		require.NoError(t, err)
		//		require.Equal(t, []interface{}{
		//			map[string]interface{}{"name": "Jennifer Cruickshank"},
		//			map[string]interface{}{"name": "Arely Zemlak"},
		//			map[string]interface{}{"name": "Chase Yundt"},
		//			map[string]interface{}{"name": "Porter Kiehn"}},
		//			v)
		//	},
		//},
		//{
		//	name: "persons as any",
		//	req: &Request{
		//		Names: []string{"persons"},
		//	},
		//	test: func(t *testing.T, v interface{}, err error) {
		//		require.NoError(t, err)
		//		require.Equal(t, []interface{}{
		//			map[string]interface{}{
		//				"email":     "brookmuller@yundt.info",
		//				"firstname": "Jennifer",
		//				"gender":    "male",
		//				"lastname":  "Cruickshank"},
		//			map[string]interface{}{
		//				"email":     "modestowiza@farrell.name",
		//				"firstname": "Thad",
		//				"gender":    "male",
		//				"lastname":  "Gerhold",
		//			},
		//		}, v)
		//	},
		//},
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
				Schema: schematest.New("object"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"actor": "jfjqa6WoJUOv",
						"block": int64(7868624871047129135),
						"crew":  true,
						"envy":  "Q WavQeozIeDmm",
						"part":  "10256636970",
						"party": map[string]interface{}{
							"happiness": "35548959830",
							"intelligence": []interface{}{
								int64(7479033884925130368), true,
							},
						},
						"pod":    false,
						"school": int64(8114684127683409738),
						"train":  "44709213031",
					},
					v)
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
				require.Equal(t, "Lockman7291", v)
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
				require.Equal(t, map[string]interface{}{"firstname": "Zoey", "lastname": "Cruickshank"}, v)
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
