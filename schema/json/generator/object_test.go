package generator

import (
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
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
			name: "with one property is required",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", nil),
					schematest.WithProperty("foo", nil),
					schematest.WithRequired("name"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ink", "foo": int64(5155350187252080587)}, v)
			},
		},
		{
			name: "with required and maxProperties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", nil),
					schematest.WithProperty("foo", nil),
					schematest.WithProperty("bar", nil),
					schematest.WithProperty("yuh", nil),
					schematest.WithProperty("zzz", nil),
					schematest.WithRequired("name"),
					schematest.WithMaxProperties(3),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"name": "Ink", "foo": int64(5155350187252080587), "zzz": false}, v)
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
			name: "with additional properties true should not create additional properties because users do not expect this",
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
			name: "with additional properties string but maxProperties=1 should not create additional properties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("name", schematest.New("string")),
					schematest.WithRequired("name"),
					schematest.WithAdditionalProperties(schematest.New("string")),
					schematest.WithMaxProperties(1),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Velvet"}, v)
			},
		},
		{
			name: "additionalProperties string",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithAdditionalProperties(schematest.New("string")),
					schematest.WithMaxProperties(3),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"company": "+fjywXKo", "luck": "hbEO6wpu", "sunshine": "jxkDng",
					},
					v)
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
					schematest.WithMaxProperties(3),
				),
			},
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{
					"company": map[string]any{
						"age": int64(81), "gender": "female",
					},
					"luck": map[string]any{
						"age": int64(52), "gender": "female",
					},
					"sunshine": map[string]any{
						"age": int64(69), "gender": "female",
					},
				}, v)
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
			name: "object no properties with examples",
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
			name: "object no properties with required properties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithRequired("foo", "bar"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"foo": int64(5155350187252080587),
						"bar": int64(-4567356855949603266),
					},
					v)
			},
		},
		{
			name: "object with properties and additional required properties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithProperty("foo", schematest.New("string")),
					schematest.WithRequired("bar"),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"foo": "FqwCrwMfkOjojx",
						"bar": 1.6043524827049678e+308,
					},
					v)
			},
		},
		{
			name: "patternProperties",
			req: &Request{
				Schema: schematest.New("object",
					schematest.WithPatternProperty("^S_", schematest.New("string")),
					schematest.WithPatternProperty("^I_", schematest.New("integer")),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"I_8wEl7": int64(1725511503074869948),
						"S_Z":     "fkOjoj",
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
						schematest.WithRequired("country"),
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
						"postal_code": "80291-0936",
					},
					v)
			},
		},
		{
			name: "if but maxProperties should return empty object",
			req: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("country", schematest.New("string")),
					schematest.WithIf(schematest.NewTypes(nil,
						schematest.WithProperty("country", schematest.New("string",
							schematest.WithConst("Slovenia")),
						),
						schematest.WithRequired("country"),
					)),
					schematest.WithThen(schematest.NewTypes(nil,
						schematest.WithProperty("postal_code", schematest.New("string",
							schematest.WithPattern("[0-9]{5}(-[0-9]{4})?"))),
						schematest.WithRequired("postal_code"),
					)),
					schematest.WithMaxProperties(1),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{},
					v)
			},
		},
		{
			name: "if with maxProperties but country is required",
			req: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("country", schematest.New("string",
						// set const value to force the error
						schematest.WithConst("Slovenia"),
					)),
					schematest.WithIf(schematest.NewTypes(nil,
						schematest.WithProperty("country", schematest.New("string",
							schematest.WithConst("Slovenia")),
						),
						schematest.WithRequired("country"),
					)),
					schematest.WithThen(schematest.NewTypes(nil,
						schematest.WithProperty("postal_code", schematest.New("string",
							schematest.WithPattern("[0-9]{5}(-[0-9]{4})?"))),
						schematest.WithRequired("postal_code"),
					)),
					schematest.WithRequired("country"),
					schematest.WithMaxProperties(1),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "failed to generate valid object: reached attempt limit (10) caused by: conditional schema could not be applied: reached attempt limit (10) caused by: reached maximum of value maxProperties=1")
			},
		},
		{
			name: "if with maxProperties but country is required first try with Slovenia (random value) fails",
			req: &Request{
				Path: []string{"address"},
				Schema: schematest.New("object",
					schematest.WithProperty("country", schematest.New("string")),
					schematest.WithIf(schematest.NewTypes(nil,
						schematest.WithProperty("country", schematest.New("string",
							schematest.WithConst("Slovenia")),
						),
						schematest.WithRequired("country"),
					)),
					schematest.WithThen(schematest.NewTypes(nil,
						schematest.WithProperty("postal_code", schematest.New("string",
							schematest.WithPattern("[0-9]{5}(-[0-9]{4})?"))),
						schematest.WithRequired("postal_code"),
					)),
					schematest.WithRequired("country"),
					schematest.WithMaxProperties(1),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]any{"country": "Tuvalu"}, v)
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
						schematest.WithRequired("country"),
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
