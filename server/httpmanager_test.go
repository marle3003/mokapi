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
	"net"
	"net/http"
	"strings"
	"sync"
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

func newHttpManager(t *testing.T) *server.HttpManager {
	logrus.SetOutput(io.Discard)

	store, err := cert.NewStore(&static.Config{})
	require.NoError(t, err)

	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	return server.NewHttpManager(&engine.Engine{}, store, app)
}

// newApiEvent returns an event for an API named name that answers GET /x with 200 on
// http://localhost:<port><basePath>.
func newApiEvent(name string, port int, basePath string, event dynamic.Event) dynamic.ConfigEvent {
	c := openapitest.NewConfig("3.0",
		openapitest.WithInfo(name, "1.0", ""),
		openapitest.WithServer(fmt.Sprintf("http://localhost:%v%v", port, basePath), ""),
		openapitest.WithPath("/x", openapitest.WithOperation("get", openapitest.WithResponse(200))),
	)
	return dynamic.ConfigEvent{
		Event:  event,
		Config: &dynamic.Config{Info: dynamic.ConfigInfo{Url: try.MustUrl(name + ".yml")}, Data: c},
	}
}

// updateInParallel mimics the ConfigWatcher, which calls the listeners of every config event
// in its own goroutine.
func updateInParallel(m *server.HttpManager, events ...dynamic.ConfigEvent) {
	var wg sync.WaitGroup
	for _, e := range events {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Update(e)
		}()
	}
	wg.Wait()
}

func isPortOpen(port int) bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// serves reports whether url answers with status 200 within a few seconds. Unlike
// try.GetRequest it does not fail the test, so the caller can decide whether to retry.
func serves(url string) bool {
	deadline := time.Now().Add(3 * time.Second)
	for {
		res, err := http.Get(url)
		if err == nil {
			_ = res.Body.Close()
			if res.StatusCode == http.StatusOK {
				return true
			}
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// bindFailed reports whether a server could not listen on its port. try.GetFreePort releases the
// port before the server binds it, so another process may take it in between. The server only
// logs such an error. The bug these tests guard against stops a server silently, without an error.
func bindFailed(hook *logtest.Hook) bool {
	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.ErrorLevel && strings.HasPrefix(entry.Message, "failed to start") {
			return true
		}
	}
	return false
}

// repeatWithFreePorts runs scenario the given number of times. A scenario reports whether all
// expected servers were reachable. If it was not, and a server failed to bind its port, the
// scenario runs again with fresh ports (up to three attempts). Any other failure fails the test.
func repeatWithFreePorts(t *testing.T, times int, scenario func(hook *logtest.Hook) (reachable bool)) {
	const maxAttempts = 3
	for i := 0; i < times; i++ {
		for attempt := 1; ; attempt++ {
			hook := logtest.NewGlobal()
			if scenario(hook) {
				break
			}
			if attempt == maxAttempts || !bindFailed(hook) {
				t.Fatalf("API is not reachable (attempt %v, port bind failed: %v)", attempt, bindFailed(hook))
			}
		}
	}
}

func TestHttpManager_ParallelUpdates_DifferentPorts(t *testing.T) {
	repeatWithFreePorts(t, 20, func(hook *logtest.Hook) bool {
		m := newHttpManager(t)
		defer m.Stop()

		ports := []int{try.GetFreePort(), try.GetFreePort(), try.GetFreePort()}
		var events []dynamic.ConfigEvent
		for n, port := range ports {
			events = append(events, newApiEvent(fmt.Sprintf("api%v", n), port, "", dynamic.Create))
		}
		updateInParallel(m, events...)

		for _, port := range ports {
			if !serves(fmt.Sprintf("http://localhost:%v/x", port)) {
				return false
			}
		}
		return true
	})
}

func TestHttpManager_ParallelUpdates_SamePort(t *testing.T) {
	repeatWithFreePorts(t, 20, func(hook *logtest.Hook) bool {
		m := newHttpManager(t)
		defer m.Stop()

		port := try.GetFreePort()
		updateInParallel(m,
			newApiEvent("a", port, "/a", dynamic.Create),
			newApiEvent("b", port, "/b", dynamic.Create),
		)

		hosts := 0
		for _, entry := range hook.AllEntries() {
			if strings.HasPrefix(entry.Message, "adding new HTTP host") {
				hosts++
			}
		}
		require.Equal(t, 1, hosts, "both APIs must share one server")

		return serves(fmt.Sprintf("http://localhost:%v/a/x", port)) &&
			serves(fmt.Sprintf("http://localhost:%v/b/x", port))
	})
}

func TestHttpManager_DeleteWhileAdding(t *testing.T) {
	repeatWithFreePorts(t, 10, func(hook *logtest.Hook) bool {
		m := newHttpManager(t)
		defer m.Stop()

		portA := try.GetFreePort()
		portB := try.GetFreePort()
		m.Update(newApiEvent("a", portA, "", dynamic.Create))
		if !serves(fmt.Sprintf("http://localhost:%v/x", portA)) {
			return false
		}

		updateInParallel(m,
			newApiEvent("a", portA, "", dynamic.Delete),
			newApiEvent("b", portB, "", dynamic.Create),
		)

		if !serves(fmt.Sprintf("http://localhost:%v/x", portB)) {
			return false
		}
		try.WaitFor(t, func() bool { return !isPortOpen(portA) })
		return true
	})
}

func TestHttpManager_RemovingLastApiStopsServer(t *testing.T) {
	m := newHttpManager(t)
	defer m.Stop()

	port := try.GetFreePort()
	m.Update(newApiEvent("a", port, "", dynamic.Create))
	try.GetRequest(t, fmt.Sprintf("http://localhost:%v/x", port), nil, try.HasStatusCode(200))

	m.Update(newApiEvent("a", port, "", dynamic.Delete))

	try.WaitFor(t, func() bool { return !isPortOpen(port) })
}
