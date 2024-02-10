package server

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/directory"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/runtime"
	"mokapi/try"
	"net/url"
	"testing"
)

func TestLdapDirectory(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*dynamic.Config
		test    func(t *testing.T, app *runtime.App)
	}{
		{
			name: "wrong config type",
			configs: []*dynamic.Config{
				newConfig("foo", "data"),
			},
			test: func(t *testing.T, app *runtime.App) {
				require.Len(t, app.Ldap, 0)
			},
		},
		{
			name: "add to runtime app",
			configs: []*dynamic.Config{
				newConfig("foo", &directory.Config{
					Info: directory.Info{Name: "foo"},
				}),
			},
			test: func(t *testing.T, app *runtime.App) {
				require.Contains(t, app.Ldap, "foo")
			},
		},
		{
			name: "one ldap server",
			configs: []*dynamic.Config{
				newConfig("foo", &directory.Config{
					Info:    directory.Info{Name: "foo"},
					Address: fmt.Sprintf(":%v", try.GetFreePort()),
				}),
			},
			test: func(t *testing.T, app *runtime.App) {
				c := ldap.NewClient(fmt.Sprintf(app.Ldap["foo"].Address))
				err := c.Dial()
				require.NoError(t, err)
				_, err = c.Bind("", "")
				require.NoError(t, err)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app := runtime.New()
			m := NewLdapDirectoryManager(enginetest.NewEngine(), nil, app)
			defer m.Stop()

			for _, c := range tc.configs {
				m.UpdateConfig(c)
			}
			tc.test(t, app)
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
