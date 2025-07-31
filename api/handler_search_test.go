package api

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_SearchQuery(t *testing.T) {
	toConfig := func(c *openapi.Config) *dynamic.Config {
		cfg := &dynamic.Config{
			Info: dynamictest.NewConfigInfo(),
			Data: c,
		}
		return cfg
	}

	testcases := []struct {
		name         string
		app          func() *runtime.App
		requestUrl   string
		responseBody string
		response     []try.ResponseCondition
	}{
		{
			name:       "empty search query",
			requestUrl: "/api/search/query",
			response: []try.ResponseCondition{
				try.HasStatusCode(200),
				try.HasHeader("Content-Type", "application/json"),
				try.HasBody(`{"results":[{"type":"HTTP","title":"foo","params":{"service":"foo","type":"http"}}],"total":1}`),
			},
			app: func() *runtime.App {
				app := runtime.New(&static.Config{Api: static.Api{Search: static.Search{
					Enabled: true,
				}}})

				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))
				app.AddHttp(toConfig(cfg))

				return app
			},
		},
		{
			name:       "search with query text",
			requestUrl: "/api/search/query?q=foo",
			response: []try.ResponseCondition{
				try.HasStatusCode(200),
				try.HasHeader("Content-Type", "application/json"),
				try.HasBody(`{"results":[{"type":"HTTP","title":"foo","fragments":["\u003cmark\u003efoo\u003c/mark\u003e"],"params":{"service":"foo","type":"http"}}],"total":1}`),
			},
			app: func() *runtime.App {
				app := runtime.New(&static.Config{Api: static.Api{Search: static.Search{
					Enabled: true,
				}}})

				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))
				app.AddHttp(toConfig(cfg))

				return app
			},
		},
		{
			name:       "search with param",
			requestUrl: "/api/search/query?q=api=foo",
			response: []try.ResponseCondition{
				try.HasStatusCode(200),
				try.HasHeader("Content-Type", "application/json"),
				try.HasBody(`{"results":[{"type":"HTTP","title":"foo","fragments":["\u003cmark\u003efoo\u003c/mark\u003e"],"params":{"service":"foo","type":"http"}}],"total":1}`),
			},
			app: func() *runtime.App {
				app := runtime.New(&static.Config{Api: static.Api{Search: static.Search{
					Enabled: true,
				}}})

				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))
				app.AddHttp(toConfig(cfg))
				cfg = openapitest.NewConfig("3.0", openapitest.WithInfo("bar", "", ""))
				app.AddHttp(toConfig(cfg))

				return app
			},
		},
		{
			name:       "limit is not a number",
			requestUrl: "/api/search/query?limit=foo",
			response: []try.ResponseCondition{
				try.HasStatusCode(400),
				try.HasHeader("Content-Type", "application/json"),
				try.HasBody(`{"message":"invalid query parameter 'limit': must be a number"}`),
			},
			app: func() *runtime.App {
				app := runtime.New(&static.Config{Api: static.Api{Search: static.Search{
					Enabled: true,
				}}})

				cfg := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))
				app.AddHttp(toConfig(cfg))
				cfg = openapitest.NewConfig("3.0", openapitest.WithInfo("bar", "", ""))
				app.AddHttp(toConfig(cfg))

				return app
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app(), static.Api{})

			try.Handler(t,
				http.MethodGet,
				tc.requestUrl,
				nil,
				"",
				h,
				tc.response...,
			)
		})
	}
}
