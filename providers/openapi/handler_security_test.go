package openapi_test

import (
	"fmt"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime/events"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandler_Security(t *testing.T) {
	testcases := []struct {
		name  string
		test  func(t *testing.T, h http.HandlerFunc, c *openapi.Config, eh events.Handler)
		event func(request *common.HttpEventRequest, response *common.HttpEventResponse) []*common.Action
	}{
		{
			name: "basic",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config, eh events.Handler) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json")),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.HttpSecurityScheme{
						Scheme: "basic",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.UseOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("Authorization", "Basic 123")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, `"Basic 123"`, rr.Body.String())

				logs := eh.GetEvents(events.NewTraits().WithNamespace("http"))
				httpLog := logs[0].Data.(*openapi.HttpLog)
				require.Equal(t, "Basic 123", httpLog.Request.Parameters[0].Value)
				require.Equal(t, "Basic 123", *httpLog.Request.Parameters[0].Raw)
			},
			event: func(request *common.HttpEventRequest, response *common.HttpEventResponse) []*common.Action {
				response.Data = request.Header["Authorization"]
				return nil
			},
		},
		{
			name: "bearer but without authorization header",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config, eh events.Handler) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json")),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.HttpSecurityScheme{
						Scheme: "bearer",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.UseOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)

				logs := eh.GetEvents(events.NewTraits().WithNamespace("http"))
				httpLog := logs[0].Data.(*openapi.HttpLog)
				require.Equal(t, "", httpLog.Request.Parameters[0].Value)
				require.Nil(t, httpLog.Request.Parameters[0].Raw)
			},
		},
		{
			name: "bearer with authorization header",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config, eh events.Handler) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json")),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.HttpSecurityScheme{
						Scheme: "bearer",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.UseOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("Authorization", "Bearer 123")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"Bearer 123"`, rr.Body.String())

				logs := eh.GetEvents(events.NewTraits().WithNamespace("http"))
				httpLog := logs[0].Data.(*openapi.HttpLog)
				require.Equal(t, "Bearer 123", httpLog.Request.Parameters[0].Value)
				require.Equal(t, "Bearer 123", *httpLog.Request.Parameters[0].Raw)
			},
			event: func(request *common.HttpEventRequest, response *common.HttpEventResponse) []*common.Action {
				response.Data = request.Header["Authorization"]
				return nil
			},
		},
		{
			name: "ApiKey in header",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config, eh events.Handler) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json")),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.ApiKeySecurityScheme{
						In:   "header",
						Name: "X-API-KEY",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.UseOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("X-API-KEY", "123")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"123"`, rr.Body.String())

				logs := eh.GetEvents(events.NewTraits().WithNamespace("http"))
				httpLog := logs[0].Data.(*openapi.HttpLog)
				require.Equal(t, "123", httpLog.Request.Parameters[0].Value)
				require.Equal(t, "123", *httpLog.Request.Parameters[0].Raw)
			},
			event: func(request *common.HttpEventRequest, response *common.HttpEventResponse) []*common.Action {
				response.Data = request.Header["X-API-KEY"]
				return nil
			},
		},
		{
			name: "ApiKey in query",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config, eh events.Handler) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json")),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.ApiKeySecurityScheme{
						In:   "query",
						Name: "apikey",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.UseOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo?apikey=123", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"123"`, rr.Body.String())

				logs := eh.GetEvents(events.NewTraits().WithNamespace("http"))
				httpLog := logs[0].Data.(*openapi.HttpLog)
				require.Equal(t, "apikey", httpLog.Request.Parameters[0].Name)
				require.Equal(t, "123", httpLog.Request.Parameters[0].Value)
				require.Equal(t, "123", *httpLog.Request.Parameters[0].Raw)
			},
			event: func(request *common.HttpEventRequest, response *common.HttpEventResponse) []*common.Action {
				response.Data = request.Query["apikey"]
				return nil
			},
		},
		{
			name: "ApiKey in cookie",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config, eh events.Handler) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json")),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.ApiKeySecurityScheme{
						In:   "cookie",
						Name: "apikey",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.UseOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.AddCookie(&http.Cookie{Name: "apikey", Value: "123"})
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"123"`, rr.Body.String())

				logs := eh.GetEvents(events.NewTraits().WithNamespace("http"))
				httpLog := logs[0].Data.(*openapi.HttpLog)
				require.Equal(t, "123", httpLog.Request.Parameters[1].Value)
				require.Equal(t, "123", *httpLog.Request.Parameters[1].Raw)
			},
			event: func(request *common.HttpEventRequest, response *common.HttpEventResponse) []*common.Action {
				response.Data = request.Cookie["apikey"]
				return nil
			},
		},
		{
			name: "security scheme not supported",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config, eh events.Handler) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json")),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.NotSupportedSecurityScheme{
						Type: "NotSupportedSecurityScheme",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.UseOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
			},
		},
		{
			name: "oauth2",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config, eh events.Handler) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json")),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.OAuth2SecurityScheme{},
				}
				openapitest.AppendPath("/foo", c, openapitest.UseOperation("GET", op))

				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("Authorization", "Bearer 123")
				rr := httptest.NewRecorder()
				h(rr, r)

				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"Bearer 123"`, rr.Body.String())

				logs := eh.GetEvents(events.NewTraits().WithNamespace("http"))
				httpLog := logs[0].Data.(*openapi.HttpLog)
				require.Equal(t, "Bearer 123", httpLog.Request.Parameters[0].Value)
				require.Equal(t, "Bearer 123", *httpLog.Request.Parameters[0].Raw)
			},
			event: func(request *common.HttpEventRequest, response *common.HttpEventResponse) []*common.Action {
				response.Data = request.Header["Authorization"]
				return nil
			},
		},
		{
			name: "oauth2 and api key required",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config, eh events.Handler) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(
						map[string][]string{
							"foo": {},
							"bar": {},
						},
					),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json")),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.OAuth2SecurityScheme{},
					"bar": &openapi.ApiKeySecurityScheme{
						In:   "header",
						Name: "apikey",
					},
				}
				openapitest.AppendPath("/foo", c, openapitest.UseOperation("GET", op))

				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("Authorization", "Bearer 123")
				r.Header.Set("apikey", "API_KEY_123")
				rr := httptest.NewRecorder()
				h(rr, r)

				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"Bearer 123 - API_KEY_123"`, rr.Body.String())
			},
			event: func(request *common.HttpEventRequest, response *common.HttpEventResponse) []*common.Action {
				response.Data = fmt.Sprintf("%s - %s", request.Header["Authorization"], request.Header["apikey"])
				return nil
			},
		},
		{
			name: "oauth2 or api key required",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config, eh events.Handler) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithSecurity(map[string][]string{"bar": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json")),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.OAuth2SecurityScheme{},
					"bar": &openapi.ApiKeySecurityScheme{
						In:   "header",
						Name: "apikey",
					},
				}
				openapitest.AppendPath("/foo", c, openapitest.UseOperation("GET", op))

				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("apikey", "API_KEY_123")
				rr := httptest.NewRecorder()
				h(rr, r)

				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"API_KEY_123"`, rr.Body.String())
			},
			event: func(request *common.HttpEventRequest, response *common.HttpEventResponse) []*common.Action {
				response.Data = request.Header["apikey"]
				return nil
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e := &events.StoreManager{}
			e.SetStore(10, events.NewTraits().WithNamespace("http"))

			config := &openapi.Config{
				Info:       openapi.Info{Name: "Testing"},
				Servers:    []*openapi.Server{{Url: "http://localhost"}},
				Components: openapi.Components{},
			}

			h := enginetest.HttpEventHandlerFunc(tc.event)

			tc.test(t, func(rw http.ResponseWriter, r *http.Request) {
				h := openapi.NewHandler(config, h, e)
				h.ServeHTTP(rw, r)
			}, config, e)
		})

	}
}
