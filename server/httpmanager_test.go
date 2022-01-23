package server

import (
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/test"
	"mokapi/try"
	"net/url"
	"testing"
)

func TestHttpServers_Monitor(t *testing.T) {
	test.NewNullLogger()
	store, err := cert.NewStore(&static.Config{})
	require.NoError(t, err)

	servers := HttpServers{}
	defer servers.Stop()

	app := runtime.New()
	m := NewHttpManager(servers, &engine.Engine{}, store, app)

	c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://localhost:8080"}}}
	m.Update(&common.File{Data: c, Url: MustParseUrl("foo.yml")})

	try.GetRequest(t, "http://localhost:8080", map[string]string{})
	require.Equal(t, float64(1), app.Monitor.Http.RequestCounter.Value())
}

func TestHttpManager_Update(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, m *HttpManager, hook *logtest.Hook)
	}{
		{"nil config",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				m.Update(&common.File{Data: nil})
				require.Nil(t, hook.LastEntry())
			}},
		{"no version",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{Info: openapi.Info{Name: "foo"}}
				m.Update(&common.File{Data: c, Url: MustParseUrl("foo.yml")})
				require.Equal(t, "validation error foo.yml: no OpenApi version defined", hook.LastEntry().Message)
			}},
		{"no server specified",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}}
				m.Update(&common.File{Data: c, Url: MustParseUrl("foo.yml")})

				require.Contains(t, m.Servers, "80")
				entries := hook.Entries
				require.Len(t, entries, 3)
				require.Equal(t, "Adding new host '' on binding :80", entries[0].Message)
				require.Equal(t, "Adding service foo on binding :80 on path /", entries[1].Message)
				require.Equal(t, "processed config foo.yml", entries[2].Message)
			}},
		{"empty url",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: ""}}}
				m.Update(&common.File{Data: c, Url: MustParseUrl("foo.yml")})

				require.Contains(t, m.Servers, "80")
				entries := hook.Entries
				require.Len(t, entries, 3)
				require.Equal(t, "Adding new host '' on binding :80", entries[0].Message)
				require.Equal(t, "Adding service foo on binding :80 on path /", entries[1].Message)
				require.Equal(t, "processed config foo.yml", entries[2].Message)
			}},
		{
			"app contains config",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://:80"}}}
				m.Update(&common.File{Data: c, Url: MustParseUrl("foo.yml")})

				require.Contains(t, m.app.Http, "foo")
			},
		},
		{
			"app contains both config",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				foo := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://:80/foo"}}}
				bar := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "bar"}, Servers: []*openapi.Server{{Url: "http://:80/bar"}}}
				m.Update(&common.File{Data: foo, Url: MustParseUrl("foo.yml")})
				m.Update(&common.File{Data: bar, Url: MustParseUrl("bar.yml")})

				require.Contains(t, m.app.Http, "foo")
				require.Contains(t, m.app.Http, "bar")
			},
		},
		{"add new host http://:80",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://:80"}}}
				m.Update(&common.File{Data: c, Url: MustParseUrl("foo.yml")})

				require.Contains(t, m.Servers, "80")
				entries := hook.Entries
				require.Len(t, entries, 3)
				require.Equal(t, "Adding new host '' on binding :80", entries[0].Message)
				require.Equal(t, "Adding service foo on binding :80 on path /", entries[1].Message)
				require.Equal(t, "processed config foo.yml", entries[2].Message)
			}},
		{"add new host http://localhost:80",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://localhost:80"}}}
				m.Update(&common.File{Data: c, Url: MustParseUrl("foo.yml")})

				require.Contains(t, m.Servers, "80")
				entries := hook.Entries
				require.Len(t, entries, 3)
				require.Equal(t, "Adding new host 'localhost' on binding :80", entries[0].Message)
				require.Equal(t, "Adding service foo on binding :80 on path /", entries[1].Message)
				require.Equal(t, "processed config foo.yml", entries[2].Message)
			}},
		{"add new host http://localhost",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://localhost"}}}
				m.Update(&common.File{Data: c, Url: MustParseUrl("foo.yml")})

				require.Contains(t, m.Servers, "80")
				entries := hook.Entries
				require.Len(t, entries, 3)
				require.Equal(t, "Adding new host 'localhost' on binding :80", entries[0].Message)
				require.Equal(t, "Adding service foo on binding :80 on path /", entries[1].Message)
				require.Equal(t, "processed config foo.yml", entries[2].Message)
			}},
		{"invalid port format",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://localhost:foo"}}}
				m.Update(&common.File{Data: c, Url: MustParseUrl("foo.yml")})

				require.Len(t, m.Servers, 0)
				entries := hook.Entries
				require.Len(t, entries, 2)
				require.Equal(t, "error foo.yml: parse \"http://localhost:foo\": invalid port \":foo\" after host", entries[0].Message)
				require.Equal(t, "processed config foo.yml", entries[1].Message)
			}},
		{"invalid url format",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "$://"}}}
				m.Update(&common.File{Data: c, Url: MustParseUrl("foo.yml")})

				require.Len(t, m.Servers, 0)
				entries := hook.Entries
				require.Len(t, entries, 2)
				require.Equal(t, "error foo.yml: parse \"$://\": first path segment in URL cannot contain colon", entries[0].Message)
				require.Equal(t, "processed config foo.yml", entries[1].Message)
			}},
		{"add on same path",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "/foo"}}}
				m.Update(&common.File{Data: c, Url: MustParseUrl("foo.yml")})
				c = &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "bar"}, Servers: []*openapi.Server{{Url: "/foo"}}}
				m.Update(&common.File{Data: c, Url: MustParseUrl("foo.yml")})

				require.Len(t, m.Servers, 1)
				entries := hook.Entries
				require.Len(t, entries, 4)
				require.Equal(t, "Adding new host '' on binding :80", entries[0].Message)
				require.Equal(t, "Adding service foo on binding :80 on path /foo", entries[1].Message)
				require.Equal(t, "processed config foo.yml", entries[2].Message)
				require.Equal(t, "error on updating foo.yml: service 'foo' is already defined on path '/foo'", entries[3].Message)
			}},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			hook := test.NewNullLogger()
			store, err := cert.NewStore(&static.Config{})
			require.NoError(t, err)

			servers := HttpServers{}
			defer servers.Stop()

			m := NewHttpManager(servers, &engine.Engine{}, store, runtime.New())

			data.fn(t, m, hook)
		})

	}
}

func MustParseUrl(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
