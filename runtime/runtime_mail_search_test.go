package runtime_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/imap"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"mokapi/runtime/search"
	"testing"
)

func TestIndex_Mail(t *testing.T) {
	toConfig := func(c *mail.Config) *dynamic.Config {
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
				cfg := &mail.Config{
					Info: mail.Info{Name: "foo"},
				}
				app.Mail.Add(toConfig(cfg))
				r, err := app.Search(search.Request{QueryText: "foo", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "MAIL",
						Title:     "foo",
						Fragments: []string{"<mark>foo</mark>"},
						Params: map[string]string{
							"type":    "mail",
							"service": "foo",
						},
					},
					r.Results[0])
			},
		},
		{
			name: "config should be removed from index",
			test: func(t *testing.T, app *runtime.App) {
				cfg := &mail.Config{
					Info: mail.Info{Name: "foo"},
				}
				app.Mail.Add(toConfig(cfg))
				r, err := app.Search(search.Request{QueryText: "foo", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)

				app.Mail.Remove(toConfig(cfg))
				r, err = app.Search(search.Request{QueryText: "foo", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 0)
			},
		},
		{
			name: "Search mailbox",
			test: func(t *testing.T, app *runtime.App) {
				cfg := &mail.Config{
					Info: mail.Info{Name: "foo"},
					Mailboxes: map[string]*mail.MailboxConfig{
						"alice@mokapi.io": {
							Username:    "username",
							Password:    "password",
							Description: "mailbox description",
							Folders: map[string]*mail.FolderConfig{
								"inbox": {Flags: []string{string(imap.Trash)}},
							},
						},
					},
				}
				app.Mail.Add(toConfig(cfg))
				r, err := app.Search(search.Request{QueryText: "alice", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t,
					search.ResultItem{
						Type:      "MAIL",
						Domain:    "foo",
						Title:     "alice@mokapi.io",
						Fragments: []string{"<mark>alice</mark>@mokapi.io"},
						Params: map[string]string{
							"type":    "mail",
							"service": "foo",
							"mailbox": "alice@mokapi.io",
						},
					},
					r.Results[0])

				r, err = app.Search(search.Request{QueryText: "username", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				r, err = app.Search(search.Request{QueryText: "password", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				r, err = app.Search(search.Request{QueryText: "description", Limit: 10})
				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				r, err = app.Search(search.Request{QueryText: "trash", Limit: 10})
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
