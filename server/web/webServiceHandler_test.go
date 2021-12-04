package web

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/test"
	"net/http"
	"net/url"
	"testing"
)

func TestResolveEndpoint(t *testing.T) {
	t.Parallel()
	t.Run("emptyConfig", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(&openapi.Config{})
		ctx := createHttpContext(t, "GET", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.EqualError(t, "no matching endpoint found", err)
	})
	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		op := &openapi.Operation{}
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Get: op,
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
		test.Equals(t, h.config.Info.Name, ctx.ServiceName)
		test.Equals(t, op, ctx.Operation)
	})
}

func TestResolveEndpointPathParameter(t *testing.T) {
	// path parameter is always required
	t.Run("optional", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo/{id}", &openapi.Endpoint{
			Get: &openapi.Operation{
				Parameters: []*openapi.ParameterRef{
					{Value: &openapi.Parameter{
						Name:     "id",
						Type:     openapi.PathParameter,
						Required: false,
					}},
				},
			},
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.EqualError(t, `no matching endpoint found`, err)
	})
	t.Run("requiredNotDefined", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo/{id}", &openapi.Endpoint{
			Get: &openapi.Operation{
				Parameters: []*openapi.ParameterRef{
					{Value: &openapi.Parameter{
						Name:     "id",
						Type:     openapi.PathParameter,
						Required: true,
					}},
				},
			},
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.EqualError(t, `no matching endpoint found`, err)
	})
	t.Run("required", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo/{id}", &openapi.Endpoint{
			Get: &openapi.Operation{
				Parameters: []*openapi.ParameterRef{
					{Value: &openapi.Parameter{
						Name:     "id",
						Type:     openapi.PathParameter,
						Required: true,
					}},
				},
			},
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo/1")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo/{id}", ctx.EndpointPath)
		test.Assert(t, len(ctx.Parameters[openapi.PathParameter]) == 1, "parameter parsed")
		test.Equals(t, "1", ctx.Parameters[openapi.PathParameter]["id"].Raw)
	})
}

func TestResolveEndpointQueryParameter(t *testing.T) {
	t.Run("optional", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Get: &openapi.Operation{
				Parameters: []*openapi.ParameterRef{
					{Value: &openapi.Parameter{
						Name:     "q",
						Type:     openapi.QueryParameter,
						Required: false,
					}},
				},
			},
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
	})
	t.Run("requiredNotDefined", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Get: &openapi.Operation{
				Parameters: []*openapi.ParameterRef{
					{Value: &openapi.Parameter{
						Name:     "q",
						Type:     openapi.QueryParameter,
						Required: true,
					}},
				},
			},
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.EqualError(t, `query parameter "q": required parameter not found`, err)
	})
	t.Run("required", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Get: &openapi.Operation{
				Parameters: []*openapi.ParameterRef{
					{Value: &openapi.Parameter{
						Name:     "q",
						Type:     openapi.QueryParameter,
						Required: true,
					}},
				},
			},
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo?q=bar")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
		test.Assert(t, len(ctx.Parameters[openapi.QueryParameter]) == 1, "parameter parsed")
		test.Equals(t, "bar", ctx.Parameters[openapi.QueryParameter]["q"].Raw)
	})
}

func TestResolveEndpointCookieParameter(t *testing.T) {
	t.Run("optional", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Get: &openapi.Operation{
				Parameters: []*openapi.ParameterRef{
					{Value: &openapi.Parameter{
						Name:     "q",
						Type:     openapi.CookieParameter,
						Required: false,
					}},
				},
			},
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
	})
	t.Run("requiredNotDefined", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Get: &openapi.Operation{
				Parameters: []*openapi.ParameterRef{
					{Value: &openapi.Parameter{
						Name:     "q",
						Type:     openapi.CookieParameter,
						Required: true,
					}},
				},
			},
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.EqualError(t, `cookie parameter "q": required parameter not found`, err)
	})
	t.Run("required", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Get: &openapi.Operation{
				Parameters: []*openapi.ParameterRef{
					{Value: &openapi.Parameter{
						Name:     "q",
						Type:     openapi.CookieParameter,
						Required: true,
					}},
				},
			},
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo?q=bar")
		ctx.Request.AddCookie(&http.Cookie{Name: "q", Value: "bar"})
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
		test.Assert(t, len(ctx.Parameters[openapi.CookieParameter]) == 1, "parameter parsed")
		test.Equals(t, "bar", ctx.Parameters[openapi.CookieParameter]["q"].Raw)
	})
}

