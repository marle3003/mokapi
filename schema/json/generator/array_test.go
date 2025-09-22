package generator

import (
	"github.com/stretchr/testify/require"
	"mokapi/schema/json/schema"
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
				require.EqualError(t, err, "invalid schema: minItems must be less than maxItems")
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
			test: func(t *testing.T, v any, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{
					map[string]any{
						"firstname": "Gabriel",
						"foo":       "yuh",
						"lastname":  "Adams",
					}, map[string]any{
						"firstname": "Jayden",
						"foo":       "bar",
						"lastname":  "Walker",
					}, map[string]any{
						"firstname": "Olivia",
						"foo":       "bar",
						"lastname":  "Lopez",
					}, map[string]any{
						"firstname": "Jackson",
						"foo":       "zzz",
						"lastname":  "Green",
					}, map[string]any{
						"firstname": "Elizabeth",
						"foo":       "bar",
						"lastname":  "Martinez",
					}, map[string]any{
						"firstname": "Olivia",
						"foo":       "yuh",
						"lastname":  "Williams",
					}, map[string]any{
						"firstname": "Olivia",
						"foo":       "bar",
						"lastname":  "Jones",
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
				require.EqualError(t, err, "invalid schema: minContains must be less than maxItems")
			},
		},
		{
			name: "maxContains is reached",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithItems("string"),
					schematest.WithContains(schematest.New("string")),
					schematest.WithMinItems(3),
					schematest.WithMaxContains(1),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "failed to generate valid array: reached maximum of value maxContains 1 within 10 attempts")
			},
		},
		{
			name: "prefixItems",
			req: &Request{
				Schema: schematest.New("array",
					schematest.WithPrefixItems(
						schematest.New("number"),
						schematest.New("string"),
						&schema.Schema{Enum: []any{"Street", "Avenue", "Boulevard"}},
						&schema.Schema{Enum: []any{"NW", "NE", "SW", "SE"}},
					),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{1.0858618837067628e+308, "qwCrwMfkOjo", "Street", "SE"}, v)
			},
		},
		{
			name: "prefixItems",
			req: &Request{
				Schema: schematest.New("array",
					schematest.WithPrefixItems(
						schematest.New("number"),
						schematest.New("string"),
						&schema.Schema{Enum: []any{"Street", "Avenue", "Boulevard"}},
						&schema.Schema{Enum: []any{"NW", "NE", "SW", "SE"}},
					),
					schematest.WithItemsNew(schematest.NewBool(false)),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{1.0858618837067628e+308, "qwCrwMfkOjo", "Street", "SE"}, v)
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
