package runtime_test

import (
	"context"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/ldap/ldaptest"
	"mokapi/providers/directory"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/runtime/search"
	"mokapi/safe"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestIndex_Ldap(t *testing.T) {
	logrus.SetOutput(io.Discard)

	toConfig := func(c *directory.Config) *dynamic.Config {
		cfg := dynamictest.NewConfig(c, dynamictest.WithInfo(dynamictest.NewConfigInfo()))
		return cfg
	}

	testcases := []struct {
		name string
		test func(t *testing.T, app *runtime.App)
	}{
		{
			name: "Search by name",
			test: func(t *testing.T, app *runtime.App) {
				cfg := &directory.Config{
					Info: directory.Info{Name: "foo"},
				}
				app.Ldap.Add(toConfig(cfg), enginetest.NewEngine())
				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "foo", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "LDAP",
						Title:     "foo",
						Fragments: []string{"<mark>foo</mark>"},
						Params: map[string]string{
							"type":    "ldap",
							"service": "foo",
							"entries": "0",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "config should be removed from index",
			test: func(t *testing.T, app *runtime.App) {
				cfg := &directory.Config{
					Info: directory.Info{Name: "foo"},
				}
				app.Ldap.Add(toConfig(cfg), enginetest.NewEngine())
				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)

				app.Ldap.Remove(toConfig(cfg))
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 0
				})
				require.Len(t, r.Results, 0)
			},
		},
		{
			name: "entry should be removed from index",
			test: func(t *testing.T, app *runtime.App) {
				entries := directory.NewEntryStore()
				entries.Set("foo", directory.Entry{
					Dn:         "cn=alice,dc=foo,dc=com",
					Attributes: map[string][]string{"objectclass": {"user"}},
				})
				cfg := &directory.Config{
					Info:    directory.Info{Name: "foo"},
					Entries: entries,
				}
				app.Ldap.Add(toConfig(cfg), enginetest.NewEngine())
				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "alice", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})

				app.Ldap.Remove(toConfig(cfg))

				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "alice", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 0
				})
			},
		},
		{
			name: "Search entry",
			test: func(t *testing.T, app *runtime.App) {
				entries := directory.NewEntryStore()
				entries.Set("foo", directory.Entry{
					Dn:         "cn=alice,dc=foo,dc=com",
					Attributes: map[string][]string{"objectclass": {"user"}},
				})

				cfg := &directory.Config{
					Info:    directory.Info{Name: "foo"},
					Entries: entries,
				}
				app.Ldap.Add(toConfig(cfg), enginetest.NewEngine())
				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "alice", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 1
				})
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "LDAP",
						Domain:    "foo",
						Title:     "cn=alice,dc=foo,dc=com",
						Fragments: []string{"cn=<mark>alice</mark>,dc=foo,dc=com"},
						Params: map[string]string{
							"type":    "ldap",
							"service": "foo",
							"entry":   "cn=alice,dc=foo,dc=com",
						},
					},
					r.Results[0])

				r, err = app.Search(search.Request{QueryText: "objectclass", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)

				r, err = app.Search(search.Request{QueryText: "user", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := runtime.New(
				&static.Config{
					Api: static.Api{
						Search: static.Search{
							Enabled:  true,
							InMemory: true,
						},
					},
				}, &dynamictest.Reader{})

			pool := safe.NewPool(context.Background())
			app.Start(pool)
			defer pool.Stop()

			tc.test(t, app)
		})
	}
}

func TestIndex_Ldap_Event(t *testing.T) {
	api := &directory.Config{Info: directory.Info{Name: "Test LDAP Events"}}
	cfg := dynamictest.NewConfig(api, dynamictest.WithInfo(dynamictest.NewConfigInfo()))

	testcases := []struct {
		name string
		test func(t *testing.T, h ldap.Handler, app *runtime.App)
	}{
		{
			name: "search event by operation",
			test: func(t *testing.T, h ldap.Handler, app *runtime.App) {
				h.ServeLDAP(&ldaptest.ResponseRecorder{}, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					BaseDN: "ou=people,o=search",
					Filter: "(cn=user)",
				}))

				r, err := waitSearchResult(t, func() (search.Result, error) {
					return app.Search(search.Request{QueryText: "+operation:search +type:event", Limit: 10})
				}, 1)

				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t, "Event", r.Results[0].Type)
				require.Equal(t, "Test LDAP Events", r.Results[0].Domain)
				require.Equal(t, "(cn=user)", r.Results[0].Title)
				require.Len(t, r.Results[0].Fragments, 2)
				require.Contains(t, r.Results[0].Fragments, "<mark>Search</mark>")
				require.Contains(t, r.Results[0].Fragments, "<mark>event</mark>")
				require.Len(t, r.Results[0].Params, 6)
				require.Equal(t, "event", r.Results[0].Params["type"])
				require.Equal(t, "ldap", r.Results[0].Params["traits.namespace"])
				require.Equal(t, "Test LDAP Events", r.Results[0].Params["traits.name"])
				require.Equal(t, "search", r.Results[0].Params["traits.operation"])
				require.Equal(t, "ou=people,o=search", r.Results[0].Params["baseDN"])
				require.Contains(t, r.Results[0].Params, "id")
				require.NotEmpty(t, r.Results[0].Time)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			app := runtime.New(
				&static.Config{
					Api: static.Api{
						Search: static.Search{
							Enabled:  true,
							InMemory: true,
						},
					},
				}, &dynamictest.Reader{})

			info := app.Ldap.Add(cfg, enginetest.NewEngine())
			pool := safe.NewPool(context.Background())
			app.Start(pool)
			defer pool.Stop()

			h := info.Handler(monitor.NewLdap(), app.Events)
			tc.test(t, h, app)
		})
	}
}

func TestIndex_Ldap_MergedFiles(t *testing.T) {
	logrus.SetOutput(io.Discard)

	ldapFile := func(url, dn string) *dynamic.Config {
		entries := directory.NewEntryStore()
		entries.Set(dn, directory.Entry{
			Dn:         dn,
			Attributes: map[string][]string{"objectclass": {"user"}},
		})
		return dynamictest.NewConfig(&directory.Config{Info: directory.Info{Name: "foo"}, Entries: entries}, dynamictest.WithInfo(dynamictest.NewConfigInfo(dynamictest.WithUrl(url))))
	}
	found := func(t *testing.T, app *runtime.App, text string) bool {
		r, err := app.Search(search.Request{QueryText: text, Limit: 10})
		require.NoError(t, err)
		return len(r.Results) > 0
	}

	app := runtime.New(&static.Config{Api: static.Api{Search: static.Search{Enabled: true, InMemory: true}}}, &dynamictest.Reader{})
	pool := safe.NewPool(context.Background())
	app.Start(pool)
	defer pool.Stop()

	app.Ldap.Add(ldapFile("https://mokapi.io/a.yml", "cn=alice,dc=foo,dc=com"), enginetest.NewEngine())
	b := ldapFile("https://mokapi.io/b.yml", "cn=bob,dc=foo,dc=com")
	app.Ldap.Add(b, enginetest.NewEngine())
	waitSearchIndex(t, func() bool { return found(t, app, "bob") && found(t, app, "alice") })

	app.Ldap.Remove(b)

	waitSearchIndex(t, func() bool {
		return found(t, app, "alice") && !found(t, app, "bob")
	})
}