func TestResolveEndpointHeaderParameter(t *testing.T) {
	t.Run("optional", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Get: &openapi.Operation{
				Parameters: []*openapi.ParameterRef{
					{Value: &openapi.Parameter{
						Name:     "q",
						Type:     openapi.HeaderParameter,
						Required: false,
					}},
				},
			},
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
	})
	t.Run("requiredNotDefined", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Get: &openapi.Operation{
				Parameters: []*openapi.ParameterRef{
					{Value: &openapi.Parameter{
						Name:     "q",
						Type:     openapi.HeaderParameter,
						Required: true,
					}},
				},
			},
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.EqualError(t, `header parameter "q": required parameter not found`, err)
	})
	t.Run("required", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Get: &openapi.Operation{
				Parameters: []*openapi.ParameterRef{
					{Value: &openapi.Parameter{
						Name:     "q",
						Type:     openapi.HeaderParameter,
						Required: true,
					}},
				},
			},
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo?q=bar")
		ctx.Request.Header.Set("q", "bar")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
		test.Assert(t, len(ctx.Parameters[openapi.HeaderParameter]) == 1, "parameter parsed")
		test.Equals(t, "bar", ctx.Parameters[openapi.HeaderParameter]["q"].Raw)
	})
}

func TestResolveEndpointHttpMethod(t *testing.T) {
	t.Parallel()
	t.Run("wrongMethod", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Post: &openapi.Operation{},
		}))
		ctx := createHttpContext(t, "GET", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.EqualError(t, "no matching endpoint found", err)
	})
	t.Run("get", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Get: &openapi.Operation{},
		}))
		ctx := createHttpContext(t, "get", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
	})
	t.Run("post", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Post: &openapi.Operation{},
		}))
		ctx := createHttpContext(t, "POST", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
	})
	t.Run("put", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Put: &openapi.Operation{},
		}))
		ctx := createHttpContext(t, "put", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
	})
	t.Run("patch", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Patch: &openapi.Operation{},
		}))
		ctx := createHttpContext(t, "patch", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
	})
	t.Run("delete", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Delete: &openapi.Operation{},
		}))
		ctx := createHttpContext(t, "delete", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
	})
	t.Run("head", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Head: &openapi.Operation{},
		}))
		ctx := createHttpContext(t, "head", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
	})
	t.Run("options", func(t *testing.T) {
		t.Parallel()
		h := NewWebServiceHandler(createConfig("http://localhost", "/foo", &openapi.Endpoint{
			Options: &openapi.Operation{},
		}))
		ctx := createHttpContext(t, "options", "http://localhost/foo")
		err := h.resolveEndpoint(ctx)
		test.Ok(t, err)
		test.Equals(t, "/foo", ctx.EndpointPath)
	})
}

func createConfig(host, path string, endpoint *openapi.Endpoint) *openapi.Config {
	return &openapi.Config{
		Info:    openapi.Info{Name: "Testing"},
		Servers: []*openapi.Server{{Url: host}},
		EndPoints: map[string]*openapi.EndpointRef{path: {
			Value: endpoint,
		}},
		Components: openapi.Components{},
	}
}

func createHttpContext(t *testing.T, method, requestUrl string) *HttpContext {
	u, err := url.Parse(requestUrl)
	test.Ok(t, err)
	return NewHttpContext(
		&http.Request{
			URL:    u,
			Method: method,
			Header: make(map[string][]string)},
		nil,
		nil,
	)
}
