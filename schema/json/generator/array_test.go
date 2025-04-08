package generator

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema/schematest"
	"testing"
)

func TestArray(t *testing.T) {
	testcases := []struct {
		name string
		req  *Request
		test func(t *testing.T, v interface{}, err error)
	}{
		{
			name: "array with example",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithItems("object",
						schematest.WithProperty("firstname", schematest.New("string")),
						schematest.WithProperty("lastname", schematest.New("string")),
						// unknown property for generator
						schematest.WithProperty("foo", schematest.New("string")),
						schematest.WithRequired("firstname", "lastname", "foo"),
					),
					schematest.WithExamples([]any{
						map[string]any{
							"firstname": "John",
							"lastname":  "Doe",
							"foo":       "bar",
						},
					}),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					map[string]interface{}{
						"firstname": "Gabriel",
						"lastname":  "Adams",
						"foo":       "bar",
					},
					map[string]interface{}{
						"firstname": "Jayden",
						"lastname":  "Walker",
						"foo":       "bar",
					},
				}, v)
			},
		},
		{
			name: "array with example and items with example",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithMinItems(5),
					schematest.WithItems("object",
						schematest.WithProperty("firstname", schematest.New("string")),
						schematest.WithProperty("lastname", schematest.New("string")),
						// unknown property for generator
						schematest.WithProperty("foo", schematest.New("string")),
						schematest.WithRequired("firstname", "lastname", "foo"),
						schematest.WithExamples(
							map[string]any{
								"firstname": "John",
								"lastname":  "Doe",
								"foo":       "yuh",
							},
							map[string]any{
								"firstname": "Mike",
								"lastname":  "Walker",
								"foo":       "zzz",
							},
						),
					),
					schematest.WithExamples([]any{
						map[string]any{
							"firstname": "John",
							"lastname":  "Doe",
							"foo":       "bar",
						},
					}),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					map[string]interface{}{
						"firstname": "Zoey",
						"lastname":  "Nguyen",
						"foo":       "zzz",
					},
					map[string]interface{}{
						"firstname": "Ella",
						"lastname":  "Torres",
						"foo":       "bar",
					},
					map[string]interface{}{
						"firstname": "Aria",
						"lastname":  "Johnson",
						"foo":       "bar",
					},
					map[string]interface{}{
						"firstname": "Audrey",
						"lastname":  "Taylor",
						"foo":       "bar",
					},
					map[string]interface{}{
						"firstname": "Liam",
						"lastname":  "Gonzalez",
						"foo":       "bar",
					},
				}, v)
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
