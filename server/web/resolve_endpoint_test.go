package web_test

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/engine"
	"mokapi/models"
	"mokapi/server/web"
	"mokapi/test"
	"net/http"
	"net/http/httptest"
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
				test.Equals(t, 404, rr.Code)
				test.Equals(t, "There was no service listening at https://foo\n", rr.Body.String())
			}},
		//
		// GET
		//
		{"no endpoint",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 404, rr.Code)
			}},
		{"no success response specified",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation()
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 500, rr.Code)
				test.Equals(t, "no success response (HTTP 2xx) in configuration\n", rr.Body.String())
			}},
		{"with endpoint",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(openapi.OK))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		//
		// POST
		//
		{"POST request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(openapi.OK))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("POST", op))
				r := httptest.NewRequest("POST", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		//
		// PUT
		//
		{"PUT request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(openapi.OK))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("PUT", op))
				r := httptest.NewRequest("PUT", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		//
		// PATCH
		//
		{"PATCH request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(openapi.OK))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("PATCH", op))
				r := httptest.NewRequest("PATCH", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		//
		// DELETE
		//
		{"DELETE request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(openapi.OK))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("DELETE", op))
				r := httptest.NewRequest("DELETE", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		//
		// HEAD
		//
		{"HEAD request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(openapi.OK))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("HEAD", op))
				r := httptest.NewRequest("HEAD", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		//
		// OPTIONS
		//
		{"OPTIONS request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(openapi.OK))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("OPTIONS", op))
				r := httptest.NewRequest("OPTIONS", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		//
		// TRACE
		//
		{"TRACE request",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(openapi.OK))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("TRACE", op))
				r := httptest.NewRequest("TRACE", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		//
		// Path parameter
		//
		{"path is always required",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithPathParam("id", false))
				openapitest.AppendEndpoint("/foo/{id}", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 404, rr.Code)
				test.Equals(t, "unable to serve http request of API Testing: no matching endpoint found\n", rr.Body.String())
			}},
		{"segment of path not match",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithPathParam("id", false))
				openapitest.AppendEndpoint("/foo/{id}/bar", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 404, rr.Code)
				test.Equals(t, "unable to serve http request of API Testing: no matching endpoint found\n", rr.Body.String())
			}},
		{"with path parameter present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithPathParam("id", false))
				openapitest.AppendEndpoint("/foo/{id}", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo/42", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		{"path parameter not present in endpoint path",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithPathParam("id", false))
				openapitest.AppendEndpoint("/foo/bar", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo/bar", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 400, rr.Code)
				test.Equals(t, "unable to serve http request of API Testing: required path parameter id not present\n", rr.Body.String())
			}},
		//
		// Query parameter
		//
		{"with optional query parameter and not present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithQueryParam("id", false))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		{"with required query parameter and not present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithQueryParam("id", true))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 400, rr.Code)
				test.Equals(t, "unable to serve http request of API Testing: query parameter id: required parameter not found\n", rr.Body.String())
			}},
		{"with required query parameter and present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithQueryParam("id", true))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo?id=42", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		//
		// Cookie parameter
		//
		{"with optional query parameter and not present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithCookieParam("id", false))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		{"with required query parameter and not present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithCookieParam("id", true))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 400, rr.Code)
				test.Equals(t, "unable to serve http request of API Testing: cookie parameter id: required parameter not found\n", rr.Body.String())
			}},
		{"with required query parameter and present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithCookieParam("id", true))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.AddCookie(&http.Cookie{Name: "id", Value: "42"})
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		//
		// Header parameter
		//
		{"with optional query parameter and not present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithHeaderParam("id", false))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		{"with required query parameter and not present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 400, rr.Code)
				test.Equals(t, "unable to serve http request of API Testing: header parameter id: required parameter not found\n", rr.Body.String())
			}},
		{"with required query parameter and present",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
			}},
		//
		// content-type
		//
		{"with content-type",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK, openapitest.WithContent("application/json")),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
				test.Equals(t, "application/json", rr.Header().Get("content-type"))
			}},
		{"with content-type extensions",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK, openapitest.WithContent("application/json;odata=verbose")),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json;odata=verbose")
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
				test.Equals(t, "application/json;odata=verbose", rr.Header().Get("content-type"))
			}},
		{"with content-type extensions",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK, openapitest.WithContent("application/json")),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json;odata=verbose")
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
				test.Equals(t, "application/json;odata=verbose", rr.Header().Get("content-type"))
			}},
		{"with content-type multiple accepted",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK, openapitest.WithContent("application/json")),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "text/plain,application/json")
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 200, rr.Code)
				test.Equals(t, "application/json", rr.Header().Get("content-type"))
			}},
		{"with content-type not supported",
			func(t *testing.T, f serveHTTP, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(openapi.OK),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendEndpoint("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				f(rr, r)
				test.Equals(t, 415, rr.Code)
				test.Equals(t, "none of requests content type(s) are supported: application/json\n", rr.Body.String())
			}},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			test.NewNullLogger()

			b := web.NewBinding(":80", func(metric *models.RequestMetric) {
			}, func(s string, i ...interface {
			}) []*engine.Summary {
				return nil
			})
			config := &openapi.Config{
				Info:       openapi.Info{Name: "Testing"},
				Servers:    []*openapi.Server{{Url: "http://localhost"}},
				Components: openapi.Components{},
			}

			data.fn(t, func(rw http.ResponseWriter, r *http.Request) {
				err := b.Apply(config)
				test.Ok(t, err)
				b.ServeHTTP(rw, r)
			}, config)
		})

	}
}
