package openapi_test

import (
	"context"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"mokapi/engine/common"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime/events"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type serveHTTP func(rw http.ResponseWriter, r *http.Request)

func TestResolveEndpoint(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, f serveHTTP, c *openapi.Config)
	}{
		{"wrong hostname",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				r := httptest.NewRequest("GET", "https://foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 404, rr.Code)
				require.Equal(t, "no matching endpoint found: GET https://foo\n", rr.Body.String())
			},
		},
		{"base path",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				c.Servers[0].Url = "http://localhost/root"
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/root/foo", nil)
				r = r.WithContext(context.WithValue(r.Context(), "servicePath", "/root"))
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			},
		},
		{"base path single slash",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				c.Servers[0].Url = "http://localhost/root"
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/root", nil)
				r = r.WithContext(context.WithValue(r.Context(), "servicePath", "/root"))
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			},
		},
		//
		// GET
		//
		{"no endpoint",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 404, rr.Code)
			},
		},
		{"no success response specified",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation()
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 500, rr.Code)
				require.Equal(t, "no success response (HTTP 2xx) in configuration\n", rr.Body.String())
			},
		},
		{"with endpoint",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			},
		},
		{"with multiple success response 1/2",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusNoContent, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithResponse(http.StatusAccepted, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 204, rr.Code)
			},
		},
		{"with multiple success response 2/2",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusAccepted, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithResponse(http.StatusNoContent, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 202, rr.Code)
			},
		},
		{"empty response body",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusNoContent))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 204, rr.Code)
			},
		},
		//
		// POST
		//
		{"POST request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("POST", op))
				r := httptest.NewRequest("POST", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{"POST request invalid data",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithRequestBody("", true,
						openapitest.WithRequestContent("application/json",
							openapitest.NewContent(openapitest.WithSchema(schematest.New("string", schematest.WithMinLength(4)))))))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("POST", op))
				r := httptest.NewRequest("POST", "http://localhost/foo", strings.NewReader(`"foo"`))
				r.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 500, rr.Code)
				require.Equal(t, "read request body 'application/json' failed: length of 'foo' is too short, expected schema type=string minLength=4\n", rr.Body.String())
			},
		},
		//
		// PUT
		//
		{"PUT request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("PUT", op))
				r := httptest.NewRequest("PUT", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// PATCH
		//
		{"PATCH request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("PATCH", op))
				r := httptest.NewRequest("PATCH", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// DELETE
		//
		{"DELETE request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("DELETE", op))
				r := httptest.NewRequest("DELETE", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// HEAD
		//
		{"HEAD request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("HEAD", op))
				r := httptest.NewRequest("HEAD", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// OPTIONS
		//
		{"OPTIONS request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("OPTIONS", op))
				r := httptest.NewRequest("OPTIONS", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// TRACE
		//
		{"TRACE request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("TRACE", op))
				r := httptest.NewRequest("TRACE", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// Path parameter
		//
		{"path is always required",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithOperationParam("id", false))
				openapitest.AppendPath("/foo/{id}", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 404, rr.Code)
				require.Equal(t, "no matching endpoint found: GET http://localhost/foo\n", rr.Body.String())
			},
		},
		{"segment of path not match",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithOperationParam("id", false))
				openapitest.AppendPath("/foo/{id}/bar", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 404, rr.Code)
				require.Equal(t, "no matching endpoint found: GET http://localhost/foo\n", rr.Body.String())
			},
		},
		{"with path parameter present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithOperationParam("id", false))
				openapitest.AppendPath("/foo/{id}", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo/42", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{"path parameter not present in endpoint path",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithOperationParam("id", false))
				openapitest.AppendPath("/foo/bar", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo/bar", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 400, rr.Code)
				require.Equal(t, "parse path parameter 'id' failed: parameter is required\n", rr.Body.String())
			},
		},
		//
		// Query parameter
		//
		{"with optional query parameter and not present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithQueryParam("id", false))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{"with required query parameter and not present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithQueryParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 400, rr.Code)
				require.Equal(t, "parse query parameter 'id' failed: parameter is required\n", rr.Body.String())
			},
		},
		{"with required query parameter and present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithQueryParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo?id=42", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// Cookie parameter
		//
		{"with optional query parameter and not present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithCookieParam("id", false))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{"with required query parameter and not present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithCookieParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 400, rr.Code)
				require.Equal(t, "parse cookie parameter 'id' failed: parameter is required\n", rr.Body.String())
			},
		},
		{"with required query parameter and present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithCookieParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.AddCookie(&http.Cookie{Name: "id", Value: "42"})
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// Header parameter
		//
		{"with optional query parameter and not present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", false))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{"with required query parameter and not present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 400, rr.Code)
				require.Equal(t, "parse header parameter 'id' failed: parameter is required\n", rr.Body.String())
			},
		},
		{"with required query parameter and present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// content-type
		//
		{"with content-type",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("content-type"))
			},
		},
		{"with content-type extensions",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json;odata=verbose", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json;odata=verbose")
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json;odata=verbose", rr.Header().Get("content-type"))
			},
		},
		{"with content-type extensions",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json;odata=verbose")
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("content-type"))
			},
		},
		{"with content-type extensions exactly",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json;odata=verbose", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json;odata=verbose")
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json;odata=verbose", rr.Header().Get("content-type"))
			},
		},
		{"with content-type multiple accepted",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "text/plain,application/json")
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("content-type"))
			},
		},
		{"with content-type not supported",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("text/plain", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 415, rr.Code)
				require.Equal(t, "none of requests content type(s) are supported: \"application/json\"\n", rr.Body.String())
			},
		},
		{
			// endpoint /pet/{petId} and /pet/findByStatus overlaps in segments but is different by type
			// /pet/1
			// /pet/findByStatus
			"endpoints overlap",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				byId := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithOperationParam("petId", true, openapitest.WithParamSchema(schematest.New("integer"))))
				find := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/pet/{petId}", c, openapitest.WithOperation("get", byId))
				openapitest.AppendPath("/pet/findByStatus", c, openapitest.WithOperation("get", find))
				r := httptest.NewRequest("get", "http://localhost/pet/findByStatus", nil)
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			test.NewNullLogger()
			events.SetStore(10, events.NewTraits().WithNamespace("http"))
			defer events.Reset()

			config := &openapi.Config{
				Info:       openapi.Info{Name: "Testing"},
				Servers:    []*openapi.Server{{Url: "http://localhost"}},
				Components: openapi.Components{},
			}

			data.fn(t, func(rw http.ResponseWriter, r *http.Request) {
				h := openapi.NewHandler(config, &engine{})
				h.ServeHTTP(rw, r)
			}, config)
		})

	}
}

