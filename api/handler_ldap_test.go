package api

import (
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/directory"
	"mokapi/runtime"
	"mokapi/try"
	"net/http"
	"testing"
	"time"
)

func TestHandler_Ldap(t *testing.T) {
	mustTime := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		return t
	}

	testcases := []struct {
		name         string
		app          func() *runtime.App
		requestUrl   string
		contentType  string
		responseBody string
	}{
		{
			name: "get ldap services",
			app: func() *runtime.App {
				return &runtime.App{
					Ldap: map[string]*runtime.LdapInfo{
						"foo": {
							Config: &directory.Config{Info: directory.Info{Name: "foo", Description: "bar", Version: "1.0"}},
						},
					},
				}
			},
			requestUrl:   "http://foo.api/api/services",
			contentType:  "application/json",
			responseBody: `[{"name":"foo","description":"bar","version":"1.0","type":"ldap"}]`,
		},
		{
			name: "get ldap service",
			app: func() *runtime.App {
				app := runtime.New()
				cfg := &dynamic.Config{
					Info: dynamictest.NewConfigInfo(),
					Data: &directory.Config{
						Info:    directory.Info{Name: "foo", Description: "bar", Version: "1.0"},
						Address: "0.0.0.0:389",
					},
				}
				cfg.Info.Time = mustTime("2023-12-27T13:01:30+00:00")
				app.AddLdap(cfg, enginetest.NewEngine())
				return app
			},
			requestUrl:   "http://foo.api/api/services/ldap/foo",
			contentType:  "application/json",
			responseBody: `{"name":"foo","description":"bar","version":"1.0","server":"0.0.0.0:389","configs":[{"id":"64613435-3062-6462-3033-316532633233","url":"file://foo.yml","provider":"test","time":"2023-12-27T13:01:30Z"}]}`,
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := New(tc.app(), static.Api{})

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
