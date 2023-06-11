package api

import (
	"mokapi/config/dynamic/directory"
	"mokapi/config/static"
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"testing"
)

func TestHandler_Ldap(t *testing.T) {
	testcases := []struct {
		name         string
		app          *runtime.App
		requestUrl   string
		contentType  string
		responseBody string
	}{
		{
			name: "get ldap services",
			app: &runtime.App{
				Ldap: map[string]*runtime.LdapInfo{
					"foo": {
						Config: &directory.Config{Info: directory.Info{Name: "foo", Description: "bar", Version: "1.0"}},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services",
			contentType:  "application/json",
			responseBody: `[{"name":"foo","description":"bar","version":"1.0","type":"ldap"}]`,
		},
		{
			name: "get ldap service",
			app: &runtime.App{
				Ldap: map[string]*runtime.LdapInfo{
					"foo": {
						Config: &directory.Config{
							Info:    directory.Info{Name: "foo", Description: "bar", Version: "1.0"},
							Address: "0.0.0.0:389",
						},
					},
				},
			},
			requestUrl:   "http://foo.api/api/services/ldap/foo",
			contentType:  "application/json",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","server":"0.0.0.0:389"}`,
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
				try.HasHeader("Content-Type", tc.contentType),
				try.HasBody(tc.responseBody))
		})
	}
}
