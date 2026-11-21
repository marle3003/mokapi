package runtime_test

import (
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/ldap/ldaptest"
	"mokapi/providers/directory"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/runtime/events/eventstest"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"net/url"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestApp_AddLdap(t *testing.T) {
	logrus.SetOutput(io.Discard)

	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App)
	}{
		{
			name: "event store available",
			test: func(t *testing.T, app *runtime.App) {
				app.Ldap.Add(newLdapConfig("https://mokapi.io", &directory.Config{Info: directory.Info{Name: "foo"}}), enginetest.NewEngine())

				require.NotNil(t, app.Ldap.Get("foo"))
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("ldap").WithName("foo"))
				require.NoError(t, err, "event store should be available")
			},
		},
		{
			name: "bind request is counted in monitor",
			test: func(t *testing.T, app *runtime.App) {
				info := app.Ldap.Add(newLdapConfig("https://mokapi.io", &directory.Config{Info: directory.Info{Name: "foo"}}), enginetest.NewEngine())
				m := monitor.NewLdap()
				h := info.Handler(m, app.Events)

				r := ldaptest.NewRequest(0, &ldap.BindRequest{Version: 3, Auth: ldap.Simple})
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, r)

				require.Equal(t, float64(1), m.RequestCounter.Sum(metrics.NewQuery()))
			},
		},
		{
			name: "retrieve configs",
			test: func(t *testing.T, app *runtime.App) {
				info := app.Ldap.Add(newLdapConfig("https://mokapi.io", &directory.Config{}), enginetest.NewEngine())

				configs := filesOf(info)
				require.Len(t, configs, 1)
				require.Equal(t, "https://mokapi.io", configs[0].Info.Url.String())
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg, &dynamictest.Reader{})
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
				info := app.Ldap.Get("foo")
				require.Equal(t, "bar", info.Config().Info.Description)
				configs := filesOf(info)
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
				info := app.Ldap.Get("foo")
				require.Equal(t, "foo", info.Config().Info.Description)
			},
		},
		{
			name: "order only by filename",
			configs: []*dynamic.Config{
				newLdapConfig("https://a.io/b", &directory.Config{Info: directory.Info{Name: "foo", Description: "foo"}}),
				newLdapConfig("https://mokapi.io/a", &directory.Config{Info: directory.Info{Name: "foo", Description: "bar"}}),
			},
			test: func(t *testing.T, app *runtime.App) {
				info := app.Ldap.Get("foo")
				require.Equal(t, "foo", info.Config().Info.Description)
			},
		},
		{
			name: "patch does not reset events",
			configs: []*dynamic.Config{
				newLdapConfig("https://a.io/b", &directory.Config{Info: directory.Info{Name: "foo", Description: "foo"}}),
			},
			test: func(t *testing.T, app *runtime.App) {
				err := app.Events.Push(&eventstest.Event{Name: "bar"}, events.NewTraits().WithNamespace("ldap").WithName("foo"))
				require.NoError(t, err)

				app.Ldap.Add(newLdapConfig("https://mokapi.io/a", &directory.Config{Info: directory.Info{Name: "foo", Description: "bar"}}), enginetest.NewEngine())

				e := app.Events.GetEvents(events.NewTraits().WithNamespace("ldap").WithName("foo"))
				require.Len(t, e, 1)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := &static.Config{}
			app := runtime.New(cfg, &dynamictest.Reader{})
			for _, c := range tc.configs {
				app.Ldap.Add(c, enginetest.NewEngine())
			}
			tc.test(t, app)
		})
	}
}

func TestIsLdapConfig(t *testing.T) {
	_, ok := runtime.IsLdapConfig(&dynamic.Config{Data: &directory.Config{}})
	require.True(t, ok)
	_, ok = runtime.IsLdapConfig(&dynamic.Config{Data: "foo"})
	require.False(t, ok)
}

func newLdapConfig(name string, config *directory.Config) *dynamic.Config {
	c := &dynamic.Config{Data: config}
	u, _ := url.Parse(name)
	c.Info.Url = u
	return c
}

