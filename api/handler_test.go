package api

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_ServeHTTP(t *testing.T) {
	testcases := []struct {
		name string
		fn   func(t *testing.T, h http.Handler)
	}{
		{
			name: "should 405 when POST to file server",
			fn: func(t *testing.T, h http.Handler) {
				r := httptest.NewRequest(http.MethodPost, "http://foo.api", nil)
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, r)
				require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
			},
		},
		{
			name: "cors is set",
			fn: func(t *testing.T, h http.Handler) {
				r := httptest.NewRequest(http.MethodGet, "http://foo.api/api/info", nil)
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
			},
		},
		{
			name: "/api/info",
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/info",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"version":""}`))
			},
		},
		{
			name: "/api/services/http",
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/info",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"version":""}`))
			},
		},
		{
			name: "openapi path should return index.html",
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/dashboard/http/services/petstore/paths/%2Fpets%2F%7BpetId%7D",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "text/html; charset=utf-8"))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := New(runtime.New(), static.Api{Dashboard: true})
			tc.fn(t, h)
		})
	}
}

func TestHandler_ApiPath_ServeHTTP(t *testing.T) {
	testcases := []struct {
		name string
		path string
		fn   func(t *testing.T, h http.Handler)
	}{
		{
			name: "request api info",
			path: "/mokapi",
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
		},
		{
			name: "request web app",
			path: "/mokapi",
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/index.html",
					nil,
					"",
					h,
					try.HasStatusCode(200))
			},
		},
		{
			name: "request web app",
			path: "/mokapi/dashboard",
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/mokapi/dashboard/index.html",
					nil,
					"",
					h,
					try.HasStatusCode(200), try.BodyContains(`<base href="/mokapi/dashboard/" />`))
			},
		},
	}

	t.Parallel()
	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			h := New(runtime.New(), static.Api{Path: test.path})
			hh := h.(*handler)
			hh.fileServer = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path == "/index.html" {
					writer.WriteHeader(200)
					return
				}
				writer.WriteHeader(404)
			})
			test.fn(t, h)
		})
	}
}
