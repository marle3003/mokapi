package server_test

import (
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/ldap/ldaptest"
	"mokapi/providers/directory"
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"mokapi/runtime/runtimetest"
	"mokapi/server"
	"mokapi/try"
	"net"
	"net/url"
	goruntime "runtime"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestLdapDirectory(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name    string
		configs []*dynamic.Config
		test    func(t *testing.T, m *server.LdapDirectoryManager, app *runtime.App)
	}{
		{
			name: "wrong config type",
			test: func(t *testing.T, m *server.LdapDirectoryManager, app *runtime.App) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: newConfig("foo", "data"),
				})

				require.Len(t, app.Ldap.List(), 0)
			},
		},
		{
			name: "add to runtime app",
			test: func(t *testing.T, m *server.LdapDirectoryManager, app *runtime.App) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: newConfig("foo", &directory.Config{
						Info: directory.Info{Name: "foo"},
					}),
				})
				require.NotNil(t, app.Ldap.Get("foo"))
			},
		},
		{
			name: "one ldap server",
			test: func(t *testing.T, m *server.LdapDirectoryManager, app *runtime.App) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: newConfig("foo", &directory.Config{
						Info:    directory.Info{Name: "foo"},
						Address: fmt.Sprintf(":%v", try.GetFreePort()),
					}),
				})

				c := ldap.NewClient(app.Ldap.Get("foo").Config().Address)
				err := c.Dial()
				require.NoError(t, err)
				_, err = c.Bind("", "")
				require.NoError(t, err)
			},
		},
		{
			name: "update ldap event",
			test: func(t *testing.T, m *server.LdapDirectoryManager, app *runtime.App) {
				addr1 := fmt.Sprintf(":%v", try.GetFreePort())
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: newConfig("foo", &directory.Config{
						Info:    directory.Info{Name: "foo"},
						Address: addr1,
					}),
				})

				addr2 := fmt.Sprintf(":%v", try.GetFreePort())
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: newConfig("foo", &directory.Config{
						Info:    directory.Info{Name: "foo"},
						Address: addr2,
					}),
				})

				c := ldap.NewClient(addr1)
				err := c.Dial()
				require.Error(t, err)

				c = ldap.NewClient(addr2)
				err = c.Dial()
				require.NoError(t, err)
				_, err = c.Bind("", "")
				require.NoError(t, err)
			},
		},
		{
			name: "delete ldap event",
			test: func(t *testing.T, m *server.LdapDirectoryManager, app *runtime.App) {
				addr := fmt.Sprintf(":%v", try.GetFreePort())
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: newConfig("foo", &directory.Config{
						Info:    directory.Info{Name: "foo"},
						Address: addr,
					}),
				})
				m.UpdateConfig(dynamic.ConfigEvent{
					Event: dynamic.Delete,
					Config: newConfig("foo", &directory.Config{
						Info:    directory.Info{Name: "foo"},
						Address: addr,
					}),
				})

				c := ldap.NewClient(addr)
				err := c.Dial()
				require.Error(t, err)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg, &dynamictest.Reader{})
			m := server.NewLdapDirectoryManager(enginetest.NewEngine(), nil, app)
			defer m.Stop()

			tc.test(t, m, app)
		})
	}
}

func newConfig(path string, data interface{}) *dynamic.Config {
	u, _ := url.Parse(path)
	return dynamictest.NewConfig(data, dynamictest.WithInfo(dynamic.ConfigInfo{Url: u}))
}

func TestLdapManager_PortIsBoundWhenUpdateConfigReturns(t *testing.T) {
	m := server.NewLdapDirectoryManager(enginetest.NewEngine(), nil, runtime.New(&static.Config{}, &dynamictest.Reader{}))
	defer m.Stop()

	requireBoundWhenUpdateReturns(t, func(port int) { m.UpdateConfig(ldapConfigEvent(port)) })
}

func TestLdapManager_BindErrorIsLoggedAndServerStaysRegistered(t *testing.T) {
	m := server.NewLdapDirectoryManager(enginetest.NewEngine(), nil, runtime.New(&static.Config{}, &dynamictest.Reader{}))
	defer m.Stop()

	requireBindErrorIsLoggedAndServerStaysRegistered(t,
		func(port int) { m.UpdateConfig(ldapConfigEvent(port)) },
		func(port int) string {
			return fmt.Sprintf("adding LDAP host 'x' on binding :%d failed: listen tcp :%d: bind: address already in use", port, port)
		},
		func(port int) string { return fmt.Sprintf("adding LDAP host 'x' on binding :%d", port) },
	)
}

func ldapConfigEvent(port int) dynamic.ConfigEvent {
	return dynamic.ConfigEvent{
		Config: newConfig("x.yml", &directory.Config{
			Info:    directory.Info{Name: "x"},
			Address: fmt.Sprintf(":%d", port),
		}),
	}
}

