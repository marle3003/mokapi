package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestCircularReference(t *testing.T) {
	testcases := []struct {
		name string
		req  func() *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "array",
			req: func() *Request {
				s := &schema.Schema{Type: schema.Types{"object"}}
				list := &schema.Schema{Type: schema.Types{"array"}, Items: s}

				s.Properties = &schema.Schemas{}
				s.Properties.Set("list", list)

				return &Request{
					Schema: s,
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{"list": []interface{}{}}, v)
			},
		},
		{
			name: "object",
			req: func() *Request {
				s := &schema.Schema{Type: schema.Types{"object"}}

				s.Properties = &schema.Schemas{}
				s.Properties.Set("loop", s)

				return &Request{
					Schema: s,
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{}, v)
			},
		},
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
			name: "object with required loop",
			req: func() *Request {
				s := schematest.New("object")
				s.Properties = &schema.Schemas{}
				s.Properties.Set("loop", s)
				s.Required = append(s.Required, "loop")

				return &Request{
					Schema: s,
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "recursion in object path found but schema does not allow null: schema type=object properties=[loop] required=[loop]")
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
			name: "loop with array with minItems=1",
			req: func() *Request {
				loop := schematest.New("object", schematest.WithRequired("array"))
				loop.Properties = &schema.Schemas{}
				minItems := 1
				loop.Properties.Set("array", &schema.Schema{
					Type:     schema.Types{"array"},
					Items:    loop,
					MinItems: &minItems,
				})

				return &Request{
					Schema: loop,
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "recursion in object path found but schema does not allow null: schema type=object properties=[array] required=[array]")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)
			Seed(1234567)

			v, err := New(tc.req())
			tc.test(t, v, err)
		})
	}
}
