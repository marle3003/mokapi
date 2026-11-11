package api_test

import (
	"mokapi/api"
	"mokapi/config/static"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/providers/openapi/schema/schematest"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/metrics"
	"mokapi/runtime/runtimetest"
	"mokapi/try"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func newProfileApi(title string, required ...string) *openapi.Config {
	return openapitest.NewConfig("3.0",
		openapitest.WithInfo(title, "1.0", ""),
		openapitest.WithServer("/api/v1", ""),
		openapitest.WithPath("/profile", openapitest.WithOperation("get",
			openapitest.WithQueryParam("id", true, openapitest.WithParamSchema(schematest.New("integer"))),
			openapitest.WithResponse(200,
				openapitest.WithContent("application/json", openapitest.WithSchema(
					schematest.New("object",
						schematest.WithProperty("name", schematest.New("string")),
						schematest.WithProperty("email", schematest.New("string")),
						schematest.WithRequired(required...))))))))
}

const validateUrl = "http://foo.api/api/services/http/validate"

func TestHandler_Http_Validate(t *testing.T) {
	jsonHeader := map[string]string{"Content-Type": "application/json"}

	testcases := []struct {
		name   string
		app    func() *runtime.App
		method string
		body   string
		status int
		result string
	}{
		{
			name: "valid pair",
			app: func() *runtime.App {
				return runtimetest.NewHttpApp(newProfileApi("Profile API", "name"))
			},
			body: `{
				"request": { "method": "get", "url": "/api/v1/profile?id=1" },
				"response": { "statusCode": 200, "headers": { "Content-Type": "application/json" }, "body": "{\"name\":\"Jane\"}" }
			}`,
			status: 200,
			result: `{"valid":true,"operation":{"api":"Profile API","method":"GET","path":"/profile"},"errors":[]}`,
		},
		{
			name: "invalid pair",
			app: func() *runtime.App {
				return runtimetest.NewHttpApp(newProfileApi("Profile API", "name", "email"))
			},
			body: `{
				"request": { "method": "GET", "url": "/api/v1/profile" },
				"response": { "statusCode": 200, "headers": { "Content-Type": "application/json" }, "body": "{\"name\":\"Jane\"}" }
			}`,
			status: 200,
			result: `{"valid":false,"operation":{"api":"Profile API","method":"GET","path":"/profile"},"errors":[
				{"location":"request.query.id","message":"parameter is required"},
				{"location":"response.body","schema":"#/required","message":"required properties are missing: email"}
			]}`,
		},
		{
			name: "no operation",
			app: func() *runtime.App {
				return runtimetest.NewHttpApp(newProfileApi("Profile API"))
			},
			body:   `{"request":{"method":"GET","url":"/api/v1/unknown"},"response":{"statusCode":200}}`,
			status: 200,
			result: `{"valid":false,"errors":[{"location":"operation","message":"no operation found for GET /api/v1/unknown"}]}`,
		},
		{
			name: "no http api",
			app: func() *runtime.App {
				return runtimetest.NewHttpApp()
			},
			body:   `{"request":{"method":"GET","url":"/api/v1/profile"},"response":{"statusCode":200}}`,
			status: 200,
			result: `{"valid":false,"errors":[{"location":"operation","message":"no operation found for GET /api/v1/profile"}]}`,
		},
		{
			name: "operation is found in the second api",
			app: func() *runtime.App {
				other := openapitest.NewConfig("3.0", openapitest.WithInfo("Other API", "1.0", ""),
					openapitest.WithPath("/other", openapitest.WithOperation("get", openapitest.WithResponse(200))))
				return runtimetest.NewHttpApp(newProfileApi("Profile API"), other)
			},
			body:   `{"request":{"method":"GET","url":"/other"},"response":{"statusCode":200}}`,
			status: 200,
			result: `{"valid":true,"operation":{"api":"Other API","method":"GET","path":"/other"},"errors":[]}`,
		},
		{
			name: "ambiguous without api",
			app: func() *runtime.App {
				return runtimetest.NewHttpApp(newProfileApi("Profile API v2"), newProfileApi("Profile API"))
			},
			body:   `{"request":{"method":"GET","url":"/api/v1/profile?id=1"},"response":{"statusCode":200}}`,
			status: 200,
			result: `{"valid":false,"errors":[{"location":"operation","message":"ambiguous operation: GET /api/v1/profile?id=1 matches the APIs 'Profile API', 'Profile API v2'. Set 'api' to select one."}]}`,
		},
		{
			name: "api selects one of two matching apis",
			app: func() *runtime.App {
				return runtimetest.NewHttpApp(newProfileApi("Profile API v2"), newProfileApi("Profile API"))
			},
			body:   `{"api":"Profile API v2","request":{"method":"GET","url":"/api/v1/profile?id=1"},"response":{"statusCode":200,"body":"{}"}}`,
			status: 200,
			result: `{"valid":true,"operation":{"api":"Profile API v2","method":"GET","path":"/profile"},"errors":[]}`,
		},
		{
			name: "api does not contain the operation",
			app: func() *runtime.App {
				other := openapitest.NewConfig("3.0", openapitest.WithInfo("Other API", "1.0", ""))
				return runtimetest.NewHttpApp(newProfileApi("Profile API"), other)
			},
			body:   `{"api":"Other API","request":{"method":"GET","url":"/api/v1/profile"},"response":{"statusCode":200}}`,
			status: 200,
			result: `{"valid":false,"errors":[{"location":"operation","message":"no operation found for GET /api/v1/profile"}]}`,
		},
		{
			name: "unknown api",
			app: func() *runtime.App {
				return runtimetest.NewHttpApp(newProfileApi("Profile API"))
			},
			body:   `{"api":"foo","request":{"method":"GET","url":"/api/v1/profile"},"response":{"statusCode":200}}`,
			status: 404,
			result: `{"message":"api 'foo' not found"}`,
		},
		{
			name:   "invalid JSON",
			app:    func() *runtime.App { return runtimetest.NewHttpApp() },
			body:   `{`,
			status: 400,
			result: `{"message":"invalid JSON: unexpected EOF"}`,
		},
		{
			name:   "missing request",
			app:    func() *runtime.App { return runtimetest.NewHttpApp() },
			body:   `{"response":{"statusCode":200}}`,
			status: 400,
			result: `{"message":"request is required"}`,
		},
		{
			name:   "missing response",
			app:    func() *runtime.App { return runtimetest.NewHttpApp() },
			body:   `{"request":{"method":"GET","url":"/"}}`,
			status: 400,
			result: `{"message":"response is required"}`,
		},
		{
			name:   "missing method",
			app:    func() *runtime.App { return runtimetest.NewHttpApp() },
			body:   `{"request":{"url":"/"},"response":{"statusCode":200}}`,
			status: 400,
			result: `{"message":"request.method is required"}`,
		},
		{
			name:   "missing url",
			app:    func() *runtime.App { return runtimetest.NewHttpApp() },
			body:   `{"request":{"method":"GET"},"response":{"statusCode":200}}`,
			status: 400,
			result: `{"message":"request.url is required"}`,
		},
		{
			name:   "missing statusCode",
			app:    func() *runtime.App { return runtimetest.NewHttpApp() },
			body:   `{"request":{"method":"GET","url":"/"},"response":{}}`,
			status: 400,
			result: `{"message":"response.statusCode is required"}`,
		},
		{
			name:   "invalid url",
			app:    func() *runtime.App { return runtimetest.NewHttpApp(newProfileApi("Profile API")) },
			body:   `{"request":{"method":"GET","url":"http://[::1"},"response":{"statusCode":200}}`,
			status: 400,
			result: `{"message":"invalid request.url: invalid url: parse \"http://[::1\": missing ']' in host"}`,
		},
		{
			name:   "GET is not allowed",
			app:    func() *runtime.App { return runtimetest.NewHttpApp() },
			method: http.MethodGet,
			status: 405,
			result: `{"message":"method GET is not allowed"}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			method := tc.method
			if method == "" {
				method = http.MethodPost
			}
			h := api.New(tc.app(), static.Api{})
			try.Handler(t, method, validateUrl, jsonHeader, tc.body, h,
				try.HasStatusCode(tc.status),
				try.HasHeader("Content-Type", "application/json"),
				try.AssertBody(func(t *testing.T, body string) {
					require.JSONEq(t, tc.result, body)
				}))
		})
	}
}

func TestHandler_Http_Validate_HasNoSideEffects(t *testing.T) {
	app := runtimetest.NewHttpApp(newProfileApi("Profile API"))
	h := api.New(app, static.Api{})

	try.Handler(t, http.MethodPost, validateUrl, nil,
		`{"request":{"method":"GET","url":"/api/v1/profile?id=1"},"response":{"statusCode":200}}`, h,
		try.HasStatusCode(200))

	require.Empty(t, app.Events.GetEvents(events.NewTraits()))
	require.Equal(t, 0.0, app.Monitor.Http.RequestCounter.Sum(metrics.NewQuery()))
	require.Equal(t, 0.0, app.Monitor.Http.RequestErrorCounter.Sum(metrics.NewQuery()))
}
