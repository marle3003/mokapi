package runtime_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/providers/directory"
	"mokapi/runtime"
	"mokapi/runtime/search"
	"mokapi/sortedmap"
	"testing"
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
				r, err := app.Search(search.Request{QueryText: "foo", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "LDAP",
						Title:     "foo",
						Fragments: []string{"<mark>foo</mark>"},
						Params: map[string]string{
							"type":    "ldap",
							"service": "foo",
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
				r, err := app.Search(search.Request{Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)

				app.Ldap.Remove(toConfig(cfg))
				r, err = app.Search(search.Request{Limit: 10})
				require.NoError(t, err)
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
				r, err := app.Search(search.Request{QueryText: "alice", Limit: 10})
				require.NoError(t, err)
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
							Enabled: true,
						},
					},
				})
			tc.test(t, app)
		})
	}
}
