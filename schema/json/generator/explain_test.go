package generator_test

import (
	"errors"
	"mokapi/schema/json/generator"
	"mokapi/schema/json/schema"
	"mokapi/schema/json/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func propertyPaths(matches []generator.Match) []string {
	var paths []string
	for _, m := range matches {
		paths = append(paths, m.PropertyPath)
	}
	return paths
}

func addCustomNode(t *testing.T, n *generator.Node) {
	t.Helper()
	root := generator.FindByName(generator.RootName)
	root.Append(n)
	t.Cleanup(func() {
		require.NoError(t, root.Remove(n.Name))
	})
}

func TestExplain(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "path only",
			test: func(t *testing.T) {
				result, err := generator.Explain(generator.NewRequest([]string{"customerEmail"}, nil, nil))

				require.NoError(t, err)
				require.Len(t, result.Matches, 1)
				m := result.Matches[0]
				require.Equal(t, "customerEmail", m.PropertyPath)
				require.Equal(t, "/email", m.Node)
				require.Equal(t, []string{"email"}, m.Tokens)
				require.Equal(t, generator.SourceNode, m.Source)
				require.Equal(t, result.Value, m.Value)
			},
		},
		{
			name: "path is the start of the property path",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithProperty("email", schematest.New("string")),
					schematest.WithRequired("email"),
				)

				result, err := generator.Explain(generator.NewRequest([]string{"user"}, s, nil))

				require.NoError(t, err)
				require.Equal(t, []string{"user.email"}, propertyPaths(result.Matches))
			},
		},
		{
			name: "schema only without path",
			test: func(t *testing.T) {
				s := schematest.New("integer", schematest.WithMinimum(1), schematest.WithMaximum(3))

				result, err := generator.Explain(generator.NewRequest(nil, s, nil))

				require.NoError(t, err)
				require.Len(t, result.Matches, 1)
				m := result.Matches[0]
				require.Equal(t, "(value)", m.PropertyPath)
				require.Equal(t, "fakeBySchema", m.Node)
				require.Empty(t, m.Tokens)
				require.Equal(t, generator.SourceNode, m.Source)
				require.Equal(t, result.Value, m.Value)
			},
		},
		{
			name: "nested object has one row per leaf with the full property path",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithProperty("email", schematest.New("string")),
					schematest.WithProperty("address", schematest.New("object",
						schematest.WithProperty("city", schematest.New("string")),
						schematest.WithRequired("city"),
					)),
					schematest.WithRequired("email", "address"),
				)

				result, err := generator.Explain(generator.NewRequest(nil, s, nil))

				require.NoError(t, err)
				require.Equal(t, []string{"email", "address.city"}, propertyPaths(result.Matches))
				require.Equal(t, "/contact/email", result.Matches[0].Node, "email and address select the domain contact")
				obj := result.Value.(map[string]any)
				require.Equal(t, obj["email"], result.Matches[0].Value)
				require.Equal(t, obj["address"].(map[string]any)["city"], result.Matches[1].Value)
			},
		},
		{
			name: "array has one row for its items",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithProperty("xqzs", schematest.New("array",
						schematest.WithItems("string"),
						schematest.WithMinItems(2),
					)),
					schematest.WithRequired("xqzs"),
				)

				result, err := generator.Explain(generator.NewRequest(nil, s, nil))

				require.NoError(t, err)
				require.Equal(t, []string{"xqzs[]"}, propertyPaths(result.Matches))
				items := result.Value.(map[string]any)["xqzs"].([]any)
				require.Equal(t, items[len(items)-1], result.Matches[0].Value, "value of the last generated item")
			},
		},
		{
			name: "array of objects",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithProperty("xqzs", schematest.New("array",
						schematest.WithItemsNew(schematest.New("object",
							schematest.WithProperty("name", schematest.New("string")),
							schematest.WithRequired("name"),
						)),
						schematest.WithMinItems(1),
					)),
					schematest.WithRequired("xqzs"),
				)

				result, err := generator.Explain(generator.NewRequest(nil, s, nil))

				require.NoError(t, err)
				require.Equal(t, []string{"xqzs[].name"}, propertyPaths(result.Matches))
			},
		},
		{
			name: "array without items generated has no row",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithProperty("xqzs", schematest.New("array",
						schematest.WithItems("string"),
						schematest.WithMaxItems(0),
					)),
					schematest.WithRequired("xqzs"),
				)

				result, err := generator.Explain(generator.NewRequest(nil, s, nil))

				require.NoError(t, err)
				require.Empty(t, result.Matches)
			},
		},
		{
			name: "additional properties have one row because their names are not known upfront",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithAdditionalProperties(schematest.New("integer")),
					schematest.WithMinProperties(2),
				)

				result, err := generator.Explain(generator.NewRequest(nil, s, nil))

				require.NoError(t, err)
				require.NotEmpty(t, result.Value.(map[string]any))
				require.Equal(t, []string{"*"}, propertyPaths(result.Matches))
			},
		},
		{
			name: "custom node added to the tree",
			test: func(t *testing.T) {
				addCustomNode(t, &generator.Node{
					Name:       "frequency",
					Attributes: []string{"frequency"},
					Custom:     true,
					Fake:       func(r *generator.Request) (any, error) { return "weekly", nil },
				})

				result, err := generator.Explain(generator.NewRequest([]string{"billingFrequency"}, nil, nil))

				require.NoError(t, err)
				require.Equal(t, "weekly", result.Value)
				require.Equal(t, generator.Match{
					PropertyPath: "billingFrequency",
					Node:         "/frequency",
					Tokens:       []string{"frequency"},
					Source:       generator.SourceNode,
					Value:        "weekly",
				}, result.Matches[0])
			},
		},
		{
			name: "fallback when node is not supported",
			test: func(t *testing.T) {
				addCustomNode(t, &generator.Node{
					Name:       "pass",
					Attributes: []string{"pass"},
					Fake:       func(r *generator.Request) (any, error) { return nil, generator.NotSupported },
				})

				result, err := generator.Explain(generator.NewRequest([]string{"pass"}, schematest.New("string"), nil))

				require.NoError(t, err)
				require.Len(t, result.Matches, 1)
				require.Equal(t, "fakeBySchema", result.Matches[0].Node)
				require.Equal(t, result.Value, result.Matches[0].Value)
			},
		},
		{
			name: "fallback when value of node is not valid for the schema",
			test: func(t *testing.T) {
				addCustomNode(t, &generator.Node{
					Name:       "broken",
					Attributes: []string{"broken"},
					Fake:       func(r *generator.Request) (any, error) { return map[string]any{"x": 1}, nil },
				})

				result, err := generator.Explain(generator.NewRequest([]string{"broken"}, schematest.New("integer"), nil))

				require.NoError(t, err)
				require.Len(t, result.Matches, 1)
				require.Equal(t, "fakeBySchema", result.Matches[0].Node)
				require.IsType(t, int64(0), result.Value)
				require.Equal(t, result.Value, result.Matches[0].Value)
			},
		},
		{
			name: "source enum",
			test: func(t *testing.T) {
				s := schematest.New("string", schematest.WithEnumValues("open", "closed"))

				result, err := generator.Explain(generator.NewRequest([]string{"status"}, s, nil))

				require.NoError(t, err)
				require.Len(t, result.Matches, 1)
				require.Equal(t, generator.NoNode, result.Matches[0].Node)
				require.Equal(t, generator.SourceEnum, result.Matches[0].Source)
				require.Equal(t, result.Value, result.Matches[0].Value)
			},
		},
		{
			name: "source enum of unique array items",
			test: func(t *testing.T) {
				unique := true
				s := schematest.New("object",
					schematest.WithProperty("xqzs", schematest.New("array",
						schematest.WithItemsNew(schematest.New("string", schematest.WithEnumValues("a", "b", "c"))),
						schematest.WithMinItems(1),
						func(s *schema.Schema) { s.UniqueItems = &unique },
					)),
					schematest.WithRequired("xqzs"),
				)

				result, err := generator.Explain(generator.NewRequest(nil, s, nil))

				require.NoError(t, err)
				require.Len(t, result.Matches, 1)
				require.Equal(t, "xqzs[]", result.Matches[0].PropertyPath)
				require.Equal(t, generator.NoNode, result.Matches[0].Node)
				require.Equal(t, generator.SourceEnum, result.Matches[0].Source)
			},
		},
		{
			name: "source const",
			test: func(t *testing.T) {
				s := schematest.New("string", schematest.WithConst("fixed"))

				result, err := generator.Explain(generator.NewRequest([]string{"kind"}, s, nil))

				require.NoError(t, err)
				require.Equal(t, "fixed", result.Value)
				require.Equal(t, generator.NoNode, result.Matches[0].Node)
				require.Equal(t, generator.SourceConst, result.Matches[0].Source)
			},
		},
		{
			name: "source examples",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithProperty("xqz", schematest.New("string")),
					schematest.WithRequired("xqz"),
					schematest.WithExamples(map[string]any{"xqz": "sample"}),
				)

				result, err := generator.Explain(generator.NewRequest(nil, s, nil))

				require.NoError(t, err)
				require.Equal(t, map[string]any{"xqz": "sample"}, result.Value)
				require.Len(t, result.Matches, 1)
				require.Equal(t, generator.NoNode, result.Matches[0].Node)
				require.Equal(t, generator.SourceExamples, result.Matches[0].Source)
			},
		},
		{
			name: "source examples of an object schema without properties",
			test: func(t *testing.T) {
				s := schematest.New("object", schematest.WithExamples(map[string]any{"id": 1}))

				result, err := generator.Explain(generator.NewRequest(nil, s, nil))

				require.NoError(t, err)
				require.Equal(t, map[string]any{"id": 1}, result.Value)
				require.Len(t, result.Matches, 1)
				require.Equal(t, generator.SourceExamples, result.Matches[0].Source)
			},
		},
		{
			name: "source context",
			test: func(t *testing.T) {
				req := generator.NewRequest([]string{"status"}, schematest.New("string"), map[string]any{"status": "reused"})

				result, err := generator.Explain(req)

				require.NoError(t, err)
				require.Equal(t, "reused", result.Value)
				require.Equal(t, generator.NoNode, result.Matches[0].Node)
				require.Equal(t, generator.SourceContext, result.Matches[0].Source)
			},
		},
		{
			name: "source null",
			test: func(t *testing.T) {
				s := schematest.NewTypes([]string{"string", "null"})

				// null is chosen with a probability of 5 %
				for range 3000 {
					result, err := generator.Explain(generator.NewRequest([]string{"xqz"}, s, nil))
					require.NoError(t, err)
					if result.Value == nil && result.Matches[0].Source == generator.SourceNull {
						require.Equal(t, generator.NoNode, result.Matches[0].Node)
						return
					}
				}
				require.Fail(t, "null was never generated")
			},
		},
		{
			name: "optional property that is not emitted has no row",
			test: func(t *testing.T) {
				s := schematest.New("object",
					schematest.WithProperty("a", schematest.New("string")),
					schematest.WithProperty("b", schematest.New("string")),
					schematest.WithRequired("a"),
				)

				omitted := false
				for range 100 {
					result, err := generator.Explain(generator.NewRequest(nil, s, nil))
					require.NoError(t, err)
					obj := result.Value.(map[string]any)
					_, hasB := obj["b"]
					if hasB {
						require.Equal(t, []string{"a", "b"}, propertyPaths(result.Matches))
					} else {
						omitted = true
						require.Equal(t, []string{"a"}, propertyPaths(result.Matches))
					}
				}
				require.True(t, omitted, "optional property was always emitted")
			},
		},
		{
			name: "recursion guard error",
			test: func(t *testing.T) {
				s := schematest.New("object")
				s.Properties = &schema.Schemas{}
				s.Properties.Set("loop", s)
				s.Required = []string{"loop"}

				_, err := generator.Explain(generator.NewRequest(nil, s, nil))

				var guard *generator.RecursionGuard
				require.True(t, errors.As(err, &guard), "got: %v", err)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func TestNew_DoesNotTrace(t *testing.T) {
	s := schematest.New("object",
		schematest.WithProperty("email", schematest.New("string")),
		schematest.WithRequired("email"),
	)

	v, err := generator.New(generator.NewRequest(nil, s, nil))

	require.NoError(t, err)
	require.Contains(t, v, "email")
}
