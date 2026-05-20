package runtime_test

import (
	"context"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/runtime/search"
	"mokapi/safe"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", "An API description"))
				app.Http.Add(toConfig(cfg))

				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "foo", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:        "HTTP",
						Title:       "foo",
						Description: "An API description",
						Fragments:   []string{"<mark>foo</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "summary takes precedence over description.",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0",
					openapitest.WithInfo("foo", "", "An API description"),
					openapitest.WithSummary("A short summary"),
				)
				app.Http.Add(toConfig(cfg))

				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "foo", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:        "HTTP",
						Title:       "foo",
						Description: "A short summary An API description",
						Fragments:   []string{"<mark>foo</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "truncate summary",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0",
					openapitest.WithInfo("foo", "", "An API description"),
					openapitest.WithSummary("Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum."),
				)
				app.Http.Add(toConfig(cfg))

				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "foo", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:        "HTTP",
						Title:       "foo",
						Description: "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam...",
						Fragments:   []string{"<mark>foo</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "config should be remove from index",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))
				app.Http.Add(toConfig(cfg))
				waitSearchIndex(t, func() bool {
					r, err := app.Search(search.Request{QueryText: "foo", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				r, err := app.Search(search.Request{QueryText: "foo", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				app.Http.Remove(toConfig(cfg))
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "foo", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 0
				})
			},
		},
		{
			name: "Search by substring",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("My petstore API", "", ""))
				app.Http.Add(toConfig(cfg))

				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "pet*", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
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
			name: "mailpiece should not match mailbox",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("mailbox", "", ""))
				app.Http.Add(toConfig(cfg))

				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "mailpiece", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 0
				})
				require.Len(t, r.Results, 0)
			},
		},
		{
			name: "Search by version",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "1.0", ""))
				app.Http.Add(toConfig(cfg))

				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "1.0", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
						Title:     "foo",
						Fragments: []string{"<mark>1</mark><mark>.</mark><mark>0</mark>"},
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
					openapitest.WithPath("/pets", openapitest.WithOperation(http.MethodGet)),
				)
				app.Http.Add(toConfig(cfg))

				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "pets", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 2
				})
				require.Len(t, r.Results, 2)
				require.Contains(t,
					r.Results,
					search.ResultItem{
						Type:      "HTTP",
						Domain:    "foo",
						Title:     "/pets",
						Fragments: []string{"/<mark>pets</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
							"path":    "/pets",
							"methods": "GET",
						},
					},
				)
				require.Contains(t,
					r.Results,
					search.ResultItem{
						Type:      "HTTP",
						Domain:    "foo",
						Title:     "/pets",
						Fragments: []string{"/<mark>pets</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
							"path":    "/pets",
							"method":  "GET",
						},
					},
				)
			},
		},
		{
			name: "Search config name and path description have same text",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0",
					openapitest.WithInfo("foo", "1.0", "a description"),
					openapitest.WithPath("/pets", openapitest.WithPathInfo("", "a description")),
				)
				app.Http.Add(toConfig(cfg))

				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "description", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 2
				})
				require.Len(t, r.Results, 2)
				require.Equal(t,
					search.ResultItem{
						Type:        "HTTP",
						Title:       "foo",
						Description: "a description",
						Fragments:   []string{"a <mark>description</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
						},
					},
					r.Results[1])
				require.Equal(t,
					search.ResultItem{
						Type:        "HTTP",
						Domain:      "foo",
						Title:       "/pets",
						Description: `a description`,
						Fragments:   []string{"a <mark>description</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
							"path":    "/pets",
							"methods": "",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "Search operation",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0",
					openapitest.WithInfo("foo", "1.0", "a description"),
					openapitest.WithPath("/pets",
						openapitest.WithPathInfo("", "a description"),
						openapitest.WithOperation(http.MethodGet,
							openapitest.WithOperationInfo("Summary value", "Description value", "", false),
							openapitest.WithHeaderParam("foo", true, openapitest.WithParamInfo("parameter description")),
						),
					),
				)
				app.Http.Add(toConfig(cfg))

				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "\"parameter description\"", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:        "HTTP",
						Domain:      "foo",
						Title:       "/pets",
						Description: `Summary value Description value`,
						Fragments:   []string{"<mark>parameter</mark> <mark>description</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "foo",
							"path":    "/pets",
							"method":  "GET",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "Search by api field",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("Petstore", "", ""),
					openapitest.WithPath("/pets",
						openapitest.WithPathInfo("", "path"),
						openapitest.WithOperation("get",
							openapitest.WithHeaderParam("foo", true, openapitest.WithParamInfo("parameter")),
						),
					),
				)
				app.Http.Add(toConfig(cfg))
				// search response should only have one the root OpenAPI object
				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "Petstore", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)

				// search by api should return all items in the OpenAPI
				r, err = app.Search(search.Request{QueryText: "api:Petstore", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 3)
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
						Title:     "Petstore",
						Fragments: []string{"<mark>Petstore</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "Petstore",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "Search by api with space",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo bar", "", ""))
				app.Http.Add(toConfig(cfg))

				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "api:\"foo bar\"", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
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
		{
			name: "Search by api field",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("Petstore", "", ""),
					openapitest.WithPath("/pets",
						openapitest.WithPathInfo("", "path"),
						openapitest.WithOperation("get",
							openapitest.WithHeaderParam("foo", true, openapitest.WithParamInfo("parameter")),
						),
					),
				)
				app.Http.Add(toConfig(cfg))
				// search response should only have one the root OpenAPI object
				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "Petstore", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)

				// search by api should return all items in the OpenAPI
				r, err = app.Search(search.Request{QueryText: "api:Petstore", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 3)
				require.Equal(t,
					search.ResultItem{
						Type:      "HTTP",
						Title:     "Petstore",
						Fragments: []string{"<mark>Petstore</mark>"},
						Params: map[string]string{
							"type":    "http",
							"service": "Petstore",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "Search fuzzy",
			test: func(t *testing.T, app *runtime.App) {
				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("Petstore", "", ""),
					openapitest.WithPath("/pets",
						openapitest.WithPathInfo("", "path"),
						openapitest.WithOperation("get",
							openapitest.WithHeaderParam("foo", true, openapitest.WithParamInfo("parameter")),
						),
					),
				)
				app.Http.Add(toConfig(cfg))
				// search response should only have one the root OpenAPI object
				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "pest~", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 2
				})
				require.Len(t, r.Results, 2)
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
							InMemory: true,
						},
					},
				}, &dynamictest.Reader{})

			pool := safe.NewPool(context.Background())
			app.Start(pool)
			defer pool.Stop()

			tc.test(t, app)
		})
	}
}