// The tests below stop an LDAP host right after it was started and expect its port to refuse
// connections when the call returns. GOMAXPROCS(1) keeps the Serve goroutine from running in between.
func TestLdapManager_PortIsFreeWhenDeleteReturns(t *testing.T) {
	requireOldPortIsFree(t, func(m *server.LdapDirectoryManager, port int) {
		e := ldapConfigEvent(port)
		e.Event = dynamic.Delete
		m.UpdateConfig(e)
	})
}

func TestLdapManager_PortIsFreeWhenAddressChanges(t *testing.T) {
	requireOldPortIsFree(t, func(m *server.LdapDirectoryManager, _ int) {
		m.UpdateConfig(ldapConfigEvent(try.GetFreePort()))
	})
}

func TestLdapManager_PortIsFreeWhenStopReturns(t *testing.T) {
	requireOldPortIsFree(t, func(m *server.LdapDirectoryManager, _ int) { m.Stop() })
}

// requireOldPortIsFree adds an LDAP host, calls release and expects that nothing listens on the host's
// first port anymore. It dials instead of binding the port again, because rebinding a just-released
// port occasionally fails with "address already in use".
func requireOldPortIsFree(t *testing.T, release func(m *server.LdapDirectoryManager, port int)) {
	logrus.SetOutput(io.Discard)
	defer goruntime.GOMAXPROCS(goruntime.GOMAXPROCS(1))

	for i := 0; i < 100; i++ {
		m := server.NewLdapDirectoryManager(enginetest.NewEngine(), nil, runtime.New(&static.Config{}, &dynamictest.Reader{}))
		port := try.GetFreePort()
		m.UpdateConfig(ldapConfigEvent(port))
		release(m, port)

		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
		if err == nil {
			_ = conn.Close()
		}
		m.Stop()
		require.Error(t, err, "port %d still accepts connections in iteration %d", port, i)
	}
}

func TestLdapManager_DeleteOfUnknownConfigDoesNotPanic(t *testing.T) {
	app := runtime.New(&static.Config{}, &dynamictest.Reader{})
	m := server.NewLdapDirectoryManager(enginetest.NewEngine(), nil, app)
	defer m.Stop()

	e := ldapConfigEvent(try.GetFreePort())
	e.Event = dynamic.Delete

	require.NotPanics(t, func() { m.UpdateConfig(e) })
}

func TestLdapManager_DeleteOfInfoWithoutConfigSnapshot(t *testing.T) {
	app := runtimetest.NewApp(runtimetest.WithLdapInfo("x", &runtime.LdapInfo{}))
	m := server.NewLdapDirectoryManager(enginetest.NewEngine(), nil, app)
	defer m.Stop()

	e := ldapConfigEvent(try.GetFreePort())
	e.Event = dynamic.Delete
	require.NotPanics(t, func() { m.UpdateConfig(e) })

	require.Nil(t, app.Ldap.Get("x"))
}

// ldapFile returns a file of the directory "x" with an entry cn=bob and the given address.
func ldapFile(path, address string) dynamic.ConfigEvent {
	entries := directory.NewEntryStore()
	entries.Set("cn=bob", directory.Entry{Dn: "cn=bob", Attributes: map[string][]string{"cn": {"bob"}}})
	return dynamic.ConfigEvent{
		Config: newConfig(path, &directory.Config{
			Info:    directory.Info{Name: "x"},
			Address: address,
			Entries: entries,
		}),
	}
}

func ldapSearch(t *testing.T, c *ldap.Client, cn string) []ldap.SearchResult {
	t.Helper()
	res, err := c.Search(&ldap.SearchRequest{Scope: ldap.ScopeWholeSubtree, Filter: fmt.Sprintf("(cn=%s)", cn)})
	require.NoError(t, err)
	return res.Results
}

func ldapAdd(t *testing.T, h ldap.Handler, dn string) {
	t.Helper()
	rr := ldaptest.NewRecorder()
	h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.AddRequest{Dn: dn, Attributes: []ldap.Attribute{{Type: "cn", Values: []string{dn[3:]}}}}))
	require.Equal(t, uint8(ldap.Success), rr.Message.(*ldap.AddResponse).ResultCode)
}

func dialLdap(t *testing.T, address string) *ldap.Client {
	t.Helper()
	c := ldap.NewClient(address)
	require.NoError(t, c.Dial())
	t.Cleanup(c.Close)
	return c
}

