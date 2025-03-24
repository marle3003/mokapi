package openapi_test

import (
	"context"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/script"
	"mokapi/engine/common"
	"mokapi/engine/enginetest"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime/events"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestResolveEndpoint(t *testing.T) {
	testdata := []struct {
		name string
		test func(t *testing.T, h http.HandlerFunc, c *openapi.Config)
	}{
		{
			name: "wrong hostname",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				r := httptest.NewRequest("GET", "https://foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 404, rr.Code)
				require.Equal(t, "no matching endpoint found: GET https://foo\n", rr.Body.String())
			},
		},
		{
			name: "base path",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				c.Servers[0].Url = "http://localhost/root"
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/root/foo", nil)
				r = r.WithContext(context.WithValue(r.Context(), "servicePath", "/root"))
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			},
		},
		{
			name: "base path single slash",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				c.Servers[0].Url = "http://localhost/root"
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/root", nil)
				r = r.WithContext(context.WithValue(r.Context(), "servicePath", "/root"))
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			},
		},
		{
			// there is no official specification for trailing slash. For ease of use, mokapi considers it equivalent
			name: "spec define suffix / but request does not",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				c.Servers[0].Url = "http://localhost"
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo/", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			},
		},
		{
			// there is no official specification for trailing slash. For ease of use, mokapi considers it equivalent
			name: "spec define suffix no / but request does",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				c.Servers[0].Url = "http://localhost"
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo/", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			},
		},
		{
			// there is no official specification for trailing slash. For ease of use, mokapi considers it equivalent
			name: "both uses trailing slash",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				c.Servers[0].Url = "http://localhost"
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo/", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo/", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			},
		},
		//
		// GET
		//
		{
			name: "no endpoint",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 404, rr.Code)
			},
		},
		{
			name: "no success response specified",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusNotFound, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 500, rr.Code)
				require.Equal(t, "neither success response (HTTP 2xx) nor 'default' response found\n", rr.Body.String())
			},
		},
		{
			name: "no response specified",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation()
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 500, rr.Code)
				require.Equal(t, "neither success response (HTTP 2xx) nor 'default' response found\n", rr.Body.String())
			},
		},
		{
			name: "no success response specified but default",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				// 0 = default
				op := openapitest.NewOperation(openapitest.WithResponse(0, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{
			name: "with endpoint",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			},
		},
		{
			name: "with multiple success response 1/2",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusNoContent, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithResponse(http.StatusAccepted, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 204, rr.Code)
			},
		},
		{
			name: "with multiple success response 2/2",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusAccepted, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithResponse(http.StatusNoContent, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 202, rr.Code)
			},
		},
		{
			name: "empty response body",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusNoContent))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 204, rr.Code)
			},
		},
		//
		// POST
		//
		{
			name: "POST request",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("POST", op))
				r := httptest.NewRequest("POST", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{
			name: "POST request invalid data",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithRequestBody("", true,
						openapitest.WithRequestContent("application/json",
							openapitest.NewContent(openapitest.WithSchema(schematest.New("string", schematest.WithMinLength(4)))))))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("POST", op))
				r := httptest.NewRequest("POST", "http://localhost/foo", strings.NewReader(`"foo"`))
				r.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 500, rr.Code)
				require.Equal(t, "read request body 'application/json' failed: error count 1:\n\t- #/minLength: string 'foo' is less than minimum of 4\n", rr.Body.String())
			},
		},
		//
		// PUT
		//
		{
			name: "PUT request",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("PUT", op))
				r := httptest.NewRequest("PUT", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// PATCH
		//
		{
			name: "PATCH request",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("PATCH", op))
				r := httptest.NewRequest("PATCH", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// DELETE
		//
		{
			name: "DELETE request",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("DELETE", op))
				r := httptest.NewRequest("DELETE", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// HEAD
		//
		{
			name: "HEAD request",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("HEAD", op))
				r := httptest.NewRequest("HEAD", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// OPTIONS
		//
		{
			name: "OPTIONS request",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("OPTIONS", op))
				r := httptest.NewRequest("OPTIONS", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// TRACE
		//
		{
			name: "TRACE request",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("TRACE", op))
				r := httptest.NewRequest("TRACE", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// Path parameter
		//
		{
			name: "path is always required",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithOperationParam("id", false))
				openapitest.AppendPath("/foo/{id}", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 404, rr.Code)
				require.Equal(t, "no matching endpoint found: GET http://localhost/foo\n", rr.Body.String())
			},
		},
		{
			name: "segment of path not match",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithOperationParam("id", false))
				openapitest.AppendPath("/foo/{id}/bar", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 404, rr.Code)
				require.Equal(t, "no matching endpoint found: GET http://localhost/foo\n", rr.Body.String())
			},
		},
		{
			name: "with path parameter present",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithOperationParam("id", false))
				openapitest.AppendPath("/foo/{id}", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo/42", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{
			name: "path parameter not present in endpoint path",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithOperationParam("id", false))
				openapitest.AppendPath("/foo/bar", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo/bar", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 400, rr.Code)
				require.Equal(t, "parse path parameter 'id' failed: parameter is required\n", rr.Body.String())
			},
		},
		//
		// Query parameter
		//
		{
			name: "with optional query parameter and not present",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithQueryParam("id", false))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{
			name: "with required query parameter and not present",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithQueryParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 400, rr.Code)
				require.Equal(t, "parse query parameter 'id' failed: parameter is required\n", rr.Body.String())
			},
		},
		{
			name: "with required query parameter and present",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithQueryParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo?id=42", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// Cookie parameter
		//
		{
			name: "with optional query parameter and not present",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithCookieParam("id", false))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{
			name: "with required query parameter and not present",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithCookieParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 400, rr.Code)
				require.Equal(t, "parse cookie parameter 'id' failed: parameter is required\n", rr.Body.String())
			},
		},
		{
			name: "with required query parameter and present",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithCookieParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.AddCookie(&http.Cookie{Name: "id", Value: "42"})
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// Header parameter
		//
		{
			name: "with optional query parameter and not present",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", false))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		{
			name: "with required query parameter and not present",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 400, rr.Code)
				require.Equal(t, "parse header parameter 'id' failed: parameter is required\n", rr.Body.String())
			},
		},
		{
			name: "with required query parameter and present",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
			},
		},
		//
		// content-type
		//
		{
			name: "with content-type",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("content-type"))
			},
		},
		{
			name: "with content-type extensions",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json;odata=verbose", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json;odata=verbose")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json;odata=verbose", rr.Header().Get("content-type"))
			},
		},
		{
			name: "with content-type extensions",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json;odata=verbose")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("content-type"))
			},
		},
		{
			name: "with content-type extensions exactly",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json;odata=verbose", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json;odata=verbose")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json;odata=verbose", rr.Header().Get("content-type"))
			},
		},
		{
			name: "with content-type multiple accepted",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "text/plain,application/json")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 200, rr.Code)
				require.Equal(t, "application/json", rr.Header().Get("content-type"))
			},
		},
		{
			name: "with content-type not supported",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("text/plain", openapitest.NewContent())),
					openapitest.WithHeaderParam("id", true))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("id", "42")
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, 415, rr.Code)
				require.Equal(t, "none of requests content type(s) are supported: \"application/json\"\n", rr.Body.String())
			},
		},
		{
			// endpoint /pet/{petId} and /pet/findByStatus overlaps in segments but is different by type
			// /pet/1
			// /pet/findByStatus
			name: "endpoints overlap",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
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
				h(rr, r)
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

			data.test(t, func(rw http.ResponseWriter, r *http.Request) {
				h := openapi.NewHandler(config, &engine{})
				h.ServeHTTP(rw, r)
			}, config)
		})

	}
}

