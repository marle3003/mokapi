package runtime_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/runtime/search"
	"testing"
)

func TestIndex_Http(t *testing.T) {
	toConfig := func(c *openapi.Config) *dynamic.Config {
		cfg := &dynamic.Config{
			Info: dynamictest.NewConfigInfo(),
			Data: c,
		}
		return cfg
	}

	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App)
	}{
		{
			name: "Search by name",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))
				app.AddHttp(toConfig(cfg))
				r, err := app.Search(search.Request{Query: "foo", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
						Domain:    "foo",
						Title:     "foo",
						Fragments: []string{"<mark>foo</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "Search by substring",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("My petstore API", "", ""))
				app.AddHttp(toConfig(cfg))
				r, err := app.Search(search.Request{Query: "pet", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
						Domain:    "My petstore API",
						Title:     "My petstore API",
						Fragments: []string{"My <mark>petstore</mark> API"},
						Params: map[string]string{
							"type":    "http",
							"service": "My petstore API",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "Search by version",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "1.0", ""))
				app.AddHttp(toConfig(cfg))
				r, err := app.Search(search.Request{Query: "1.0", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
						Domain:    "foo",
						Title:     "foo",
						Fragments: []string{"<mark>1.0</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "Search by path",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0",
					openapitest.WithInfo("foo", "1.0", ""),
					openapitest.WithPath("/pets", openapitest.NewPath()),
				)
				app.AddHttp(toConfig(cfg))
				r, err := app.Search(search.Request{Query: "pets", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
						Domain:    "foo",
						Title:     "/pets",
						Fragments: []string{"/<mark>pets</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
							"path":    "/pets",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "Search config name and path description have same text",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0",
					openapitest.WithInfo("foo", "1.0", "a description"),
					openapitest.WithPath("/pets", openapitest.NewPath(openapitest.WithPathInfo("", "a description"))),
				)
				app.AddHttp(toConfig(cfg))
				r, err := app.Search(search.Request{Query: "description", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 2)
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
						Domain:    "foo",
						Title:     "foo",
						Fragments: []string{"a <mark>description</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
						},
					},
					r.Results[0])
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
						Domain:    "foo",
						Title:     "/pets",
						Fragments: []string{"a <mark>description</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
							"path":    "/pets",
						},
					},
					r.Results[1])
			},
		},
		{
			name: "Search operation",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0",
					openapitest.WithInfo("foo", "1.0", "a description"),
					openapitest.WithPath("/pets", openapitest.NewPath(
						openapitest.WithPathInfo("", "a description"),
						openapitest.WithOperation("get", openapitest.NewOperation(
							openapitest.WithHeaderParam("foo", true, openapitest.WithParamInfo("parameter description")),
						)),
					)),
				)
				app.AddHttp(toConfig(cfg))
				r, err := app.Search(search.Request{Query: "\"parameter description\"", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
						Domain:    "foo",
						Title:     "GET /pets",
						Fragments: []string{"<mark>parameter</mark> <mark>description</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
							"path":    "/pets",
							"method":  "get",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "Search by api",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))
				app.AddHttp(toConfig(cfg))
				r, err := app.Search(search.Request{Terms: map[string]string{"api": "foo"}, Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
						Domain:    "foo",
						Title:     "foo",
						Fragments: []string{"<mark>foo</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "Search by api with space",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo bar", "", ""))
				app.AddHttp(toConfig(cfg))
				r, err := app.Search(search.Request{Terms: map[string]string{"api": "foo bar"}, Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
						Domain:    "foo bar",
						Title:     "foo bar",
						Fragments: []string{"<mark>foo</mark> <mark>bar</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo bar",
						},
					},
					r.Results[0])
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := runtime.New(
				&static.Config{
					Api: static.Api{
						Search: static.Search{
							Enabled:  true,
							Analyzer: "ngram",
							Ngram: static.NgramAnalyzer{
								Min: 3,
								Max: 5,
							},
						},
					},
				})
			tc.test(t, app)
		})
	}
}
