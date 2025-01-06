package runtime_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/directory"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/ldap/ldaptest"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"net/url"
	"testing"
)

func TestApp_AddLdap(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App)
	}{
		{
			name: "event store available",
			test: func(t *testing.T, app *runtime.App) {
				app.AddLdap(newLdapConfig("https://mokapi.io", &directory.Config{Info: directory.Info{Name: "foo"}}), enginetest.NewEngine())

				require.Contains(t, app.Ldap, "foo")
				err := events.Push("bar", events.NewTraits().WithNamespace("ldap").WithName("foo"))
				require.NoError(t, err, "event store should be available")
			},
		},
		{
			name: "bind request is counted in monitor",
			test: func(t *testing.T, app *runtime.App) {
				info := app.AddLdap(newLdapConfig("https://mokapi.io", &directory.Config{Info: directory.Info{Name: "foo"}}), enginetest.NewEngine())
				m := monitor.NewLdap()
				h := info.Handler(m)

				r := ldaptest.NewRequest(0, &ldap.BindRequest{Version: 3, Auth: ldap.Simple})
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, r)

				require.Equal(t, float64(1), m.Bind.Sum())
			},
		},
		{
			name: "retrieve configs",
			test: func(t *testing.T, app *runtime.App) {
				info := app.AddLdap(newLdapConfig("https://mokapi.io", &directory.Config{}), enginetest.NewEngine())

				configs := info.Configs()
				require.Len(t, configs, 1)
				require.Equal(t, "https://mokapi.io", configs[0].Info.Url.String())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			defer events.Reset()

			app := runtime.New()
			tc.test(t, app)
		})
	}
}

func TestApp_AddLdap_Patching(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*dynamic.Config
		test    func(t *testing.T, app *runtime.App)
	}{
		{
			name: "overwrite value",
			configs: []*dynamic.Config{
				newLdapConfig("https://mokapi.io/a", &directory.Config{Info: directory.Info{Name: "foo", Description: "foo"}}),
				newLdapConfig("https://mokapi.io/b", &directory.Config{Info: directory.Info{Name: "foo", Description: "bar"}}),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Ldap["foo"]
				require.Equal(t, "bar", info.Info.Description)
				configs := info.Configs()
				require.Len(t, configs, 2)
			},
		},
		{
			name: "a is patched with b",
			configs: []*dynamic.Config{
				newLdapConfig("https://mokapi.io/b", &directory.Config{Info: directory.Info{Name: "foo", Description: "foo"}}),
				newLdapConfig("https://mokapi.io/a", &directory.Config{Info: directory.Info{Name: "foo", Description: "bar"}}),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Ldap["foo"]
				require.Equal(t, "foo", info.Info.Description)
			},
		},
		{
			name: "order only by filename",
			configs: []*dynamic.Config{
				newLdapConfig("https://a.io/b", &directory.Config{Info: directory.Info{Name: "foo", Description: "foo"}}),
				newLdapConfig("https://mokapi.io/a", &directory.Config{Info: directory.Info{Name: "foo", Description: "bar"}}),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Ldap["foo"]
				require.Equal(t, "foo", info.Info.Description)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer events.Reset()

			app := runtime.New()
			for _, c := range tc.configs {
				app.AddLdap(c, enginetest.NewEngine())
			}
			tc.test(t, app)
		})
	}
}

func TestIsLdapConfig(t *testing.T) {
	require.True(t, runtime.IsLdapConfig(&dynamic.Config{Data: &directory.Config{}}))
	require.False(t, runtime.IsLdapConfig(&dynamic.Config{Data: "foo"}))
}

func newLdapConfig(name string, config *directory.Config) *dynamic.Config {
	c := &dynamic.Config{Data: config}
	u, _ := url.Parse(name)
	c.Info.Url = u
	return c
}
