package generator

import (
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
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
				require.Equal(t, []any{"vMZsBhpyDmbo", "YvsVnIkdsa ", "PE5psgu", "hPH4"}, v)
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
			name: "unique items",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithItems("integer",
						schematest.WithMinimum(1),
						schematest.WithMaximum(10),
					),
					schematest.WithMinItems(3),
					schematest.WithUniqueItems(),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{int64(5), int64(2), int64(8)}, v)
			},
		},
		{
			name: "unique items but not possible",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithItems("integer",
						schematest.WithMinimum(1),
						schematest.WithMaximum(3),
					),
					schematest.WithMinItems(5),
					schematest.WithUniqueItems(),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.EqualError(t, err, "failed to generate valid array: reached attempt limit (10) caused by: cannot fill array with unique items")
			},
		},
		{
			name: "array with example",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithMinItems(1),
					schematest.WithItems("object",
						schematest.WithProperty("firstname", schematest.New("string")),
						schematest.WithProperty("lastname", schematest.New("string")),
						// unknown property for generator
						schematest.WithProperty("foo", schematest.New("string")),
						schematest.WithRequired("firstname", "lastname", "foo"),
					),
					schematest.WithExamples([]any{
						map[string]any{
							"firstname": "Emily",
							"lastname":  "Nelson",
							"foo":       "bar",
						},
					}),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []interface{}{
					map[string]interface{}{
						"firstname": "Emily",
						"lastname":  "Nelson",
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
								"firstname": "Emily",
								"lastname":  "Nelson",
								"foo":       "bar",
							},
							map[string]any{
								"firstname": "Mia",
								"lastname":  "Carter",
								"foo":       "bar",
							},
							map[string]any{
								"firstname": "James",
								"lastname":  "Brown",
								"foo":       "zzz",
							},
							map[string]any{
								"firstname": "John",
								"lastname":  "Lewis",
								"foo":       "yuh",
							},
							map[string]any{
								"firstname": "Benjamin",
								"lastname":  "Davis",
								"foo":       "yuh",
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
						"firstname": "Emily",
						"foo":       "bar",
						"lastname":  "Nelson",
					}, map[string]any{
						"firstname": "Mia",
						"foo":       "bar",
						"lastname":  "Carter",
					}, map[string]any{
						"firstname": "James",
						"foo":       "bar",
						"lastname":  "Brown",
					}, map[string]any{
						"firstname": "John",
						"foo":       "bar",
						"lastname":  "Lewis",
					}, map[string]any{
						"firstname": "Benjamin",
						"foo":       "bar",
						"lastname":  "Davis",
					},
				}, v)
			},
		},
		{
			name: "contains",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithMinItems(1),
					schematest.WithContains(schematest.New("string")),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{
					"vMZsBhpyDmbo",
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
					"vMZsBhpyDmbo", "PE5psgu", "YvsVnIkdsa ",
				}, v)
			},
		},
		{
			name: "contains with maxContains",
			req: &Request{
				Path: []string{"people"},
				Schema: schematest.New("array",
					schematest.WithMinItems(1),
					schematest.WithContains(schematest.New("string")),
					schematest.WithMaxContains(3),
				),
			},
			test: func(t *testing.T, v interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, []any{
					"",
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
				require.EqualError(t, err, "failed to generate valid array: reached attempt limit (10) caused by: reached maximum of value maxContains=1")
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
				a := v.([]any)
				require.InDelta(t, -170715.30581115812, a[0], 0.000001)
				require.Equal(t, "", a[1])
				require.Equal(t, "Boulevard", a[2])
				require.Equal(t, "NE", a[3])
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
				a := v.([]any)
				require.InDelta(t, -170715.30581115812, a[0], 0.000001)
				require.Equal(t, "", a[1])
				require.Equal(t, "Boulevard", a[2])
				require.Equal(t, "NE", a[3])
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