func TestHandler_Event(t *testing.T) {
	testcases := []struct {
		name  string
		fn    func(t *testing.T, f serveHTTP, c *openapi.Config)
		event func(event string, args ...interface{}) []*common.Action
	}{
		{
			"no response found",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, http.StatusInternalServerError, rr.Code)
				require.Equal(t, "no configuration was found for HTTP status code 415, https://swagger.io/docs/specification/describing-responses\n", rr.Body.String())
			},
			func(event string, args ...interface{}) []*common.Action {
				r := args[1].(*common.EventResponse)
				r.StatusCode = http.StatusUnsupportedMediaType
				return nil
			},
		},
		{
			"event sets unknown status code",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, http.StatusInternalServerError, rr.Code)
				require.Equal(t, "no configuration was found for HTTP status code 415, https://swagger.io/docs/specification/describing-responses\n", rr.Body.String())
			},
			func(event string, args ...interface{}) []*common.Action {
				r := args[1].(*common.EventResponse)
				r.StatusCode = http.StatusUnsupportedMediaType
				return nil
			},
		},
		{
			"event changes content type",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK,
						openapitest.WithContent("application/json", openapitest.NewContent()),
						openapitest.WithContent("text/plain", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				f(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
			},
			func(event string, args ...interface{}) []*common.Action {
				r := args[1].(*common.EventResponse)
				r.Headers["Content-Type"] = "text/plain"
				r.Body = "Hello"
				return nil
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			test.NewNullLogger()

			config := &openapi.Config{
				Info:       openapi.Info{Name: "Testing"},
				Servers:    []*openapi.Server{{Url: "http://localhost"}},
				Components: openapi.Components{},
			}

			tc.fn(t, func(rw http.ResponseWriter, r *http.Request) {
				h := openapi.NewHandler(config, &engine{emit: tc.event})
				h.ServeHTTP(rw, r)
			}, config)
		})

	}
}

func TestHandler_Log(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, f serveHTTP, c *openapi.Config)
	}{
		{
			"simple",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				f(rr, r)

				logs := events.GetEvents(events.NewTraits().WithNamespace("http"))
				require.Len(t, logs, 1)
				log := logs[0]
				require.NotEmpty(t, log.Id)
				httpLog := log.Data.(*openapi.HttpLog)
				require.Equal(t, "GET", httpLog.Request.Method)
				require.Equal(t, "http://localhost/foo", httpLog.Request.Url)
				require.Equal(t, "application/json", httpLog.Response.Headers["Content-Type"])
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			test.NewNullLogger()
			events.SetStore(10, events.NewTraits().WithNamespace("http"))
			defer events.Reset()

			config := &openapi.Config{
				Info:       openapi.Info{Name: "Testing"},
				Servers:    []*openapi.Server{{Url: "http://localhost"}},
				Components: openapi.Components{},
			}

			tc.fn(t, func(rw http.ResponseWriter, r *http.Request) {
				h := openapi.NewHandler(config, &engine{})
				h.ServeHTTP(rw, r)
			}, config)
		})

	}
}

type engine struct {
	emit func(event string, args ...interface{}) []*common.Action
}

func (e *engine) Emit(event string, args ...interface{}) []*common.Action {
	if e.emit != nil {
		return e.emit(event, args...)
	}
	return nil
}
