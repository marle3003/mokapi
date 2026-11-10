package openapi_test

import (
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema"
	"mokapi/providers/openapi/schema/schematest"
	"testing"

	"github.com/stretchr/testify/require"
)

func getOperation(method string) openapitest.PathOptions {
	return openapitest.WithOperation(method, openapitest.WithResponse(200))
}

func TestConfig_ResolveOperation(t *testing.T) {
	testcases := []struct {
		name   string
		config *openapi.Config
		method string
		url    string
		path   string // expected path of matched operation, empty if none
	}{
		{
			name: "relative server url",
			config: openapitest.NewConfig("3.0",
				openapitest.WithServer("/api/v1", ""),
				openapitest.WithPath("/profile", getOperation("get")),
			),
			method: "GET", url: "/api/v1/profile", path: "/profile",
		},
		{
			name: "absolute server url",
			config: openapitest.NewConfig("3.0",
				openapitest.WithServer("http://localhost:8080/api/v1", ""),
				openapitest.WithPath("/profile", getOperation("get")),
			),
			method: "GET", url: "/api/v1/profile", path: "/profile",
		},
		{
			name: "no servers",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/profile", getOperation("get")),
			),
			method: "GET", url: "/profile", path: "/profile",
		},
		{
			name: "server url with trailing slash",
			config: openapitest.NewConfig("3.0",
				openapitest.WithServer("/api/v1/", ""),
				openapitest.WithPath("/profile", getOperation("get")),
			),
			method: "GET", url: "/api/v1/profile", path: "/profile",
		},
		{
			name: "base path is case-insensitive",
			config: openapitest.NewConfig("3.0",
				openapitest.WithServer("/api/v1", ""),
				openapitest.WithPath("/profile", getOperation("get")),
			),
			method: "GET", url: "/API/V1/profile", path: "/profile",
		},
		{
			name: "trailing slash on request",
			config: openapitest.NewConfig("3.0",
				openapitest.WithServer("/api/v1", ""),
				openapitest.WithPath("/profile", getOperation("get")),
			),
			method: "GET", url: "/api/v1/profile/", path: "/profile",
		},
		{
			name: "query is ignored for lookup",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/profile", getOperation("get")),
			),
			method: "GET", url: "/profile?id=1", path: "/profile",
		},
		{
			name: "origin is ignored",
			config: openapitest.NewConfig("3.0",
				openapitest.WithServer("/api/v1", ""),
				openapitest.WithPath("/profile", getOperation("get")),
			),
			method: "GET", url: "http://localhost:6006/api/v1/profile?id=1", path: "/profile",
		},
		{
			name: "root path with base path",
			config: openapitest.NewConfig("3.0",
				openapitest.WithServer("/api/v1", ""),
				openapitest.WithPath("/", getOperation("get")),
			),
			method: "GET", url: "/api/v1", path: "/",
		},
		{
			name: "path is not stripped when no server matches",
			config: openapitest.NewConfig("3.0",
				openapitest.WithServer("/api/v1", ""),
				openapitest.WithPath("/other/profile", getOperation("get")),
			),
			method: "GET", url: "/other/profile", path: "/other/profile",
		},
		{
			name: "longest base path wins",
			config: openapitest.NewConfig("3.0",
				openapitest.WithServer("/api", ""),
				openapitest.WithServer("/api/v1", ""),
				openapitest.WithPath("/profile", getOperation("get")),
			),
			method: "GET", url: "/api/v1/profile", path: "/profile",
		},
		{
			name: "shorter base path is tried when the longest finds nothing",
			config: openapitest.NewConfig("3.0",
				openapitest.WithServer("/api", ""),
				openapitest.WithServer("/api/v1", ""),
				openapitest.WithPath("/v1/profile", getOperation("get")),
			),
			method: "GET", url: "/api/v1/profile", path: "/v1/profile",
		},
		{
			name: "literal path beats parameterized path",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users/{id}", getOperation("get")),
				openapitest.WithPath("/users/me", getOperation("get")),
			),
			method: "GET", url: "/users/me", path: "/users/me",
		},
		{
			name: "parameterized path",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users/{id}", getOperation("get")),
			),
			method: "GET", url: "/users/12", path: "/users/{id}",
		},
		{
			name: "server variables do not match as base path",
			config: openapitest.NewConfig("3.0",
				openapitest.WithServer("/api/{version}", ""),
				openapitest.WithPath("/profile", getOperation("get")),
			),
			method: "GET", url: "/api/v1/profile",
		},
		{
			name: "method not defined",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/profile", getOperation("get")),
			),
			method: "POST", url: "/profile",
		},
		{
			name: "path not defined",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/profile", getOperation("get")),
			),
			method: "GET", url: "/unknown",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := tc.config.ResolveOperation(tc.method, tc.url)
			require.NoError(t, err)
			if tc.path == "" {
				require.Nil(t, match)
				return
			}
			require.NotNil(t, match)
			require.Equal(t, tc.path, match.Operation.Path.Path)
		})
	}
}