func TestLdapManager_Reload(t *testing.T) {
	logrus.SetOutput(io.Discard)

	newManager := func(t *testing.T) (*server.LdapDirectoryManager, *runtime.App) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		m := server.NewLdapDirectoryManager(enginetest.NewEngine(), nil, app)
		t.Cleanup(m.Stop)
		return m, app
	}

	t.Run("reload resets entries written by a client on the open connection", func(t *testing.T) {
		m, _ := newManager(t)
		address := fmt.Sprintf("127.0.0.1:%d", try.GetFreePort())
		m.UpdateConfig(ldapFile("x.yml", address))
		c := dialLdap(t, address)
		res, err := c.Add(&ldap.AddRequest{Dn: "cn=alice", Attributes: []ldap.Attribute{{Type: "cn", Values: []string{"alice"}}}})
		require.NoError(t, err)
		require.Equal(t, uint8(ldap.Success), res.ResultCode)
		require.Len(t, ldapSearch(t, c, "alice"), 1)

		m.UpdateConfig(ldapFile("x.yml", address))

		require.Empty(t, ldapSearch(t, c, "alice"))
		require.Len(t, ldapSearch(t, c, "bob"), 1)
	})

	t.Run("write on the handler of the old snapshot is lost after the reload", func(t *testing.T) {
		m, app := newManager(t)
		address := fmt.Sprintf("127.0.0.1:%d", try.GetFreePort())
		m.UpdateConfig(ldapFile("x.yml", address))
		oldHandler := app.Ldap.Get("x").Handler(monitor.NewLdap(), app.Events)
		c := dialLdap(t, address)

		m.UpdateConfig(ldapFile("x.yml", address))
		ldapAdd(t, oldHandler, "cn=alice")

		require.Empty(t, ldapSearch(t, c, "alice"))
	})

	t.Run("handler swap while searches are running", func(t *testing.T) {
		m, _ := newManager(t)
		address := fmt.Sprintf("127.0.0.1:%d", try.GetFreePort())
		m.UpdateConfig(ldapFile("x.yml", address))

		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			c := dialLdap(t, address)
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					res, err := c.Search(&ldap.SearchRequest{Scope: ldap.ScopeWholeSubtree, Filter: "(cn=bob)"})
					if err != nil || len(res.Results) != 1 {
						t.Errorf("search failed: %v", err)
						return
					}
				}
			}()
		}
		for i := 0; i < 50; i++ {
			m.UpdateConfig(ldapFile("x.yml", address))
		}
		wg.Wait()
	})

	t.Run("unbind is served by the handler of the current snapshot", func(t *testing.T) {
		m, app := newManager(t)
		address := fmt.Sprintf("127.0.0.1:%d", try.GetFreePort())
		m.UpdateConfig(ldapFile("x.yml", address))
		m.UpdateConfig(ldapFile("x.yml", address))
		c := dialLdap(t, address)
		_ = ldapSearch(t, c, "bob")
		before := app.Monitor.Ldap.RequestCounter.Sum(metrics.NewQuery(metrics.ByLabel("service", "x")))

		require.NoError(t, c.Unbind())

		require.Eventually(t, func() bool {
			return app.Monitor.Ldap.RequestCounter.Sum(metrics.NewQuery(metrics.ByLabel("service", "x"))) == before+1
		}, 2*time.Second, 5*time.Millisecond)
	})

	t.Run("a second file without address does not restart the server", func(t *testing.T) {
		m, _ := newManager(t)
		address := fmt.Sprintf("127.0.0.1:%d", try.GetFreePort())
		m.UpdateConfig(ldapFile("a.yml", address))
		c := dialLdap(t, address)

		m.UpdateConfig(ldapFile("b.yml", ""))

		require.Len(t, ldapSearch(t, c, "bob"), 1, "the connection is still served")
	})

	t.Run("a second file with another address restarts the server", func(t *testing.T) {
		m, _ := newManager(t)
		a := fmt.Sprintf("127.0.0.1:%d", try.GetFreePort())
		b := fmt.Sprintf("127.0.0.1:%d", try.GetFreePort())
		m.UpdateConfig(ldapFile("a.yml", a))

		m.UpdateConfig(ldapFile("b.yml", b))

		require.Error(t, ldap.NewClient(a).Dial())
		c := dialLdap(t, b)
		require.Len(t, ldapSearch(t, c, "bob"), 1)
	})

	t.Run("the host is started again after its last file was deleted", func(t *testing.T) {
		m, _ := newManager(t)
		address := fmt.Sprintf("127.0.0.1:%d", try.GetFreePort())
		e := ldapFile("x.yml", address)
		m.UpdateConfig(e)
		e.Event = dynamic.Delete
		m.UpdateConfig(e)

		m.UpdateConfig(ldapFile("x.yml", address))

		c := dialLdap(t, address)
		require.Len(t, ldapSearch(t, c, "bob"), 1)
	})
}
