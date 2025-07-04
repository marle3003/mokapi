package api_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/api"
	"mokapi/config/static"
	"mokapi/providers/openapi"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/try"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_ServeHTTP(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, h http.Handler)
	}{
		{
			name: "PATCH is not allowed",
			test: func(t *testing.T, h http.Handler) {
				r := httptest.NewRequest(http.MethodPatch, "http://foo.api", nil)
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, r)
				require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
			},
		},
		{
			name: "should 405 when POST to file server",
			test: func(t *testing.T, h http.Handler) {
				r := httptest.NewRequest(http.MethodPost, "http://foo.api", nil)
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, r)
				require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
			},
		},
		{
			name: "cors is set",
			test: func(t *testing.T, h http.Handler) {
				r := httptest.NewRequest(http.MethodGet, "http://foo.api/api/info", nil)
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
			},
		},
		{
			name: "info",
			test: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/info",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"version":"","buildTime":"","search":{"enabled":false}}`))
			},
		},
		{
			name: "openapi path should return index.html",
			test: func(t *testing.T, h http.Handler) {
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

			h := api.New(runtime.New(&static.Config{}), static.Api{Dashboard: true})
			tc.test(t, h)
		})
	}
}

func TestHandler_Api_Info(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		fn   func(t *testing.T, h http.Handler)
	}{
		{
			name: "version 1.0",
			app:  &runtime.App{Version: "1.0", BuildTime: "2025-01-04T23:20:50.52Z"},
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/info",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"version":"1.0","buildTime":"2025-01-04T23:20:50.52Z","search":{"enabled":false}}`))
			},
		},
		{
			name: "http active",
			app:  runtimetest.NewHttpApp(&openapi.Config{}),
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/info",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"version":"","buildTime":"","activeServices":["http"],"search":{"enabled":false}}`))
			},
		},
		{
			name: "kafka active",
			app:  runtimetest.NewApp(runtimetest.WithKafkaInfo("foo", &runtime.KafkaInfo{})),
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/info",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"version":"","buildTime":"","activeServices":["kafka"],"search":{"enabled":false}}`))
			},
		},
		{
			name: "smtp active",
			app:  runtimetest.NewApp(runtimetest.WithMailInfo("foo", &runtime.MailInfo{})),
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/info",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"version":"","buildTime":"","activeServices":["smtp"],"search":{"enabled":false}}`))
			},
		},
		{
			name: "ldap active",
			app:  runtimetest.NewApp(runtimetest.WithLdapInfo("foo", &runtime.LdapInfo{})),
			fn: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/info",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"version":"","buildTime":"","activeServices":["ldap"],"search":{"enabled":false}}`))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := api.New(tc.app, static.Api{})
			tc.fn(t, h)
		})
	}
}

func TestHandler_NoDashboard(t *testing.T) {
	h := api.New(runtime.New(&static.Config{}), static.Api{Dashboard: false})
	try.Handler(t,
		http.MethodGet,
		"http://foo.api",
		nil,
		"",
		h,
		try.HasStatusCode(404),
		try.HasHeader("Content-Type", "text/plain; charset=utf-8"),
		try.HasBody("not found\n"))
}

func TestHandler_SearchEnabled(t *testing.T) {
	h := api.New(runtime.New(&static.Config{}), static.Api{Dashboard: true, Search: static.Search{Enabled: true}})
	try.Handler(t,
		http.MethodGet,
		"http://foo.api/api/info",
		nil,
		"",
		h,
		try.HasStatusCode(200),
		try.HasHeader("Content-Type", "application/json"),
		try.HasBody(`{"version":"","buildTime":"","search":{"enabled":true}}`))
}