func TestConfig_ResolveOperation_InvalidUrl(t *testing.T) {
	config := openapitest.NewConfig("3.0")
	_, err := config.ResolveOperation("GET", "http://[::1")
	require.Error(t, err)
}

func validateRequest(t *testing.T, config *openapi.Config, method, url string, headers map[string]string, body string) []openapi.ValidationError {
	t.Helper()
	match, err := config.ResolveOperation(method, url)
	require.NoError(t, err)
	require.NotNil(t, match, "operation not found")
	return match.ValidateRequest(headers, body)
}

func locations(errs []openapi.ValidationError) []string {
	result := make([]string, 0, len(errs))
	for _, e := range errs {
		result = append(result, e.Location)
	}
	return result
}

func TestOperationMatch_ValidateRequest(t *testing.T) {
	jsonBody := func(s *schema.Schema) openapitest.OperationOptions {
		return openapitest.WithRequestBody("", true,
			openapitest.WithRequestContent("application/json", openapitest.WithSchema(s)))
	}
	user := schematest.New("object",
		schematest.WithProperty("name", schematest.New("string")),
		schematest.WithRequired("name"))

	testcases := []struct {
		name    string
		config  *openapi.Config
		method  string
		url     string
		headers map[string]string
		body    string
		test    func(t *testing.T, errs []openapi.ValidationError)
	}{
		{
			name: "valid request",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users/{id}", openapitest.WithOperation("post",
					openapitest.WithOperationParam("id", true, openapitest.WithParamSchema(schematest.New("integer"))),
					openapitest.WithQueryParam("q", true),
					openapitest.WithHeaderParam("X-Id", true),
					openapitest.WithCookieParam("session", true),
					jsonBody(user),
					openapitest.WithResponse(200)))),
			method: "POST", url: "/users/1?q=foo",
			headers: map[string]string{"X-Id": "1", "Cookie": "session=abc", "Content-Type": "application/json", "X-Extra": "ignored"},
			body:    `{"name":"foo"}`,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Empty(t, errs)
			},
		},
		{
			name: "path parameter has wrong type",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users/{id}", openapitest.WithOperation("get",
					openapitest.WithOperationParam("id", true, openapitest.WithParamSchema(schematest.New("integer"))),
					openapitest.WithResponse(200)))),
			method: "GET", url: "/users/abc",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Len(t, errs, 1)
				require.Equal(t, "request.path.id", errs[0].Location)
			},
		},
		{
			name: "path parameter below base path",
			config: openapitest.NewConfig("3.0",
				openapitest.WithServer("/api/v1", ""),
				openapitest.WithPath("/users/{id}", openapitest.WithOperation("get",
					openapitest.WithOperationParam("id", true, openapitest.WithParamSchema(schematest.New("integer"))),
					openapitest.WithResponse(200)))),
			method: "GET", url: "/api/v1/users/abc",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []string{"request.path.id"}, locations(errs))
			},
		},
		{
			name: "path-level parameter is validated",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users/{id}",
					openapitest.WithPathParam("id", openapitest.WithParamSchema(schematest.New("integer"))),
					openapitest.WithOperation("get", openapitest.WithResponse(200)))),
			method: "GET", url: "/users/abc",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []string{"request.path.id"}, locations(errs))
			},
		},
		{
			name: "path parameter is not part of the route",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users", openapitest.WithOperation("get",
					openapitest.WithOperationParam("id", true),
					openapitest.WithResponse(200)))),
			method: "GET", url: "/users",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []openapi.ValidationError{
					{Location: "request.path.id", Message: "path parameter id not found in route /users"},
				}, errs)
			},
		},
		{
			name: "several path parameters in one segment without definition",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/files/{name}-{version}", openapitest.WithOperation("get",
					openapitest.WithResponse(200)))),
			method: "GET", url: "/files/a-1",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Len(t, errs, 1)
				require.Equal(t, "request.path", errs[0].Location)
			},
		},
		{
			name: "required query parameter is missing",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users", openapitest.WithOperation("get",
					openapitest.WithQueryParam("q", true),
					openapitest.WithResponse(200)))),
			method: "GET", url: "/users",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []openapi.ValidationError{{Location: "request.query.q", Message: "parameter is required"}}, errs)
			},
		},
		{
			name: "query parameter has wrong type",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users", openapitest.WithOperation("get",
					openapitest.WithQueryParam("limit", false, openapitest.WithParamSchema(schematest.New("integer"))),
					openapitest.WithResponse(200)))),
			method: "GET", url: "/users?limit=abc",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Len(t, errs, 1)
				require.Equal(t, "request.query.limit", errs[0].Location)
				require.Equal(t, "#/type", errs[0].Schema)
			},
		},
		{
			name: "required header parameter is missing",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users", openapitest.WithOperation("get",
					openapitest.WithHeaderParam("X-Id", true),
					openapitest.WithResponse(200)))),
			method: "GET", url: "/users",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []string{"request.header.X-Id"}, locations(errs))
			},
		},
		{
			name: "required cookie parameter is missing",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users", openapitest.WithOperation("get",
					openapitest.WithCookieParam("session", true),
					openapitest.WithResponse(200)))),
			method: "GET", url: "/users",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []string{"request.cookie.session"}, locations(errs))
			},
		},
		{
			name: "all parameter errors are collected",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users/{id}", openapitest.WithOperation("get",
					openapitest.WithOperationParam("id", true, openapitest.WithParamSchema(schematest.New("integer"))),
					openapitest.WithQueryParam("q", true),
					openapitest.WithHeaderParam("X-Id", true),
					openapitest.WithResponse(200)))),
			method: "GET", url: "/users/abc",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.ElementsMatch(t, []string{"request.path.id", "request.query.q", "request.header.X-Id"}, locations(errs))
			},
		},
		{
			name: "body is required",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users", openapitest.WithOperation("post", jsonBody(user), openapitest.WithResponse(200)))),
			method: "POST", url: "/users",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []openapi.ValidationError{{Location: "request.body", Message: "request body is required"}}, errs)
			},
		},
		{
			name: "body content type is not supported",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users", openapitest.WithOperation("post", jsonBody(user), openapitest.WithResponse(200)))),
			method: "POST", url: "/users",
			headers: map[string]string{"Content-Type": "text/plain"},
			body:    "foo",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Len(t, errs, 1)
				require.Equal(t, "request.body", errs[0].Location)
				require.Contains(t, errs[0].Message, "Unsupported Content-Type: text/plain")
			},
		},
		{
			name: "body violates schema",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users", openapitest.WithOperation("post", jsonBody(user), openapitest.WithResponse(200)))),
			method: "POST", url: "/users",
			headers: map[string]string{"Content-Type": "application/json"},
			body:    `{}`,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []openapi.ValidationError{
					{Location: "request.body", Schema: "#/required", Message: "required properties are missing: name"},
				}, errs)
			},
		},
		{
			name: "body is not valid JSON",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users", openapitest.WithOperation("post", jsonBody(user), openapitest.WithResponse(200)))),
			method: "POST", url: "/users",
			headers: map[string]string{"Content-Type": "application/json"},
			body:    `{`,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []string{"request.body"}, locations(errs))
				require.Empty(t, errs[0].Schema)
			},
		},
		{
			name: "body is ignored without requestBody",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users", openapitest.WithOperation("post", openapitest.WithResponse(200)))),
			method: "POST", url: "/users",
			headers: map[string]string{"Content-Type": "text/plain"},
			body:    "foo",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Empty(t, errs)
			},
		},
		{
			name: "missing security credentials are not an error",
			config: openapitest.NewConfig("3.0",
				openapitest.WithPath("/users", openapitest.WithOperation("get",
					openapitest.WithSecurity(map[string][]string{"apiKey": {}}),
					openapitest.WithResponse(200)))),
			method: "GET", url: "/users",
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Empty(t, errs)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t, validateRequest(t, tc.config, tc.method, tc.url, tc.headers, tc.body))
		})
	}
}

