package server_test

import (
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/providers/directory"
	"mokapi/runtime"
	"mokapi/server"
	"mokapi/try"
	"net"
	"net/url"
	goruntime "runtime"
	"testing"

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

				c := ldap.NewClient(app.Ldap.Get("foo").Address)
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
	return &dynamic.Config{
		Info: dynamic.ConfigInfo{Url: u},
		Data: data,
	}
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

// The tests below stop an LDAP host right after it was started and expect its port to be free
// when the call returns. GOMAXPROCS(1) keeps the Serve goroutine from running in between.
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

// requireOldPortIsFree adds an LDAP host, calls release and binds the host's first port.
func requireOldPortIsFree(t *testing.T, release func(m *server.LdapDirectoryManager, port int)) {
	logrus.SetOutput(io.Discard)
	defer goruntime.GOMAXPROCS(goruntime.GOMAXPROCS(1))

	for i := 0; i < 100; i++ {
		m := server.NewLdapDirectoryManager(enginetest.NewEngine(), nil, runtime.New(&static.Config{}, &dynamictest.Reader{}))
		port := try.GetFreePort()
		m.UpdateConfig(ldapConfigEvent(port))
		release(m, port)

		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			_ = ln.Close()
		}
		m.Stop()
		require.NoError(t, err, "iteration %d", i)
	}
}
