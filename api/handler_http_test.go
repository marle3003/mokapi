package api

import (
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/runtime"
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
						openapitest.NewConfig("3.0.0"),
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
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app, true)
			tc.fn(t, h)
		})
	}
}
