package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestObject(t *testing.T) {
	testcases := []struct {
		name string
		req  func() *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "object",
			req: func() *Request {
				return &Request{
					Path: Path{
						&PathElement{Schema: schematest.New("object", schematest.WithProperty("name", nil))},
					},
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ink"}, v)
			},
		},
		{
			name: "no type but properties",
			req: func() *Request {
				return &Request{
					Path: Path{
						&PathElement{Schema: schematest.NewTypes(nil, schematest.WithProperty("name", nil))},
					},
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
					Path: Path{
						&PathElement{Schema: s},
					},
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
					Path: Path{
						&PathElement{Schema: s},
					},
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"loop": map[string]interface{}{"loop": nil}}, v)
			},
		},
		{
			name: "object with properties that contains loop",
			req: func() *Request {
				loop := schematest.NewTypes([]string{"object", "null"})
				loop.Properties = &schema.Schemas{}
				loop.Properties.Set("loop", loop)
				s := schematest.New("object",
					schematest.WithProperty("loop1", loop),
					schematest.WithProperty("loop2", loop),
					schematest.WithProperty("loop3", loop),
				)

				return &Request{
					Path: Path{
						&PathElement{Schema: s},
					},
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				// here first loop is already broken, because first level is not counted in the loop protection
				require.Equal(t, map[string]interface{}{
					"loop1": map[string]interface{}{"loop": nil},
					"loop2": map[string]interface{}{"loop": nil},
					"loop3": map[string]interface{}{"loop": nil},
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
					Path: Path{
						&PathElement{Schema: loop},
					},
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"array": []interface{}{
						map[string]interface{}{
							"array": []interface{}{},
						},
						map[string]interface{}{
							"array": []interface{}{},
						},
					},
				},
					v)
			},
		},
		{
			name: "no additional properties",
			req: func() *Request {
				return &Request{
					Path: Path{
						&PathElement{Schema: schematest.New("object",
							schematest.WithProperty("name", schematest.New("string")),
							schematest.WithFreeForm(false),
						)},
					},
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ink"}, v)
			},
		},
		{
			name: "with additional properties",
			req: func() *Request {
				return &Request{
					Path: Path{
						&PathElement{Schema: schematest.New("object",
							schematest.WithProperty("name", schematest.New("string")),
							schematest.WithFreeForm(true),
						)},
					},
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"name": "Ink"}, v)
			},
		},
		{
			name: "with additional properties specific",
			req: func() *Request {
				return &Request{
					Path: Path{
						&PathElement{Schema: schematest.New("object",
							schematest.WithAdditionalProperties(
								schematest.New("object",
									schematest.WithProperty("age", schematest.New("integer")),
									schematest.WithProperty("gender", schematest.New("string")),
									schematest.WithRequired("age", "gender"),
								),
							),
						)},
					},
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t,
					map[string]interface{}{
						"brace": map[string]interface{}{
							"age":    int64(-2231804065324102411),
							"gender": "female",
						},
						"child": map[string]interface{}{
							"age":    int64(-5996906719781834924),
							"gender": "male",
						},
						"comb": map[string]interface{}{
							"age":    int64(298783498632242168),
							"gender": "male",
						},
						"life": map[string]interface{}{
							"age":    int64(-7574890634918414754),
							"gender": "male",
						},
						"person": map[string]interface{}{
							"age":    int64(-5048116001741826297),
							"gender": "female",
						},
						"string": map[string]interface{}{
							"age":    int64(-7377528278928660358),
							"gender": "female",
						},
						"sunshine": map[string]interface{}{
							"age":    int64(-3652171958352792229),
							"gender": "male",
						},
					},
					v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.req())
			tc.test(t, v, err)
		})
	}
}
