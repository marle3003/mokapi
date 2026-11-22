package api

import (
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/directory"
	"mokapi/runtime"
	"mokapi/runtime/runtimetest"
	"mokapi/try"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
				return runtimetest.NewApp(runtimetest.WithLdap(
					&directory.Config{Info: directory.Info{Name: "foo", Description: "bar", Version: "1.0"}},
				))
			},
			requestUrl:   "http://foo.api/api/services",
			contentType:  "application/json",
			responseBody: `[{"name":"foo","description":"bar","version":"1.0","type":"ldap","metrics":{"requests_total":0,"request_timestamp":0}}]`,
		},
		{
			name: "get ldap service",
			app: func() *runtime.App {
				app := runtime.New(&static.Config{}, &dynamictest.Reader{})
				cfg := dynamictest.NewConfig(&directory.Config{
					Info:    directory.Info{Name: "foo", Description: "bar", Version: "1.0"},
					Address: "0.0.0.0:389",
				}, dynamictest.WithInfo(dynamictest.NewConfigInfo()))
				cfg.Info.Time = mustTime("2023-12-27T13:01:30+00:00")
				app.Ldap.Add(cfg, enginetest.NewEngine())
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

func TestHandler_Ldap_ServiceHasNoItems(t *testing.T) {
	app := runtimetest.NewApp(runtimetest.WithLdap(
		&directory.Config{Info: directory.Info{Name: "foo", Version: "1.0"}},
	))

	try.Handler(t,
		http.MethodGet,
		"http://foo.api/api/services",
		nil,
		"",
		New(app, static.Api{}),
		try.HasStatusCode(200),
		try.HasBody(`[{"name":"foo","version":"1.0","type":"ldap","metrics":{"requests_total":0,"request_timestamp":0}}]`))
}

// An info without a config snapshot is what a request sees when the last file of a directory is
// removed while the request is running.
func TestHandler_Ldap_ConfigRemoved(t *testing.T) {
	newApp := func() *runtime.App {
		return runtimetest.NewApp(
			runtimetest.WithLdap(&directory.Config{Info: directory.Info{Name: "bar", Version: "1.0"}}),
			runtimetest.WithLdapInfo("foo", &runtime.LdapInfo{}),
		)
	}

	t.Run("service list skips it", func(t *testing.T) {
		try.Handler(t,
			http.MethodGet,
			"http://foo.api/api/services",
			nil,
			"",
			New(newApp(), static.Api{}),
			try.HasStatusCode(200),
			try.HasBody(`[{"name":"bar","version":"1.0","type":"ldap","metrics":{"requests_total":0,"request_timestamp":0}}]`))
	})

	t.Run("service returns 404", func(t *testing.T) {
		try.Handler(t,
			http.MethodGet,
			"http://foo.api/api/services/ldap/foo",
			nil,
			"",
			New(newApp(), static.Api{}),
			try.HasStatusCode(404))
	})

	t.Run("service returns 404 after the file was removed from the store", func(t *testing.T) {
		app := runtimetest.NewApp()
		file := dynamictest.NewConfig(&directory.Config{Info: directory.Info{Name: "foo"}}, dynamictest.WithInfo(dynamictest.NewConfigInfo()))
		info := app.Ldap.Add(file, enginetest.NewEngine())
		app.Ldap.Remove(file)

		require.Nil(t, info.Config())
		try.Handler(t,
			http.MethodGet,
			"http://foo.api/api/services/ldap/foo",
			nil,
			"",
			New(app, static.Api{}),
			try.HasStatusCode(404))
	})
}
