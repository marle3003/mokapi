package server

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/providers/directory"
	"mokapi/runtime"
	"mokapi/try"
	"net/url"
	"testing"
)

func TestLdapDirectory(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*dynamic.Config
		test    func(t *testing.T, m *LdapDirectoryManager, app *runtime.App)
	}{
		{
			name: "wrong config type",
			test: func(t *testing.T, m *LdapDirectoryManager, app *runtime.App) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: newConfig("foo", "data"),
				})

				require.Len(t, app.Ldap.List(), 0)
			},
		},
		{
			name: "add to runtime app",
			test: func(t *testing.T, m *LdapDirectoryManager, app *runtime.App) {
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
			test: func(t *testing.T, m *LdapDirectoryManager, app *runtime.App) {
				m.UpdateConfig(dynamic.ConfigEvent{
					Config: newConfig("foo", &directory.Config{
						Info:    directory.Info{Name: "foo"},
						Address: fmt.Sprintf(":%v", try.GetFreePort()),
					}),
				})

				c := ldap.NewClient(fmt.Sprintf(app.Ldap.Get("foo").Address))
				err := c.Dial()
				require.NoError(t, err)
				_, err = c.Bind("", "")
				require.NoError(t, err)
			},
		},
		{
			name: "update ldap event",
			test: func(t *testing.T, m *LdapDirectoryManager, app *runtime.App) {
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
			test: func(t *testing.T, m *LdapDirectoryManager, app *runtime.App) {
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
			app := runtime.New(cfg)
			m := NewLdapDirectoryManager(enginetest.NewEngine(), nil, app)
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
