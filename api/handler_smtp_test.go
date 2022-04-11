package api

import (
	"mokapi/config/dynamic/smtp"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Smtp(t *testing.T) {
	testcases := []struct {
		name string
		app  *runtime.App
		f    func(t *testing.T, h http.Handler)
	}{
		{
			name: "/api/services/smtp",
			app: &runtime.App{
				Smtp: map[string]*runtime.SmtpInfo{
					"foo": {
						&smtp.Config{Name: "foo"},
					},
				},
			},
			f: func(t *testing.T, h http.Handler) {
				try.Handler(t,
					http.MethodGet,
					"http://foo.api/api/services/smtp/foo",
					nil,
					"",
					h,
					try.HasStatusCode(200),
					try.HasHeader("Content-Type", "application/json"),
					try.HasBody(`{"name":"foo","server":""}`))
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := New(tc.app, static.Api{})
			tc.f(t, h)
		})
	}
}
