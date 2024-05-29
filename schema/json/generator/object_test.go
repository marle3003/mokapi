package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/ref"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schematest"
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
						&PathElement{Schema: schematest.NewRef("object", schematest.WithProperty("name", nil))},
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
				s.Properties.Set("loop", &schema.Ref{Reference: ref.Reference{Ref: "#/components/schemas/loop"}, Value: s})

				return &Request{
					Path: Path{
						&PathElement{Schema: &schema.Ref{Value: s}},
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
				s.Properties.Set("loop", &schema.Ref{Reference: ref.Reference{Ref: "#/components/schemas/loop"}, Value: s})

				return &Request{
					Path: Path{
						&PathElement{Schema: &schema.Ref{Value: s}},
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
				loop.Properties.Set("loop", &schema.Ref{Reference: ref.Reference{Ref: "#/components/schemas/loop"}, Value: loop})
				s := schematest.New("object",
					schematest.WithProperty("loop1", loop),
					schematest.WithProperty("loop2", loop),
					schematest.WithProperty("loop3", loop),
				)

				return &Request{
					Path: Path{
						&PathElement{Schema: &schema.Ref{Value: s}},
					},
				}
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, map[string]interface{}{
					"loop1": map[string]interface{}{"loop": map[string]interface{}{"loop": nil}},
					"loop2": map[string]interface{}{"loop": map[string]interface{}{"loop": nil}},
					"loop3": map[string]interface{}{"loop": map[string]interface{}{"loop": nil}},
				},
					v)
			},
		},
		{
			name: "loop with array",
			req: func() *Request {
				loop := schematest.NewRef("object")
				loop.Value.Properties = &schema.Schemas{}
				loop.Value.Properties.Set("array", &schema.Ref{Value: &schema.Schema{
					Type:  schema.Types{"array"},
					Items: loop,
				}})

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
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gofakeit.Seed(1234567)

			v, err := New(tc.req())
			tc.test(t, v, err)
		})
	}
}
