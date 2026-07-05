package server_test

import (
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/openapitest"
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/server"
	"mokapi/server/cert"
	"mokapi/try"
	"mokapi/version"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestHttpServers_Monitor(t *testing.T) {
	logrus.SetOutput(io.Discard)
	logtest.NewGlobal()
	cfg := &static.Config{}
	store, err := cert.NewStore(cfg)
	require.NoError(t, err)

	app := runtime.New(cfg, &dynamictest.Reader{})
	m := server.NewHttpManager(&engine.Engine{}, store, app)
	defer m.Stop()

	port := try.GetFreePort()
	u := fmt.Sprintf("http://localhost:%v", port)
	c := openapitest.NewConfig("3.0", openapitest.WithInfo("test", "1.0", ""), openapitest.WithServer(u, ""))
	openapitest.AppendPath("/foo", c, openapitest.WithOperation("get"))
	//c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: url}}}
	m.Update(dynamic.ConfigEvent{Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}, Data: c}})

	// give server time to start
	time.Sleep(time.Second * 1)
	try.GetRequest(t, u+"/foo", map[string]string{})
	require.Equal(t, float64(1), app.Monitor.Http.RequestCounter.Sum(metrics.NewQuery()))
}

func TestHttpManager_Update(t *testing.T) {
	testdata := []struct {
		name string
		test func(t *testing.T, app *runtime.App, m *server.HttpManager, hook *logtest.Hook)
	}{
		{
			name: "nil config",
			test: func(t *testing.T, app *runtime.App, m *server.HttpManager, hook *logtest.Hook) {
				m.Update(dynamic.ConfigEvent{Config: &dynamic.Config{Data: nil}})
				require.Nil(t, hook.LastEntry())
			}},
		{
			name: "app contains config",
			test: func(t *testing.T, app *runtime.App, m *server.HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: version.New("3.0"), Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://:80"}}}
				m.Update(dynamic.ConfigEvent{Config: &dynamic.Config{Data: c, Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}}})

				list := app.Http.List()
				require.Len(t, list, 1)
				require.Equal(t, "foo", list[0].Info.Name)
			},
		},
		{
			name: "app contains both config",
			test: func(t *testing.T, app *runtime.App, m *server.HttpManager, hook *logtest.Hook) {
				port := try.GetFreePort()
				u := fmt.Sprintf("http://localhost:%v", port)
				foo := &openapi.Config{OpenApi: version.New("3.0"), Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: u + "/foo"}}}
				bar := &openapi.Config{OpenApi: version.New("3.0"), Info: openapi.Info{Name: "bar"}, Servers: []*openapi.Server{{Url: u + "/bar"}}}
				m.Update(dynamic.ConfigEvent{Config: &dynamic.Config{Data: foo, Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}}})
				m.Update(dynamic.ConfigEvent{Config: &dynamic.Config{Data: bar, Info: dynamic.ConfigInfo{Url: try.MustUrl("bar.yml")}}})

				list := app.Http.List()
				require.Len(t, list, 2)
			},
		},
		{
			name: "add new host",
			test: func(t *testing.T, app *runtime.App, m *server.HttpManager, hook *logtest.Hook) {
				port := try.GetFreePort()
				c := &openapi.Config{OpenApi: version.New("3.0"), Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: fmt.Sprintf("http://:%v", port)}}}
				m.Update(dynamic.ConfigEvent{Config: &dynamic.Config{Data: c, Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}}})

				entries := hook.Entries
				require.Len(t, entries, 3)
				require.Equal(t, fmt.Sprintf("adding new HTTP host '' on binding :%v", port), entries[0].Message)
				require.Equal(t, fmt.Sprintf("adding service 'foo' on binding :%v on path /", port), entries[1].Message)
				require.Equal(t, "processed foo.yml", entries[2].Message)
			},
		},
		{
			name: "invalid port format",
			test: func(t *testing.T, app *runtime.App, m *server.HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: version.New("3.0"), Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://localhost:foo"}}}
				m.Update(dynamic.ConfigEvent{Config: &dynamic.Config{Data: c, Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}}})

				entries := hook.Entries
				require.Len(t, entries, 2)
				require.Equal(t, "url syntax error foo.yml: parse \"http://localhost:foo\": invalid port \":foo\" after host", entries[0].Message)
				require.Equal(t, "processed foo.yml", entries[1].Message)
			}},
		{
			name: "invalid url format",
			test: func(t *testing.T, app *runtime.App, m *server.HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: version.New("3.0"), Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "$://"}}}
				m.Update(dynamic.ConfigEvent{Config: &dynamic.Config{Data: c, Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}}})

				entries := hook.Entries
				require.Len(t, entries, 2)
				require.Equal(t, "url syntax error foo.yml: parse \"$://\": first path segment in URL cannot contain colon", entries[0].Message)
				require.Equal(t, "processed foo.yml", entries[1].Message)
			},
		},
		{
			name: "add on same path",
			test: func(t *testing.T, app *runtime.App, m *server.HttpManager, hook *logtest.Hook) {
				port := try.GetFreePort()
				u := fmt.Sprintf("http://:%v", port)
				c := &openapi.Config{OpenApi: version.New("3.0"), Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: u + "/foo"}}}
				m.Update(dynamic.ConfigEvent{Config: &dynamic.Config{Data: c, Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}}})
				c = &openapi.Config{OpenApi: version.New("3.0"), Info: openapi.Info{Name: "bar"}, Servers: []*openapi.Server{{Url: u + "/foo"}}}
				m.Update(dynamic.ConfigEvent{Config: &dynamic.Config{Data: c, Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}}})

				entries := hook.Entries
				require.Len(t, entries, 5)
				require.Equal(t, fmt.Sprintf("adding new HTTP host '' on binding :%v", port), entries[0].Message)
				require.Equal(t, fmt.Sprintf("adding service 'foo' on binding :%v on path /foo", port), entries[1].Message)
				require.Equal(t, "processed foo.yml", entries[2].Message)
				require.Equal(t, fmt.Sprintf("adding service 'bar' on binding :%v on path /foo", port), entries[3].Message)
			},
		},
		{
			name: "patching server",
			test: func(t *testing.T, app *runtime.App, m *server.HttpManager, hook *logtest.Hook) {
				port1 := try.GetFreePort()
				port2 := try.GetFreePort()

				u := fmt.Sprintf("http://:%v", port1)
				c := &openapi.Config{OpenApi: version.New("3.0"), Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: u}}}
				m.Update(dynamic.ConfigEvent{Config: &dynamic.Config{Data: c, Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}}})
				u = fmt.Sprintf("http://:%v", port2)
				c = &openapi.Config{OpenApi: version.New("3.0"), Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: u + "/foo"}}}
				m.Update(dynamic.ConfigEvent{Config: &dynamic.Config{Data: c, Info: dynamic.ConfigInfo{Url: try.MustUrl("foo.yml")}}})

				entries := hook.Entries
				require.Equal(t, fmt.Sprintf("adding new HTTP host '' on binding :%v", port1), entries[0].Message)
				require.Equal(t, fmt.Sprintf("adding service 'foo' on binding :%v on path /", port1), entries[1].Message)
				require.Equal(t, "processed foo.yml", entries[2].Message)
				require.Equal(t, fmt.Sprintf("removing 'foo' on binding %v on path /", port1), entries[3].Message)
				require.Equal(t, fmt.Sprintf("adding new HTTP host '' on binding :%v", port2), entries[4].Message)
				require.Equal(t, fmt.Sprintf("adding service 'foo' on binding :%v on path /foo", port2), entries[5].Message)
				require.Equal(t, fmt.Sprintf("stopping HTTP server on binding :%v", port1), entries[6].Message)
				require.Equal(t, "processed foo.yml", entries[7].Message)
			},
		},
	}

	for _, tc := range testdata {
		t.Run(tc.name, func(t *testing.T) {
			logrus.SetOutput(io.Discard)
			hook := logtest.NewGlobal()

			store, err := cert.NewStore(&static.Config{})
			require.NoError(t, err)

			cfg := &static.Config{Log: static.MokApiLog{Level: "debug"}}
			app := runtime.New(cfg, &dynamictest.Reader{})
			m := server.NewHttpManager(&engine.Engine{}, store, app)
			defer m.Stop()

			tc.test(t, app, m, hook)
		})

	}
}
