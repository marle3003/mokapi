package health_test

import (
	"mokapi/config/static"
	"mokapi/health"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	testcases := []struct {
		name string
		cfg  static.Health
		test func(t *testing.T, h http.Handler, hook *test.Hook)
	}{
		{
			name: "Health OK",
			cfg:  static.Health{},
			test: func(t *testing.T, h http.Handler, hook *test.Hook) {
				r := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080/health", nil)
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `{"status":"healthy"}`, rr.Body.String())
			},
		},
		{
			name: "empty path invalid request URL",
			cfg:  static.Health{},
			test: func(t *testing.T, h http.Handler, hook *test.Hook) {
				r := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080", nil)
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, r)
				require.Equal(t, http.StatusNotFound, rr.Code)
			},
		},
		{
			name: "set path",
			cfg:  static.Health{Path: "/health/live"},
			test: func(t *testing.T, h http.Handler, hook *test.Hook) {
				r := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080/health/live", nil)
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, r)
				require.Equal(t, http.StatusOK, rr.Code)
				require.Equal(t, `{"status":"healthy"}`, rr.Body.String())
			},
		},
		{
			name: "404",
			cfg:  static.Health{},
			test: func(t *testing.T, h http.Handler, hook *test.Hook) {
				r := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080/foo", nil)
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, r)
				require.Equal(t, http.StatusNotFound, rr.Code)
			},
		},
		{
			name: "404 with logging",
			cfg:  static.Health{Log: true},
			test: func(t *testing.T, h http.Handler, hook *test.Hook) {
				r := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080/foo", nil)
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, r)
				require.Len(t, hook.Entries, 1)
				require.Equal(t, logrus.DebugLevel, hook.LastEntry().Level)
				require.Equal(t, "healthcheck: not found: GET http://127.0.0.1:8080/foo", hook.LastEntry().Message)
			},
		},
		{
			name: "method not allowed with logging",
			cfg:  static.Health{Log: true},
			test: func(t *testing.T, h http.Handler, hook *test.Hook) {
				r := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/foo", nil)
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, r)
				require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
				require.Equal(t, http.MethodGet, rr.Header().Get("Allow"))
				require.Len(t, hook.Entries, 1)
				require.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
				require.Equal(t, "healthcheck: method not allowed: POST http://127.0.0.1:8080/foo", hook.LastEntry().Message)
			},
		},
		{
			name: "healthy with logging",
			cfg:  static.Health{Log: true},
			test: func(t *testing.T, h http.Handler, hook *test.Hook) {
				r := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080/health", nil)
				rr := httptest.NewRecorder()
				h.ServeHTTP(rr, r)
				require.Len(t, hook.Entries, 1)
				require.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
				require.Equal(t, "healthcheck: GET http://127.0.0.1:8080/health: healthy", hook.LastEntry().Message)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			logrus.SetLevel(logrus.DebugLevel)
			hook := test.NewGlobal()

			h := health.New(tc.cfg)
			tc.test(t, h, hook)
		})
	}
}
