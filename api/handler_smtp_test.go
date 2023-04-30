package api

import (
	"mokapi/config/dynamic/mail"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Smtp(t *testing.T) {
	testcases := []struct {
		name         string
		app          *runtime.App
		requestUrl   string
		responseBody string
	}{
		{
			name: "get smtp services",
			app: &runtime.App{
				Smtp: map[string]*runtime.SmtpInfo{
					"foo": {
						&mail.Config{Info: mail.Info{Name: "foo", Description: "bar", Version: "1.0"}},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services",
			responseBody: `[{"name":"foo","description":"bar","version":"1.0","type":"smtp"}]`,
		},
		{
			name: "/api/services/smtp",
			app: &runtime.App{
				Smtp: map[string]*runtime.SmtpInfo{
					"foo": {
						&mail.Config{Info: mail.Info{Name: "foo"}},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/smtp/foo",
			responseBody: `{"name":"foo","server":""}`,
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
