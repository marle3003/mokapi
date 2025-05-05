package generator

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestObject(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "object",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", nil),
					schematest.WithRequired("name"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ink"}, v)
			},
		},
		{
			name: "no type but properties",
			req: &Request{
				Schema: schematest.NewTypes(nil,
					schematest.WithProperty("name", nil),
					schematest.WithRequired("name"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ink"}, v)
			},
		},
		{
			name: "no additional properties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithRequired("name"),
					schematest.WithFreeForm(false),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ink"}, v)
			},
		},
		{
			name: "with additional properties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithRequired("name"),
					schematest.WithFreeForm(true),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ink"}, v)
			},
		},
		{
			name: "with additional properties specific",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithAdditionalProperties(
						schematest.New("object",
							schematest.WithProperty("age", schematest.New("integer")),
							schematest.WithProperty("gender", schematest.New("string")),
							schematest.WithRequired("age", "gender"),
						),
					),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"collection": map[string]interface{}{
							"age": int64(5), "gender": "male",
						},
						"comb": map[string]interface{}{
							"age": int64(85), "gender": "male",
						},
						"company": map[string]interface{}{
							"age": int64(41), "gender": "male",
						},
						"luck": map[string]interface{}{
							"age": int64(10), "gender": "female",
						},
						"person": map[string]interface{}{
							"age": int64(55), "gender": "male",
						},
						"problem": map[string]interface{}{
							"age": int64(53), "gender": "female",
						},
						"sunshine": map[string]interface{}{
							"age": int64(51), "gender": "female",
						},
					},

					v)
			},
		},
		{
			name: "dictionary with min and max length",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithAdditionalProperties(
						schematest.New("string"),
					),
					schematest.WithMinProperties(10),
					schematest.WithMaxProperties(12),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Len(t, v, 11)
				require.Equal(t,
					map[string]interface{}{"brace": "ne", "chapter": "cILXzNQ", "collection": "JZGR", "comb": "q", "company": "gPSz", "life": "BYxST", "luck": "qa6WoJUOvts", "person": "NWavQeozIe", "problem": "sgzs", "string": "ticuuBCbV0biw0", "sunshine": "gPNseoOLAIqos"},
					v)
			},
		},
		{
			name: "object no properties",
			req: &Request{
				Schema: schematest.New("object"),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"brace":      9.489807638481764e+307,
						"collection": int64(2116312089470753225),
						"comb":       false,
						"company":    "Redfin",
						"luck":       int64(-7574890634918414754),
						"person": map[string]interface{}{
							"email":     "oliver.nelson@globalfacilitate.com",
							"firstname": "Oliver",
							"gender":    "male",
							"lastname":  "Nelson",
						},
						"problem":  true,
						"sunshine": true,
					},
					v)
			},
		},
		{
			name: "object no properties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithExamples(map[string]interface{}{"foo": "bar"}),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"foo": "bar",
					},
					v)
			},
		},
		{
			name: "fallback to example",
			req: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("foo", schematest.New("string")),
					schematest.WithExamples(map[string]interface{}{"foo": "bar"}),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"foo": "bar",
					},
					v)
			},
		},
		{
			name: "if-then",
			req: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("country", schematest.New("string")),
					schematest.WithIf(schematest.NewTypes(nil,
						schematest.WithProperty("country", schematest.New("string",
							schematest.WithConst("Slovenia")),
						),
					)),
					schematest.WithThen(schematest.NewTypes(nil,
						schematest.WithProperty("postal_code", schematest.New("string",
							schematest.WithPattern("[0-9]{5}(-[0-9]{4})?"))),
						schematest.WithRequired("postal_code"),
					)),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"country":     "Slovenia",
						"postal_code": "29109",
					},
					v)
			},
		},
		{
			name: "if-else",
			req: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("country", schematest.New("string")),
					schematest.WithIf(schematest.NewTypes(nil,
						schematest.WithProperty("country", schematest.New("string",
							schematest.WithConst("Canada")),
						),
					)),
					schematest.WithElse(schematest.NewTypes(nil,
						schematest.WithProperty("postal_code", schematest.New("string",
							schematest.WithPattern("[A-Z][0-9][A-Z] [0-9][A-Z][0-9]")),
						),
					)),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"country":     "Slovenia",
						"postal_code": "C0O 9F0",
					},
					v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			Seed(1234567)

			v, err := New(tc.req)
			tc.test(t, v, err)
		})
	}
}