func TestIndex_Http_Event(t *testing.T) {
	api := openapitest.NewConfig("3.0",
		openapitest.WithInfo("Test HTTP Events", "", ""),
		openapitest.WithPath("/foo",
			openapitest.WithOperation(http.MethodGet,
				openapitest.WithResponse(http.StatusOK),
			),
		),
	)
	cfg := &dynamic.Config{
		Info: dynamictest.NewConfigInfo(),
		Data: api,
	}

	testcases := []struct {
		name string
		test func(t *testing.T, h openapi.Handler, app *runtime.App)
	}{
		{
			name: "search event by method",
			test: func(t *testing.T, h openapi.Handler, app *runtime.App) {
				req := httptest.NewRequest("GET", "http://localhost/foo", nil)
				w := httptest.NewRecorder()
				he := h.ServeHTTP(w, req)
				require.Nil(t, he)

				r, err := waitSearchResult(t, func() (search.Result, error) {
					return app.Search(search.Request{QueryText: "+method:GET +type:event", Limit: 10})
				}, 1)

				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t, "Event", r.Results[0].Type)
				require.Equal(t, "Test HTTP Events", r.Results[0].Domain)
				require.Equal(t, "http://localhost/foo", r.Results[0].Title)
				require.Len(t, r.Results[0].Fragments, 2)
				require.Contains(t, r.Results[0].Fragments, "<mark>GET</mark>")
				require.Contains(t, r.Results[0].Fragments, "<mark>event</mark>")
				require.Len(t, r.Results[0].Params, 6)
				require.Equal(t, "event", r.Results[0].Params["type"])
				require.Equal(t, "http", r.Results[0].Params["traits.namespace"])
				require.Equal(t, "Test HTTP Events", r.Results[0].Params["traits.name"])
				require.Equal(t, "/foo", r.Results[0].Params["traits.path"])
				require.Equal(t, "GET", r.Results[0].Params["traits.method"])
				require.Contains(t, r.Results[0].Params, "id")
				require.NotEmpty(t, r.Results[0].Time)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			app := runtime.New(
				&static.Config{
					Api: static.Api{
						Search: static.Search{
							Enabled:  true,
							InMemory: true,
						},
					},
				}, &dynamictest.Reader{})

			app.Http.Add(cfg)
			pool := safe.NewPool(context.Background())
			app.Start(pool)
			defer pool.Stop()

			h := openapi.NewHandler(api, enginetest.NewEngine(), app.Events)
			tc.test(t, h, app)
		})
	}
}

func waitSearchIndex(t *testing.T, check func() bool) {
	deadline := time.Now().Add(2 * time.Second)

	for {
		if check() {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("wait search index reached deadline")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func waitSearchResult(t *testing.T, f func() (search.Result, error), expectedResults int) (search.Result, error) {
	deadline := time.Now().Add(2 * time.Second)

	for {
		r, err := f()
		if err != nil {
			return r, err
		}
		if len(r.Results) == expectedResults {
			return r, nil
		}
		if time.Now().After(deadline) {
			t.Fatalf("wait search result reached deadline: last search returned %d results, %d was expected", len(r.Results), expectedResults)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