func TestLdapStore_Remove(t *testing.T) {
	logrus.SetOutput(io.Discard)

	newFile := func(url, name string) *dynamic.Config {
		return newLdapConfig(url, &directory.Config{Info: directory.Info{Name: name}})
	}

	t.Run("unknown config does not panic and keeps the store usable", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})

		requireStoreUsableAfterRemoveOfUnknownConfig(t,
			func() { app.Ldap.Remove(newFile("https://mokapi.io/unknown.yml", "unknown")) },
			func() { app.Ldap.Add(newFile("https://mokapi.io/a.yml", "foo"), enginetest.NewEngine()) },
		)
	})

	t.Run("two files removed at the same time", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a := newFile("https://mokapi.io/a.yml", "foo")
		b := newFile("https://mokapi.io/b.yml", "foo")
		app.Ldap.Add(a, enginetest.NewEngine())
		app.Ldap.Add(b, enginetest.NewEngine())

		removeConcurrently(app.Ldap.Remove, a, b)

		require.Equal(t, 0, app.Ldap.Len())
	})
}

// ldapFileWithEntries returns a file that holds one entry per given dn, each with the attribute
// mail set to the given value.
func ldapFileWithEntries(url, name, description string, mailByDn map[string]string) *dynamic.Config {
	entries := directory.NewEntryStore()
	for dn, mail := range mailByDn {
		entries.Set(dn, directory.Entry{Dn: dn, Attributes: map[string][]string{"mail": {mail}}})
	}
	return newLdapConfig(url, &directory.Config{
		Info:    directory.Info{Name: name, Description: description},
		Entries: entries,
	})
}

func TestLdapStore_ConcurrentAccess(t *testing.T) {
	logrus.SetOutput(io.Discard)
	emitter := enginetest.NewEngine()

	t.Run("reader does not race with reload of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		reload := func(description string) *dynamic.Config {
			return ldapFileWithEntries("https://mokapi.io/a.yml", "foo", description, map[string]string{"cn=alice": "alice@foo.bar"})
		}
		app.Ldap.Add(reload("first"), enginetest.NewEngine())

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				app.Ldap.Add(reload("second"), emitter)
			}
		}()
		go func() {
			defer wg.Done()
			// what the dashboard API handler does: get the info, then read the config
			for i := 0; i < 200; i++ {
				info := app.Ldap.Get("foo")
				cfg := info.Config()
				_ = cfg.Info.Name
				_ = cfg.Entries.Entries()
				_ = filesOf(info)
			}
		}()
		wg.Wait()
	})

	t.Run("reader does not race with remove", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a := ldapFileWithEntries("https://mokapi.io/a.yml", "foo", "", map[string]string{"cn=alice": "alice@foo.bar"})
		b := ldapFileWithEntries("https://mokapi.io/b.yml", "foo", "", map[string]string{"cn=bob": "bob@foo.bar"})
		info := app.Ldap.Add(a, enginetest.NewEngine())

		done := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			defer close(done)
			for i := 0; i < 200; i++ {
				app.Ldap.Add(b, emitter)
				app.Ldap.Remove(b)
			}
		}()
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				if cfg := info.Config(); cfg != nil {
					_ = cfg.Info.Name
				}
				_ = filesOf(info)
			}
		}()
		wg.Wait()
	})
}

func TestLdapStore_MergedConfigs(t *testing.T) {
	logrus.SetOutput(io.Discard)

	t.Run("merge does not change the parsed file config", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		a := ldapFileWithEntries("https://mokapi.io/a.yml", "foo", "", map[string]string{"cn=alice": "a@foo.bar"})
		b := ldapFileWithEntries("https://mokapi.io/b.yml", "foo", "", map[string]string{"cn=alice": "b@foo.bar", "cn=bob": "bob@foo.bar"})
		app.Ldap.Add(a, enginetest.NewEngine())

		app.Ldap.Add(b, enginetest.NewEngine())

		fileA := a.Data.(*directory.Config)
		require.Equal(t, 1, fileA.Entries.Len(), "entry of b.yml is not added to the file a.yml")
		alice, _ := fileA.Entries.Get("cn=alice")
		require.Equal(t, []string{"a@foo.bar"}, alice.Attributes["mail"], "attribute of b.yml is not added to the file a.yml")
	})
}

