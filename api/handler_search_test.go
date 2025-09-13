package api

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi/asyncapitest"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/runtime/search"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_SearchQuery(t *testing.T) {
	toConfig := func(c any) *dynamic.Config {
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
				try.HasBody(`{"results":[{"type":"HTTP","title":"foo","params":{"service":"foo","type":"http"}}],"facets":{"type":[{"value":"HTTP","count":1}]},"total":1}`),
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
				try.HasBody(`{"results":[{"type":"HTTP","title":"foo","fragments":["\u003cmark\u003efoo\u003c/mark\u003e"],"params":{"service":"foo","type":"http"}}],"facets":{"type":[{"value":"HTTP","count":1}]},"total":1}`),
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
				try.HasBody(`{"results":[{"type":"HTTP","title":"foo","fragments":["\u003cmark\u003efoo\u003c/mark\u003e"],"params":{"service":"foo","type":"http"}}],"facets":{"type":[{"value":"HTTP","count":1}]},"total":1}`),
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
		{
			name:       "using facet type=Kafka should filter out HTTP (case insensitive)",
			requestUrl: "/api/search/query?q=foo&type=Kafka",
			response: []try.ResponseCondition{
				try.HasStatusCode(200),
				try.HasHeader("Content-Type", "application/json"),
				try.AssertBody(func(t *testing.T, body string) {
					var result search.Result
					err := json.Unmarshal([]byte(body), &result)
					require.NoError(t, err)
					require.Len(t, result.Facets, 1)
					require.Equal(t, []search.FacetValue{{Value: "HTTP", Count: 1}, {Value: "Kafka", Count: 1}}, result.Facets["type"])
					require.Len(t, result.Results, 1)
					require.Equal(t, result.Results[0].Type, "Kafka")
				}),
			},
			app: func() *runtime.App {
				app := runtime.New(&static.Config{Api: static.Api{Search: static.Search{
					Enabled: true,
				}}})

				h := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))
				app.AddHttp(toConfig(h))
				k := asyncapitest.NewConfig(asyncapitest.WithInfo("foo", "", ""))
				_, err := app.Kafka.Add(toConfig(k), enginetest.NewEngine())
				require.NoError(t, err)

				return app
			},
		},
		{
			name:       "search with additional query parameter should not be used as facet",
			requestUrl: "/api/search/query?q=foo&refresh=20",
			response: []try.ResponseCondition{
				try.HasStatusCode(200),
				try.HasHeader("Content-Type", "application/json"),
				try.AssertBody(func(t *testing.T, body string) {
					var result search.Result
					err := json.Unmarshal([]byte(body), &result)
					require.NoError(t, err)
					require.Len(t, result.Facets, 1)
					require.Equal(t, []search.FacetValue{{Value: "HTTP", Count: 1}, {Value: "Kafka", Count: 1}}, result.Facets["type"])
					require.Len(t, result.Results, 2)
				}),
			},
			app: func() *runtime.App {
				app := runtime.New(&static.Config{Api: static.Api{Search: static.Search{
					Enabled: true,
				}}})

				h := openapitest.NewConfig("3.0", openapitest.WithInfo("foo", "", ""))
				app.AddHttp(toConfig(h))
				k := asyncapitest.NewConfig(asyncapitest.WithInfo("foo", "", ""))
				_, err := app.Kafka.Add(toConfig(k), enginetest.NewEngine())
				require.NoError(t, err)

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
