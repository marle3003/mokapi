package runtime_test

import (
	"context"
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
	"mokapi/sortedmap"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndex_Ldap(t *testing.T) {
	toConfig := func(c *directory.Config) *dynamic.Config {
		cfg := &dynamic.Config{
			Info: dynamictest.NewConfigInfo(),
			Data: c,
		}
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
					r, err = app.Search(search.Request{QueryText: "Test", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 0
				})
				require.Len(t, r.Results, 0)
			},
		},
		{
			name: "Search entry",
			test: func(t *testing.T, app *runtime.App) {
				entries := &sortedmap.LinkedHashMap[string, directory.Entry]{}
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
	cfg := &dynamic.Config{
		Info: dynamictest.NewConfigInfo(),
		Data: api,
	}

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