func TestHandler_Event(t *testing.T) {
	testcases := []struct {
		name  string
		test  func(t *testing.T, f http.HandlerFunc, c *openapi.Config)
		event func(event string, args ...interface{}) []*common.Action
	}{
		{
			name: "no response found",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusInternalServerError, rr.Code)
				require.Equal(t, "no configuration was found for HTTP status code 415, https://swagger.io/docs/specification/describing-responses\n", rr.Body.String())
			},
			event: func(event string, args ...interface{}) []*common.Action {
				r := args[1].(*common.EventResponse)
				r.StatusCode = http.StatusUnsupportedMediaType
				return nil
			},
		},
		{
			name: "event sets unknown status code",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusInternalServerError, rr.Code)
				require.Equal(t, "no configuration was found for HTTP status code 415, https://swagger.io/docs/specification/describing-responses\n", rr.Body.String())
			},
			event: func(event string, args ...interface{}) []*common.Action {
				r := args[1].(*common.EventResponse)
				r.StatusCode = http.StatusUnsupportedMediaType
				return nil
			},
		},
		{
			name: "event changes content type",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK,
						openapitest.WithContent("application/json", openapitest.NewContent()),
						openapitest.WithContent("text/plain", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, "text/plain", rr.Header().Get("Content-Type"))
			},
			event: func(event string, args ...interface{}) []*common.Action {
				r := args[1].(*common.EventResponse)
				r.Headers["Content-Type"] = "text/plain"
				r.Body = "Hello"
				return nil
			},
		},
		{
			name: "post request using body in event function",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithRequestBody("", true, openapitest.WithRequestContent("application/json",
						openapitest.NewContent())),
					openapitest.WithResponse(http.StatusOK,
						openapitest.WithContent("application/json", openapitest.NewContent()),
					))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation(http.MethodPost, op))
				r := httptest.NewRequest("post", "http://localhost/foo", strings.NewReader(`{ "foo": "bar" }`))
				r.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `{"foo":"bar"}`, rr.Body.String())
			},
			event: func(event string, args ...interface{}) []*common.Action {
				req := args[0].(*common.EventRequest)
				res := args[1].(*common.EventResponse)
				res.Data = req.Body
				return nil
			},
		},
		{
			name: "post request without defining requestBody, body should not be available in event",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK,
						openapitest.WithContent("application/json", openapitest.NewContent()),
					))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation(http.MethodPost, op))
				r := httptest.NewRequest("post", "http://localhost/foo", strings.NewReader(`{ "foo": "bar" }`))
				r.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, "", rr.Body.String())
			},
			event: func(event string, args ...interface{}) []*common.Action {
				req := args[0].(*common.EventRequest)
				res := args[1].(*common.EventResponse)
				res.Data = req.Body
				return nil
			},
		},
		{
			name: "path parameter",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK,
						openapitest.WithContent("application/json", openapitest.NewContent()),
					))
				openapitest.AppendPath("/foo/{id}", c,
					openapitest.WithOperation(http.MethodPost, op),
					openapitest.WithPathParam("id", openapitest.WithParamSchema(schematest.New("string"))),
				)
				r := httptest.NewRequest("post", "http://localhost/foo/123", strings.NewReader(`{ "foo": "bar" }`))
				r.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"123"`, rr.Body.String())
			},
			event: func(event string, args ...interface{}) []*common.Action {
				req := args[0].(*common.EventRequest)
				res := args[1].(*common.EventResponse)
				res.Data = req.Path["id"]
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

			tc.test(t, func(rw http.ResponseWriter, r *http.Request) {
				h := openapi.NewHandler(config, &engine{emit: tc.event})
				h.ServeHTTP(rw, r)
			}, config)
		})

	}
}

func TestHandler_Log(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, h http.HandlerFunc, c *openapi.Config)
	}{
		{
			name: "simple",
			test: func(t *testing.T, h http.HandlerFunc, c *openapi.Config) {
				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("GET", "http://localhost/foo", nil)
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				h(rr, r)

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

			tc.test(t, func(rw http.ResponseWriter, r *http.Request) {
				h := openapi.NewHandler(config, &engine{})
				h.ServeHTTP(rw, r)
			}, config)
		})

	}
}

func TestHandler_Event_TypeScript(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "async event handler",
			test: func(t *testing.T) {
				e := enginetest.NewEngine()
				err := e.AddScript(newScript("test.ts", `
					import {on, sleep} from 'mokapi'
					export default function() {
						on('http', async (request, response) => {
							response.data = await getData()
						});
					}
					let getData = async () => {
						return new Promise(async (resolve, reject) => {
						  setTimeout(() => {
							resolve('foo');
						  }, 800);
						});
					}
				`))
				require.NoError(t, err)

				config := &openapi.Config{
					Info:       openapi.Info{Name: "Testing"},
					Servers:    []*openapi.Server{{Url: "http://localhost"}},
					Components: openapi.Components{},
				}

				h := func(rw http.ResponseWriter, r *http.Request) {
					h := openapi.NewHandler(config, e)
					h.ServeHTTP(rw, r)
				}

				op := openapitest.NewOperation(
					openapitest.WithResponse(http.StatusOK, openapitest.WithContent("application/json", openapitest.NewContent())))
				openapitest.AppendPath("/foo", config, openapitest.WithOperation("get", op))
				r := httptest.NewRequest("get", "http://localhost/foo", nil)
				r.Header.Set("accept", "application/json")
				rr := httptest.NewRecorder()
				h(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `"foo"`, rr.Body.String())
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.test(t)
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

func newScript(path, src string) dynamic.ConfigEvent {
	return dynamic.ConfigEvent{Config: &dynamic.Config{
		Info: dynamic.ConfigInfo{Url: mustParse(path)},
		Raw:  []byte(src),
		Data: &script.Script{Code: src, Filename: path},
	}}
}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
