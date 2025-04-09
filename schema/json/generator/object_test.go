package generator

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
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
						"brace":      false,
						"collection": 8.515060348610526e+307,
						"comb":       "eyNWavQeo",
						"company":    "OjojxkDngP",
						"luck":       "eoOLAIqosamhfi",
						"person": map[string]interface{}{
							"email":     "camila.white@corporatebleeding-edge.org",
							"firstname": "Camila",
							"gender":    "female",
							"lastname":  "White",
						},
						"problem":  int64(4212695239388227044),
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
							schematest.WithConst("Bahrain")),
						),
					)),
					schematest.WithThen(schematest.NewTypes(nil,
						schematest.WithProperty("postal_code", schematest.New("string",
							schematest.WithPattern("[0-9]{5}(-[0-9]{4})?"))),
					)),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"country":     "Bahrain",
						"postal_code": "10936",
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
						"country":     "Bahrain",
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

func TestLoop(t *testing.T) {
	testcases := []struct {
		name string
		req  func() *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "no type but properties",
			req: func() *Request {
				return &Request{
					Schema: schematest.NewTypes(nil,
						schematest.WithProperty("name", nil),
						schematest.WithRequired("name"),
					),
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ink"}, v)
			},
		},
		{
			name: "object with loop",
			req: func() *Request {
				s := schematest.New("object")
				s.Properties = &schema.Schemas{}
				s.Properties.Set("loop", s)

				return &Request{
					Schema: s,
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "recursion in object path found but schema does not allow null: schema type=object properties=[loop]")
			},
		},
		{
			name: "object with loop and nullable",
			req: func() *Request {
				s := schematest.NewTypes([]string{"object", "null"})
				s.Properties = &schema.Schemas{}
				s.Properties.Set("loop", s)

				return &Request{
					Schema: s,
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"loop": nil}, v)
			},
		},
		{
			name: "object with properties that contains loop",
			req: func() *Request {
				loop := schematest.NewTypes([]string{"object", "null"}, schematest.WithRequired("loop"))
				loop.Properties = &schema.Schemas{}
				loop.Properties.Set("loop", loop)
				s := schematest.New("object",
					schematest.WithProperty("loop1", loop),
					schematest.WithProperty("loop2", loop),
					schematest.WithProperty("loop3", loop),
					schematest.WithRequired("loop1", "loop2", "loop3"),
				)

				return &Request{
					Schema: s,
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"loop1": map[string]interface{}{
						"loop": nil,
					},
					"loop2": map[string]interface{}{
						"loop": nil,
					},
					"loop3": map[string]interface{}{
						"loop": nil,
					},
				},
					v)
			},
		},
		{
			name: "loop with array",
			req: func() *Request {
				loop := schematest.New("object")
				loop.Properties = &schema.Schemas{}
				loop.Properties.Set("array", &schema.Schema{
					Type:  schema.Types{"array"},
					Items: loop,
				})

				return &Request{
					Schema: loop,
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "recursion in object path found but schema does not allow null: schema type=object properties=[array]")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			Seed(1234567)

			v, err := New(tc.req())
			tc.test(t, v, err)
		})
	}
}
