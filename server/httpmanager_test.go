package server

import (
	"fmt"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"io"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/openapitest"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/try"
	"net/url"
	"testing"
	"time"
)

func TestHttpServers_Monitor(t *testing.T) {
	logrus.SetOutput(io.Discard)
	logtest.NewGlobal()
	store, err := cert.NewStore(&static.Config{})
	require.NoError(t, err)

	app := runtime.New()
	m := NewHttpManager(&engine.Engine{}, store, app)
	defer m.Stop()

	port := try.GetFreePort()
	url := fmt.Sprintf("http://localhost:%v", port)
	c := openapitest.NewConfig("3.0", openapitest.WithInfo("test", "1.0", ""), openapitest.WithServer(url, ""))
	openapitest.AppendPath("/foo", c, openapitest.WithOperation("get", openapitest.NewOperation()))
	//c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: url}}}
	m.Update(common.NewConfig(MustParseUrl("foo.yml"), common.WithData(c)))

	// give server time to start
	time.Sleep(time.Second * 1)
	try.GetRequest(t, url+"/foo", map[string]string{})
	require.Equal(t, float64(1), app.Monitor.Http.RequestCounter.Sum())
}

func TestHttpManager_Update(t *testing.T) {
	testdata := []struct {
		name string
		fn   func(t *testing.T, m *HttpManager, hook *logtest.Hook)
	}{
		{"nil config",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				m.Update(&common.Config{Data: nil})
				require.Nil(t, hook.LastEntry())
			}},
		{
			"app contains config",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://:80"}}}
				m.Update(&common.Config{Data: c, Info: common.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				require.Contains(t, m.app.Http, "foo")
			},
		},
		{
			"app contains both config",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				port := try.GetFreePort()
				url := fmt.Sprintf("http://localhost:%v", port)
				foo := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: url + "/foo"}}}
				bar := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "bar"}, Servers: []*openapi.Server{{Url: url + "/bar"}}}
				m.Update(&common.Config{Data: foo, Info: common.ConfigInfo{Url: MustParseUrl("foo.yml")}})
				m.Update(&common.Config{Data: bar, Info: common.ConfigInfo{Url: MustParseUrl("bar.yml")}})

				require.Contains(t, m.app.Http, "foo")
				require.Contains(t, m.app.Http, "bar")
			},
		},
		{"add new host http://:X",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				port := try.GetFreePort()
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: fmt.Sprintf("http://:%v", port)}}}
				m.Update(&common.Config{Data: c, Info: common.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				entries := hook.Entries
				require.Len(t, entries, 3)
				require.Equal(t, fmt.Sprintf("adding new host '' on binding :%v", port), entries[0].Message)
				require.Equal(t, fmt.Sprintf("adding service foo on binding :%v on path /", port), entries[1].Message)
				require.Equal(t, "processed foo.yml", entries[2].Message)
			}},
		{"invalid port format",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "http://localhost:foo"}}}
				m.Update(&common.Config{Data: c, Info: common.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				entries := hook.Entries
				require.Len(t, entries, 2)
				require.Equal(t, "url syntax error foo.yml: parse \"http://localhost:foo\": invalid port \":foo\" after host", entries[0].Message)
				require.Equal(t, "processed foo.yml", entries[1].Message)
			}},
		{"invalid url format",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: "$://"}}}
				m.Update(&common.Config{Data: c, Info: common.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				entries := hook.Entries
				require.Len(t, entries, 2)
				require.Equal(t, "url syntax error foo.yml: parse \"$://\": first path segment in URL cannot contain colon", entries[0].Message)
				require.Equal(t, "processed foo.yml", entries[1].Message)
			}},
		{"add on same path",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				port := try.GetFreePort()
				url := fmt.Sprintf("http://:%v", port)
				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: url + "/foo"}}}
				m.Update(&common.Config{Data: c, Info: common.ConfigInfo{Url: MustParseUrl("foo.yml")}})
				c = &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "bar"}, Servers: []*openapi.Server{{Url: url + "/foo"}}}
				m.Update(&common.Config{Data: c, Info: common.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				entries := hook.Entries
				require.Len(t, entries, 5)
				require.Equal(t, fmt.Sprintf("adding new host '' on binding :%v", port), entries[0].Message)
				require.Equal(t, fmt.Sprintf("adding service foo on binding :%v on path /foo", port), entries[1].Message)
				require.Equal(t, "processed foo.yml", entries[2].Message)
				require.Equal(t, fmt.Sprintf("unable to add 'bar' on %v/foo: service 'foo' is already defined on path '/foo'", url), entries[3].Message)
			}},
		{"patching server",
			func(t *testing.T, m *HttpManager, hook *logtest.Hook) {
				port := try.GetFreePort()

				c := &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}}
				m.Update(&common.Config{Data: c, Info: common.ConfigInfo{Url: MustParseUrl("foo.yml")}})
				url := fmt.Sprintf("http://:%v", port)
				c = &openapi.Config{OpenApi: "3.0", Info: openapi.Info{Name: "foo"}, Servers: []*openapi.Server{{Url: url + "/foo"}}}
				m.Update(&common.Config{Data: c, Info: common.ConfigInfo{Url: MustParseUrl("foo.yml")}})

				entries := hook.Entries
				require.Len(t, entries, 6)
				require.Equal(t, fmt.Sprintf("adding new host '' on binding :80"), entries[0].Message)
				require.Equal(t, fmt.Sprintf("adding service foo on binding :80 on path /"), entries[1].Message)
				require.Equal(t, "processed foo.yml", entries[2].Message)
				require.Equal(t, fmt.Sprintf("adding new host '' on binding :%v", port), entries[3].Message)
				require.Equal(t, fmt.Sprintf("adding service foo on binding :%v on path /foo", port), entries[4].Message)
				require.Equal(t, "processed foo.yml", entries[5].Message)
			}},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			logrus.SetOutput(io.Discard)
			hook := logtest.NewGlobal()
			logrus.SetLevel(logrus.DebugLevel)
			store, err := cert.NewStore(&static.Config{})
			require.NoError(t, err)

			m := NewHttpManager(&engine.Engine{}, store, runtime.New())
			defer m.Stop()

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