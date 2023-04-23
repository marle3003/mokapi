package api

import (
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_FileServer(t *testing.T) {
	testcases := []struct {
		name       string
		config     static.Api
		fn         func(t *testing.T, h http.Handler)
		fileServer http.Handler
	}{
		{
			name:   "request api info",
			config: static.Api{Path: "/mokapi"},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/api/info",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"version":""}`))
			},
			fileServer: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/index.html" {
					writer.WriteHeader(404)
				}
			}),
		},
		{
			name:   "request web app",
			config: static.Api{Path: "/mokapi"},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/index.html",
					nil,
					"",
					h,
					try.HasStatusCode(200))
			},
			fileServer: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/index.html" {
					writer.WriteHeader(404)
				}
			}),
		},
		{
			name:   "request web app",
			config: static.Api{Path: "/mokapi/dashboard"},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/index.html",
					nil,
					"",
					h,
					try.HasStatusCode(200), try.BodyContains(`<base href="/mokapi/dashboard/" />`))
			},
			fileServer: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/index.html" {
					writer.WriteHeader(404)
				}
			}),
		},
		{
			name:   "request asset",
			config: static.Api{Path: "/mokapi/dashboard"},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/assets/index.js",
					nil,
					"",
					h,
					try.HasStatusCode(200))
			},
			fileServer: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/assets/index.js" {
					writer.WriteHeader(404)
				}
			}),
		},
		/*{
			name:   "request svg",
			config: static.Api{Path: "/mokapi/dashboard"},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/foo/logo.svg",
					nil,
					"",
					h,
					try.HasStatusCode(200))
			},
			fileServer: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/logo.svg" {
					writer.WriteHeader(404)
				}
			}),
		},*/
		{
			name:   "url rewrite (proxy)",
			config: static.Api{Path: "/mokapi/dashboard", Base: "/foo/mokapi/dashboard"},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/foo/mokapi/dashboard/index.html",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.BodyContains(`<base href="/foo/mokapi/dashboard/" />`))
			},
			fileServer: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/index.html" {
					writer.WriteHeader(404)
				}
			}),
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := New(runtime.New(), tc.config)
			hh := h.(*handler)
			hh.fileServer = tc.fileServer
			tc.fn(t, h)
		})
	}
}
