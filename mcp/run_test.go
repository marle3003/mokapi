package mcp_test

import (
	"context"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/mcp"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/runtimetest"
	"mokapi/runtime/search"
	"mokapi/safe"
	"mokapi/schema/json/generator"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestService_Run(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service)
	}{
		{
			name: "run JavaScript code",
			app:  runtimetest.NewApp(),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `1+1`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, int64(2), r.Result)
			},
		},
		{
			name: "JSON.parse()",
			app:  runtimetest.NewApp(),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `JSON.parse('{"foo":"bar"}')`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, map[string]any{"foo": "bar"}, r.Result)
			},
		},
		{
			name: "List APIs skip empty name",
			app: runtimetest.NewHttpApp(
				openapitest.NewConfig("3.1.0"),
			),
			test: func(t *testing.T, s *mcp.Service) {
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.getApis()`,
					},
				)
				require.NoError(t, err)
				require.Len(t, r.Result, 0)
			},
		},
		{
			name: "script error",
			app:  runtimetest.NewApp(),
			test: func(t *testing.T, s *mcp.Service) {
				_, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `okapi.getApis()`,
					},
				)
				require.EqualError(t, err, `ReferenceError: okapi is not defined at mokapi_execute_code.js:1:1(0)

Tip for Correction:
It seems there is a syntax error or a misunderstanding of the API. 
To ensure you are using the correct global variables and methods:
1. Call 'mokapi_get_automation_definitions' without parameters to see the general overview.
2. Check 'category="core"' to verify the syntax of the global 'mokapi' object.`)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(123456)

			s := mcp.NewService(tc.app)
			tc.test(t, s)
		})
	}
}

func TestService_Run_Search(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		test func(t *testing.T, s *mcp.Service, app *runtime.App)
	}{
		{
			name: "search",
			app: func() *runtime.App {
				app := runtime.New(
					&static.Config{
						Api: static.Api{
							Search: static.Search{
								Enabled:  true,
								InMemory: true,
							},
						},
					}, &dynamictest.Reader{})
				app.Http.Add(&dynamic.Config{
					Info: dynamictest.NewConfigInfo(),
					Data: openapitest.NewConfig("3.1.0",
						openapitest.WithInfo("foo", "", ""),
						openapitest.WithPath("/foo/payment/bar",
							openapitest.WithOperation(http.MethodPost,
								openapitest.WithOperationSummary("Payment summary"),
							),
						),
					),
				})
				return app
			}(),
			test: func(t *testing.T, s *mcp.Service, app *runtime.App) {
				pool := safe.NewPool(context.Background())
				app.Start(pool)
				defer pool.Stop()
				waitSearchIndex(t, func() bool {
					r, err := app.Search(search.Request{QueryText: "foo", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) > 0
				})
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `mokapi.search("method:POST AND path:*payments*")`,
					},
				)
				require.NoError(t, err)
				require.Equal(t, mcp.SearchResult{Items: []mcp.SearchResultItem{
					{
						Type:      "HTTP",
						Domain:    "foo",
						Title:     "/foo/payment/bar",
						Fragments: []string{"<mark>POST</mark>"},
						Metadata:  map[string]string{"method": "POST", "path": "/foo/payment/bar", "service": "foo", "type": "http"},
						Time:      "",
					}}, Total: 1}, r.Result)

				r, err = s.GetRunResponse(
					context.Background(),
					mcp.RunInput{Code: `const searchResponse = mokapi.search("method:POST AND path:*payments*");
let op = null
if (searchResponse.items.length > 0) {
    const match = searchResponse.items[0];
    const api = mokapi.getApi(match.metadata.service);
	const opSummary = api.getOperations().find(x => x.path === match.metadata.path && x.method === match.metadata.method);
    if (opSummary) {
		op = api.getOperation(opSummary.id);
    }
}
op`},
				)
				require.NoError(t, err)
				require.IsType(t, &mcp.Operation{}, r.Result)
				op := r.Result.(*mcp.Operation)
				require.Equal(t, "/foo/payment/bar", op.Path)
			},
		},
		{
			name: "search for event",
			app: func() *runtime.App {
				app := runtime.New(
					&static.Config{
						Api: static.Api{
							Search: static.Search{
								Enabled:  true,
								InMemory: true,
							},
						},
					}, &dynamictest.Reader{})
				app.Http.Add(&dynamic.Config{
					Info: dynamictest.NewConfigInfo(),
					Data: openapitest.NewConfig("3.1.0",
						openapitest.WithInfo("Petstore", "", ""),
						openapitest.WithPath("/pets",
							openapitest.WithOperation(http.MethodPost,
								openapitest.WithOperationSummary("Payment summary"),
							),
						),
					),
				})
				err := app.Events.Push(&openapi.HttpLog{
					Request: &openapi.HttpRequestLog{
						Method: http.MethodPost,
						Url:    "/pets",
					},
					Response: &openapi.HttpResponseLog{
						StatusCode: http.StatusNotFound,
					},
				}, events.NewTraits().WithNamespace("http").WithName("Petstore"))
				require.NoError(t, err)
				err = app.Events.Push(&openapi.HttpLog{
					Request: &openapi.HttpRequestLog{
						Method: http.MethodDelete,
						Url:    "/pets",
					},
					Response: &openapi.HttpResponseLog{
						StatusCode: http.StatusInternalServerError,
					},
				}, events.NewTraits().WithNamespace("http").WithName("Petstore"))
				require.NoError(t, err)
				err = app.Events.Push(&openapi.HttpLog{
					Request: &openapi.HttpRequestLog{
						Method: http.MethodGet,
						Url:    "/pets",
					},
					Response: &openapi.HttpResponseLog{
						StatusCode: http.StatusOK,
					},
				}, events.NewTraits().WithNamespace("http").WithName("Petstore"))
				require.NoError(t, err)
				return app
			}(),
			test: func(t *testing.T, s *mcp.Service, app *runtime.App) {
				pool := safe.NewPool(context.Background())
				app.Start(pool)
				defer pool.Stop()
				waitSearchIndex(t, func() bool {
					r, err := app.Search(search.Request{QueryText: "Petstore", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) > 0
				})
				r, err := s.GetRunResponse(
					context.Background(),
					mcp.RunInput{
						Code: `const errors = mokapi.search('+api:"Petstore" +response.statusCode:>=400 +type:event')
errors.items.map(x => mokapi.getEvent(x.metadata.id))`,
					},
				)
				require.NoError(t, err)
				require.IsType(t, []any{}, r.Result)
				evts := r.Result.([]any)
				require.Len(t, evts, 2)
				evt := evts[0]
				require.IsType(t, events.Event{}, evt)
				d1 := evts[0].(events.Event).Data.(*openapi.HttpLog)
				d2 := evts[1].(events.Event).Data.(*openapi.HttpLog)
				var evt404 *openapi.HttpLog
				var evt500 *openapi.HttpLog
				if d1.Response.StatusCode == http.StatusNotFound {
					evt404 = d1
					evt500 = d2
				} else if d2.Response.StatusCode == http.StatusNotFound {
					evt404 = d2
					evt500 = d1
				}

				require.NotNil(t, evt404)
				require.NotNil(t, evt500)

				require.Equal(t, http.MethodPost, evt404.Request.Method)
				require.Equal(t, http.MethodDelete, evt500.Request.Method)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			generator.Seed(123456)

			s := mcp.NewService(tc.app)
			tc.test(t, s, tc.app)
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