func TestOperationMatch_ValidateResponse(t *testing.T) {
	user := schematest.New("object",
		schematest.WithProperty("name", schematest.New("string")),
		schematest.WithRequired("name"))
	order := schematest.New("object",
		schematest.WithProperty("id", schematest.New("integer")),
		schematest.WithRequired("id"))
	jsonHeaders := map[string]string{"Content-Type": "application/json"}

	getUser := func(method string, opts ...openapitest.OperationOptions) *openapi.Config {
		return openapitest.NewConfig("3.0",
			openapitest.WithPath("/users", openapitest.WithOperation(method, opts...)))
	}

	testcases := []struct {
		name    string
		config  *openapi.Config
		method  string
		status  int
		headers map[string]string
		body    string
		test    func(t *testing.T, errs []openapi.ValidationError)
	}{
		{
			name: "valid response",
			config: getUser("get",
				openapitest.WithResponse(200,
					openapitest.WithResponseHeader("X-Total", "", schematest.New("integer")),
					openapitest.WithContent("application/json", openapitest.WithSchema(user)))),
			status:  200,
			headers: map[string]string{"Content-Type": "application/json; charset=utf-8", "X-Total": "3", "X-Extra": "ignored"},
			body:    `{"name":"foo"}`,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Empty(t, errs)
			},
		},
		{
			name:   "status is not defined",
			config: getUser("get", openapitest.WithResponse(200)),
			status: 404,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []openapi.ValidationError{{Location: "response.status", Message: "no response defined for status code 404"}}, errs)
			},
		},
		{
			name:   "operation without responses",
			config: getUser("get"),
			status: 200,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []string{"response.status"}, locations(errs))
			},
		},
		{
			name:   "default response covers undefined status",
			config: getUser("get", withRangeResponse("default")),
			status: 500,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Empty(t, errs)
			},
		},
		{
			name:   "range response covers status",
			config: getUser("get", withRangeResponse("2XX")),
			status: 201,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Empty(t, errs)
			},
		},
		{
			name: "required header is missing",
			config: getUser("get", openapitest.WithResponse(200,
				withRequiredResponseHeader("X-Total", true, schematest.New("integer")))),
			status: 200,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []openapi.ValidationError{{Location: "response.header.X-Total", Message: "parameter is required"}}, errs)
			},
		},
		{
			name: "header has wrong type",
			config: getUser("get", openapitest.WithResponse(200,
				openapitest.WithResponseHeader("X-Total", "", schematest.New("integer")))),
			status:  200,
			headers: map[string]string{"X-Total": "many"},
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Len(t, errs, 1)
				require.Equal(t, "response.header.X-Total", errs[0].Location)
				require.Equal(t, "#/type", errs[0].Schema)
			},
		},
		{
			name: "header name in other case",
			config: getUser("get", openapitest.WithResponse(200,
				withRequiredResponseHeader("X-Total", true, schematest.New("integer")))),
			status:  200,
			headers: map[string]string{"x-total": "3"},
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Empty(t, errs)
			},
		},
		{
			name: "array header is split",
			config: getUser("get", openapitest.WithResponse(200,
				openapitest.WithResponseHeader("X-Ids", "", schematest.New("array", schematest.WithItems("integer"))))),
			status:  200,
			headers: map[string]string{"X-Ids": "1,2,3"},
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Empty(t, errs)
			},
		},
		{
			name: "Content-Type header defined in the spec is ignored",
			config: getUser("get", openapitest.WithResponse(200,
				withRequiredResponseHeader("Content-Type", true, schematest.New("integer")))),
			status: 200,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Empty(t, errs)
			},
		},
		{
			name: "body violates schema",
			config: getUser("get", openapitest.WithResponse(200,
				openapitest.WithContent("application/json", openapitest.WithSchema(user)))),
			status: 200, headers: jsonHeaders, body: `{}`,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []openapi.ValidationError{
					{Location: "response.body", Schema: "#/required", Message: "required properties are missing: name"},
				}, errs)
			},
		},
		{
			name: "body is not valid JSON",
			config: getUser("get", openapitest.WithResponse(200,
				openapitest.WithContent("application/json", openapitest.WithSchema(user)))),
			status: 200, headers: jsonHeaders, body: `{`,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []string{"response.body"}, locations(errs))
				require.Empty(t, errs[0].Schema)
			},
		},
		{
			name: "content type is not defined for the status",
			config: getUser("get", openapitest.WithResponse(200,
				openapitest.WithContent("application/json", openapitest.WithSchema(user)))),
			status: 200, headers: map[string]string{"Content-Type": "text/html"}, body: `<p>`,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []openapi.ValidationError{
					{Location: "response.body", Message: "content type 'text/html' is not defined for status code 200"},
				}, errs)
			},
		},
		{
			name:   "body without defined content",
			config: getUser("get", openapitest.WithResponse(204)),
			status: 204, headers: jsonHeaders, body: `{}`,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []string{"response.body"}, locations(errs))
			},
		},
		{
			name: "non-JSON body is not validated against the schema",
			config: getUser("get", openapitest.WithResponse(200,
				openapitest.WithContent("text/plain", openapitest.WithSchema(schematest.New("integer"))))),
			status: 200, headers: map[string]string{"Content-Type": "text/plain"}, body: `not a number`,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Empty(t, errs)
			},
		},
		{
			name: "non-JSON body is still checked for its content type",
			config: getUser("get", openapitest.WithResponse(200,
				openapitest.WithContent("text/plain", openapitest.WithSchema(schematest.New("string"))))),
			status: 200, headers: map[string]string{"Content-Type": "text/html"}, body: `x`,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []string{"response.body"}, locations(errs))
			},
		},
		{
			name: "no Content-Type and matching media type",
			config: getUser("get", openapitest.WithResponse(200,
				openapitest.WithContent("application/json", openapitest.WithSchema(user)),
				openapitest.WithContent("application/vnd.api+json", openapitest.WithSchema(order)))),
			status: 200, body: `{"id":1}`,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Empty(t, errs)
			},
		},
		{
			name: "no Content-Type and no matching media type",
			config: getUser("get", openapitest.WithResponse(200,
				openapitest.WithContent("application/json", openapitest.WithSchema(user)),
				openapitest.WithContent("application/vnd.api+json", openapitest.WithSchema(order)))),
			status: 200, body: `{"foo":1}`,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []openapi.ValidationError{
					{Location: "response.body", Message: "no matching content type defined"},
				}, errs)
			},
		},
		{
			name: "no Content-Type and one JSON media type reports its schema errors",
			config: getUser("get", openapitest.WithResponse(200,
				openapitest.WithContent("application/json", openapitest.WithSchema(user)))),
			status: 200, body: `{}`,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []openapi.ValidationError{
					{Location: "response.body", Schema: "#/required", Message: "required properties are missing: name"},
				}, errs)
			},
		},
		{
			name:   "empty body without Content-Type and without defined content",
			config: getUser("get", openapitest.WithResponse(204)),
			status: 204,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Empty(t, errs)
			},
		},
		{
			name: "empty body without Content-Type and with defined content",
			config: getUser("get", openapitest.WithResponse(200,
				openapitest.WithContent("application/json", openapitest.WithSchema(user)))),
			status: 200,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []openapi.ValidationError{
					{Location: "response.body", Message: "response body is required"},
				}, errs)
			},
		},
		{
			name: "HEAD defined is not checked for a body",
			config: getUser("head", openapitest.WithResponse(200,
				withRequiredResponseHeader("X-Total", true, schematest.New("integer")),
				openapitest.WithContent("application/json", openapitest.WithSchema(user)))),
			method:  "HEAD",
			status:  200,
			headers: map[string]string{"X-Total": "3"},
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Empty(t, errs)
			},
		},
		{
			name: "HEAD still checks the headers",
			config: getUser("head", openapitest.WithResponse(200,
				withRequiredResponseHeader("X-Total", true, schematest.New("integer")))),
			method: "HEAD",
			status: 200,
			test: func(t *testing.T, errs []openapi.ValidationError) {
				require.Equal(t, []string{"response.header.X-Total"}, locations(errs))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			method := tc.method
			if method == "" {
				method = "GET"
			}
			match, err := tc.config.ResolveOperation(method, "/users")
			require.NoError(t, err)
			require.NotNil(t, match)
			tc.test(t, match.ValidateResponse(tc.status, tc.headers, tc.body))
		})
	}
}

func TestConfig_ResolveOperation_HeadWithoutHeadOperation(t *testing.T) {
	config := openapitest.NewConfig("3.0",
		openapitest.WithPath("/users", getOperation("get")))
	match, err := config.ResolveOperation("HEAD", "/users")
	require.NoError(t, err)
	require.Nil(t, match)
}

func withRangeResponse(status string) openapitest.OperationOptions {
	return func(o *openapi.Operation) {
		o.Responses.Set(status, &openapi.ResponseRef{Value: openapitest.NewResponse()})
	}
}

func withRequiredResponseHeader(name string, required bool, s *schema.Schema) openapitest.ResponseOptions {
	return openapitest.UseResponseHeaderRef(name, &openapi.HeaderRef{Value: &openapi.Header{
		Parameter: openapi.Parameter{Name: name, Required: &required, Schema: s},
	}})
}
