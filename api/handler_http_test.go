package api

import (
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/config/static"
	"mokapi/runtime"
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
			name: "/api/services",
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
					"http://foo.api/api/services",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`[{"name":"foo","description":"","version":"","type":"http","metrics":[]}]`))
			},
		},
		{
			name: "/api/services/http/foo",
			app: &runtime.App{
				Http: map[string]*runtime.HttpInfo{
					"foo": {
						Config: openapitest.NewConfig("3.0.0", openapitest.WithInfo("foo", "1.0", "bar")),
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
					try.HasBody(`{"name":"foo","description":"bar","version":"1.0"}`))
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