func TestLdapInfo_Snapshot(t *testing.T) {
	logrus.SetOutput(io.Discard)

	alice := map[string]string{"cn=alice": "alice@foo.bar"}
	fileWithDescription := func(description string) *dynamic.Config {
		return ldapFileWithEntries("https://mokapi.io/a.yml", "foo", description, alice)
	}

	t.Run("snapshot is not changed by a reload of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		info := app.Ldap.Add(fileWithDescription("first"), enginetest.NewEngine())
		snapshot := info.Config()

		app.Ldap.Add(ldapFileWithEntries("https://mokapi.io/a.yml", "foo", "second", map[string]string{"cn=bob": "bob@foo.bar"}), enginetest.NewEngine())

		require.Equal(t, "first", snapshot.Info.Description)
		_, hasAlice := snapshot.Entries.Get("cn=alice")
		_, hasBob := snapshot.Entries.Get("cn=bob")
		require.True(t, hasAlice)
		require.False(t, hasBob)
		require.NotSame(t, snapshot, info.Config())
		require.Equal(t, "second", info.Config().Info.Description)
	})

	t.Run("published config is not the config of the file", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		file := fileWithDescription("first")

		info := app.Ldap.Add(file, enginetest.NewEngine())

		fileConfig := file.Data.(*directory.Config)
		require.NotSame(t, fileConfig, info.Config())
		require.NotSame(t, fileConfig.Entries, info.Config().Entries)
		require.Equal(t, 1, info.Config().Entries.Len())
	})

	t.Run("config is nil after the last file is removed", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		file := fileWithDescription("first")
		info := app.Ldap.Add(file, enginetest.NewEngine())

		app.Ldap.Remove(file)

		require.Nil(t, info.Config())
		require.Nil(t, app.Ldap.Get("foo"))
	})

	t.Run("Configs returns a copy", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		app.Ldap.Add(ldapFileWithEntries("https://mokapi.io/a.yml", "foo", "", alice), enginetest.NewEngine())
		info := app.Ldap.Add(ldapFileWithEntries("https://mokapi.io/b.yml", "foo", "", alice), enginetest.NewEngine())

		configs := filesOf(info)
		require.Len(t, configs, 2)
		configs[0] = nil

		require.NotContains(t, filesOf(info), (*dynamic.Config)(nil))
		require.Len(t, filesOf(info), 2)
	})

	t.Run("Configs does not race with Add", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		emitter := enginetest.NewEngine()
		info := app.Ldap.Add(ldapFileWithEntries("https://mokapi.io/a.yml", "foo", "", alice), emitter)
		b := ldapFileWithEntries("https://mokapi.io/b.yml", "foo", "", map[string]string{"cn=bob": "bob@foo.bar"})

		done := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			defer close(done)
			for i := 0; i < 200; i++ {
				app.Ldap.Add(b, emitter)
			}
		}()
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				_ = filesOf(info)
			}
		}()
		wg.Wait()
	})

	t.Run("client write to the snapshot does not reach the file config", func(t *testing.T) {
		app := runtime.New(&static.Config{}, &dynamictest.Reader{})
		file := fileWithDescription("first")
		info := app.Ldap.Add(file, enginetest.NewEngine())
		h := info.Handler(monitor.NewLdap(), app.Events)

		rr := ldaptest.NewRecorder()
		h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.AddRequest{Dn: "cn=bob"}))
		require.Equal(t, uint8(ldap.Success), rr.Message.(*ldap.AddResponse).ResultCode)
		rr = ldaptest.NewRecorder()
		h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyRequest{
			Dn: "cn=alice",
			Items: []ldap.ModificationItem{{
				Operation:    ldap.ReplaceOperation,
				Modification: ldap.Modification{Type: "mail", Values: []string{"changed@foo.bar"}},
			}},
		}))
		require.Equal(t, uint8(ldap.Success), rr.Message.(*ldap.ModifyResponse).ResultCode)

		fileConfig := file.Data.(*directory.Config)
		require.Equal(t, 1, fileConfig.Entries.Len())
		e, _ := fileConfig.Entries.Get("cn=alice")
		require.Equal(t, []string{"alice@foo.bar"}, e.Attributes["mail"])
		_, ok := info.Config().Entries.Get("cn=bob")
		require.True(t, ok, "the write is visible in the running directory")
	})
}
