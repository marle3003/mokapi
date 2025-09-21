package generator

import (
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
			name: "minItems and maxItems",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithItems("string"),
					schematest.WithMinItems(4),
					schematest.WithMaxItems(6),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{"lx0+fjywXKo", "jxkDng", "hbEO6wpu", "qosamhfi", "JUOvtsQ WavQ", ""}, v)
			},
		},
		{
			name: "minItems > maxItems",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithMinItems(3),
					schematest.WithMaxItems(2),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "minItems must be less than maxItems")
			},
		},
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
		{
			name: "contains",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithContains(schematest.New("string")),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{
					"qwCrwMfkOjo", "gPSz",
				}, v)
			},
		},
		{
			name: "contains with minContains",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithContains(schematest.New("string")),
					schematest.WithMinContains(3),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{
					"qwCrwMfkOjo", int64(6259110876194170218), "gPSz", "gPNseoOLAIqos", "qa6WoJUOvts",
				}, v)
			},
		},
		{
			name: "contains with maxContains",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithContains(schematest.New("string")),
					schematest.WithMaxContains(3),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{
					"qwCrwMfkOjo", "gPSz",
				}, v)
			},
		},
		{
			name: "contains with minContains but maxItems is lower",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithContains(schematest.New("string")),
					schematest.WithMinContains(3),
					schematest.WithMaxItems(2),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "minContains must be less than maxItems")
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
