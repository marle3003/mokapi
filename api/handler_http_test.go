package api

import (
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/config/dynamic/openapi/schema/schematest"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Http(t *testing.T) {
	testcases := []struct {
		name         string
		app          *runtime.App
		requestUrl   string
		responseBody string
	}{
		{
			name: "get http services",
			app: &runtime.App{
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0", openapitest.WithInfo("foo", "1.0", "bar")),
					},
				},
			},
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"foo","description":"bar","version":"1.0","type":"http"}]`,
		},
		{
			name: "get http services with contact",
			app: &runtime.App{
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0",
							openapitest.WithInfo("foo", "", ""),
							openapitest.WithContact("foo", "https://foo.bar", "foo@bar.com")),
					},
				},
			},
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"foo","contact":{"name":"foo","url":"https://foo.bar","email":"foo@bar.com"},"type":"http"}]`,
		},
		{
			name: "get specific http service",
			app: &runtime.App{
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0", openapitest.WithInfo("foo", "1.0", "bar")),
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/http/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0"}`,
		},
		{
			name: "get http service info",
			app: &runtime.App{
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0",
							openapitest.WithInfo("foo", "1.0", "bar"),
						),
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/http/foo",
			responseBody: `{"name":"foo","description":"bar","version":"1.0"}`,
		},
		{
			name: "get http service contact",
			app: &runtime.App{
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0",
							openapitest.WithContact("foo", "http://foo.bar", "foo@bar.com"),
						),
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/http/foo",
			responseBody: `{"name":"","contact":{"name":"foo","url":"http://foo.bar","email":"foo@bar.com"}}`,
		},
		{
			name: "get http service server",
			app: &runtime.App{
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0",
							openapitest.WithServer("https://foo.bar", "a foo description"),
						),
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/http/foo",
			responseBody: `{"name":"","servers":[{"url":"https://foo.bar","description":"a foo description"}]}`,
		},
		{
			name: "get http service with parameters",
			app: &runtime.App{
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0",
							openapitest.WithPath("/foo/{bar}", openapitest.NewPath(
								openapitest.WithPathParam("bar", "path", true, openapitest.WithParamSchema(schematest.New("string"))),
								openapitest.WithOperation("get", openapitest.NewOperation()),
							))),
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/http/foo",
			responseBody: `{"name":"","paths":[{"path":"/foo/{bar}","operations":[{"method":"get","deprecated":false,"parameters":[{"name":"bar","type":"path","required":true,"deprecated":false,"exploded":false,"schema":{"type":"string"}}]}]}]}`,
		},
		{
			name: "get http service with requestBody",
			app: &runtime.App{
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0",
							openapitest.WithPath("/foo/{bar}", openapitest.NewPath(
								openapitest.WithOperation("get", openapitest.NewOperation(
									openapitest.WithRequestBody("foo", true,
										openapitest.WithRequestContent("application/json", openapitest.WithSchema(schematest.New("string"))),
									),
								)),
							))),
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/http/foo",
			responseBody: `{"name":"","paths":[{"path":"/foo/{bar}","operations":[{"method":"get","deprecated":false,"requestBody":{"description":"foo","contents":[{"type":"application/json","schema":{"type":"string"}}],"required":true}}]}]}`,
		},
		{
			name: "get http service with response",
			app: &runtime.App{
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0",
							openapitest.WithPath("/foo/{bar}", openapitest.NewPath(
								openapitest.WithOperation("get", openapitest.NewOperation(
									openapitest.WithResponse(http.StatusOK,
										openapitest.WithResponseDescription("foo description"),
										openapitest.WithContent(
											"application/json",
											openapitest.NewContent(
												openapitest.WithSchema(schematest.New("string")),
											),
										),
										openapitest.WithResponseHeader("foo", "bar", schematest.New("string")),
									),
								)),
							))),
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/http/foo",
			responseBody: `{"name":"","paths":[{"path":"/foo/{bar}","operations":[{"method":"get","deprecated":false,"responses":[{"statusCode":200,"description":"foo description","contents":[{"type":"application/json","schema":{"type":"string"}}],"headers":[{"name":"foo","description":"bar","schema":{"type":"string"}}]}]}]}]}`,
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app, static.Api{})

			try.Handler(t,
				http.MethodGet,
				tc.requestUrl,
				nil,
				"",
				h,
				try.HasStatusCode(200),
				try.HasHeader("Content-Type", "application/json"),
				try.HasBody(tc.responseBody))
		})
	}
}

func TestHandler_Http_NotFound(t *testing.T) {
	h := New(runtime.New(), static.Api{})

	try.Handler(t,
		http.MethodGet,
		"http://foo.api/api/services/http/foo",
		nil,
		"",
		h,
		try.HasStatusCode(404))
}

func TestHandler_Http_Metrics(t *testing.T) {
	testcases := []struct {
		name         string
		app          *runtime.App
		requestUrl   string
		responseBody string
		addMetrics   func(monitor *monitor.Monitor)
	}{
		{
			name: "service list with metric",
			app: &runtime.App{
				Monitor: monitor.New(),
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0", openapitest.WithInfo("foo", "", "")),
					},
				},
			},
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"foo","type":"http","metrics":[{"name":"http_requests_total{service=\"foo\",endpoint=\"bar\"}","value":1}]}]`,
			addMetrics: func(monitor *monitor.Monitor) {
				monitor.Http.RequestCounter.WithLabel("foo", "bar").Add(1)
			},
		},
		{
			name: "specific with metric",
			app: &runtime.App{
				Monitor: monitor.New(),
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0", openapitest.WithInfo("foo", "", "")),
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/http/foo",
			responseBody: `{"name":"foo","metrics":[{"name":"http_requests_total{service=\"foo\",endpoint=\"bar\"}","value":1}]}`,
			addMetrics: func(monitor *monitor.Monitor) {
				monitor.Http.RequestCounter.WithLabel("foo", "bar").Add(1)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app, static.Api{})
			tc.addMetrics(tc.app.Monitor)

			try.Handler(t,
				http.MethodGet,
				tc.requestUrl,
				nil,
				"",
				h,
				try.HasStatusCode(200),
				try.HasHeader("Content-Type", "application/json"),
				try.HasBody(tc.responseBody))
		})
	}
}
