package openapi_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/engine/common"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Security(t *testing.T) {
	testcases := []struct {
		name  string
		test  func(t *testing.T, h http.HandlerFunc, c *openapi.Config)
		event func(event string, args ...interface{}) []*common.Action
	}{
		{
			name: "basic",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.HttpSecurityScheme{
						Scheme: "basic",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.WithOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("Authorization", "Basic 123")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, `"123"`, rr.Body.String())
			},
			event: func(event string, args ...interface{}) []*common.Action {
				req := args[0].(*common.EventRequest)
				r := args[1].(*common.EventResponse)
				r.Data = req.Header["Authorization"]
				return nil
			},
		},
		{
			name: "bearer but without authorization header",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.HttpSecurityScheme{
						Scheme: "bearer",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.WithOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusForbidden, rr.Code)
			},
		},
		{
			name: "bearer with authorization header",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.HttpSecurityScheme{
						Scheme: "bearer",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.WithOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("Authorization", "Bearer 123")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"123"`, rr.Body.String())
			},
			event: func(event string, args ...interface{}) []*common.Action {
				req := args[0].(*common.EventRequest)
				r := args[1].(*common.EventResponse)
				r.Data = req.Header["Authorization"]
				return nil
			},
		},
		{
			name: "ApiKey in header",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.ApiKeySecurityScheme{
						In:   "header",
						Name: "X-API-KEY",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.WithOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("X-API-KEY", "123")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"123"`, rr.Body.String())
			},
			event: func(event string, args ...interface{}) []*common.Action {
				req := args[0].(*common.EventRequest)
				r := args[1].(*common.EventResponse)
				r.Data = req.Header["X-API-KEY"]
				return nil
			},
		},
		{
			name: "ApiKey in query",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.ApiKeySecurityScheme{
						In:   "query",
						Name: "apikey",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.WithOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo?apikey=123", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"123"`, rr.Body.String())
			},
			event: func(event string, args ...interface{}) []*common.Action {
				req := args[0].(*common.EventRequest)
				r := args[1].(*common.EventResponse)
				r.Data = req.Query["apikey"]
				return nil
			},
		},
		{
			name: "ApiKey in cookie",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.ApiKeySecurityScheme{
						In:   "cookie",
						Name: "apikey",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.WithOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.AddCookie(&http.Cookie{Name: "apikey", Value: "123"})
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"123"`, rr.Body.String())
			},
			event: func(event string, args ...interface{}) []*common.Action {
				req := args[0].(*common.EventRequest)
				r := args[1].(*common.EventResponse)
				r.Data = req.Cookie["apikey"]
				return nil
			},
		},
		{
			name: "security scheme not supported",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.NotSupportedSecurityScheme{
						Type: "NotSupportedSecurityScheme",
					},
				}

				openapitest.AppendPath("/foo", c, openapitest.WithOperation("GET", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
			},
		},
		{
			name: "oauth2",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.OAuth2SecurityScheme{},
				}
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("GET", op))

				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("Authorization", "Bearer 123")
				rr := httptest.NewRecorder()
				h(rr, r)

				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"Bearer 123"`, rr.Body.String())
			},
			event: func(event string, args ...interface{}) []*common.Action {
				req := args[0].(*common.EventRequest)
				r := args[1].(*common.EventResponse)
				r.Data = req.Header["Authorization"]
				return nil
			},
		},
		{
			name: "oauth2 and api key required",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(
						map[string][]string{
							"foo": {},
							"bar": {},
						},
					),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.OAuth2SecurityScheme{},
					"bar": &openapi.ApiKeySecurityScheme{
						In:   "header",
						Name: "apikey",
					},
				}
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("GET", op))

				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("Authorization", "Bearer 123")
				r.Header.Set("apikey", "API_KEY_123")
				rr := httptest.NewRecorder()
				h(rr, r)

				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"Bearer 123 - API_KEY_123"`, rr.Body.String())
			},
			event: func(event string, args ...interface{}) []*common.Action {
				req := args[0].(*common.EventRequest)
				r := args[1].(*common.EventResponse)
				r.Data = fmt.Sprintf("%s - %s", req.Header["Authorization"], req.Header["apikey"])
				return nil
			},
		},
		{
			name: "oauth2 or api key required",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithSecurity(map[string][]string{"foo": {}}),
					openapitest.WithSecurity(map[string][]string{"bar": {}}),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
				)
				c.Components.SecuritySchemes = map[string]openapi.SecurityScheme{
					"foo": &openapi.OAuth2SecurityScheme{},
					"bar": &openapi.ApiKeySecurityScheme{
						In:   "header",
						Name: "apikey",
					},
				}
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("GET", op))

				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("apikey", "API_KEY_123")
				rr := httptest.NewRecorder()
				h(rr, r)

				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"API_KEY_123"`, rr.Body.String())
			},
			event: func(event string, args ...interface{}) []*common.Action {
				req := args[0].(*common.EventRequest)
				r := args[1].(*common.EventResponse)
				r.Data = req.Header["apikey"]
				return nil
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := &openapi.Config{
				Info:       openapi.Info{Name: "Testing"},
				Servers:    []*openapi.Server{{Url: "http://localhost"}},
				Components: openapi.Components{},
			}

			tc.test(t, func(rw http.ResponseWriter, r *http.Request) {
				h := openapi.NewHandler(config, &engine{emit: tc.event})
				h.ServeHTTP(rw, r)
			}, config)
		})

	}
}
