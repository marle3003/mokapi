package api

import (
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/runtime/logs"
	"mokapi/runtime/monitor"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Http(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		fn   func(t *testing.T, h http.Handler)
	}{
		{
			name: "/api/services/http",
			app: &runtime.App{
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0", openapitest.WithInfo("foo", "", "")),
					},
				},
				Monitor: monitor.New(),
			},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/services/http",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"name":"foo","lastRequest":0,"requests":0,"errors":0}]`))
			},
		},
		{
			name: "/api/services/http/foo",
			app: &runtime.App{
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0"),
					},
				},
			},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/services/http/foo",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"openapi":"3.0.0","info":{"title":"","version":""},"paths":{},"components":{}}`))
			},
		},
		{
			name: "/api/services/http/foo",
			app: &runtime.App{
				Monitor: &monitor.Monitor{Http: &monitor.Http{
					Log: []*logs.HttpLog{
						{
							Id:       "1",
							Service:  "foo",
							Time:     100,
							Duration: 200,
							Request: &logs.HttpRequestLog{
								Method: "GET",
								Url:    "foo.bar",
							},
							Response: &logs.HttpResponseLog{
								Headers: map[string]string{
									"foo": "bar",
								},
								StatusCode: http.StatusCreated,
								Body:       "foobar",
							},
						},
					},
				}},
			},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/http/requests",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"id":"1","service":"foo","time":100,"duration":200,"request":{"method":"GET","url":"foo.bar"},"response":{"statusCode":201,"headers":{"foo":"bar"},"body":"foobar"}}]`))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app, static.Api{})
			tc.fn(t, h)
		})
	}
}
