package runtime_test

import (
	"context"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/config/static"
	"mokapi/engine/enginetest"
	"mokapi/imap"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"mokapi/runtime/monitor"
	"mokapi/runtime/search"
	"mokapi/safe"
	"mokapi/smtp"
	"mokapi/smtp/smtptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
						Type:      "Mail",
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

				var r search.Result
				var err error
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 2
				})
				require.Len(t, r.Results, 2)

				app.Mail.Remove(toConfig(cfg))
				waitSearchIndex(t, func() bool {
					r, err = app.Search(search.Request{QueryText: "Test", Limit: 10})
					require.NoError(t, err)
					return len(r.Results) == 0
				})
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
						Type:      "Mail",
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

func TestIndex_Mail_Event(t *testing.T) {
	api := &mail.Config{
		Info: mail.Info{Name: "Test Mail Events"},
	}
	cfg := &dynamic.Config{
		Info: dynamictest.NewConfigInfo(),
		Data: api,
	}

	testcases := []struct {
		name string
		test func(t *testing.T, h runtime.MailHandler, app *runtime.App)
	}{
		{
			name: "search event by subject",
			test: func(t *testing.T, h runtime.MailHandler, app *runtime.App) {
				sendMail(h, "alice@foo.bar", "bob@foo.bar", "Test Mail", "A random body text")

				r, err := waitSearchResult(t, func() (search.Result, error) {
					return app.Search(search.Request{QueryText: `+subject:"Test Mail" +type:event`, Limit: 10})
				}, 1)

				require.NoError(t, err)
				require.Len(t, r.Results, 1)
				require.Equal(t, "Event", r.Results[0].Type)
				require.Equal(t, "Test Mail Events", r.Results[0].Domain)
				require.Equal(t, "Test Mail", r.Results[0].Title)
				require.Len(t, r.Results[0].Fragments, 2)
				require.Contains(t, r.Results[0].Fragments, "<mark>Test</mark> <mark>Mail</mark>")
				require.Contains(t, r.Results[0].Fragments, "<mark>event</mark>")
				require.Len(t, r.Results[0].Params, 5)
				require.Equal(t, "event", r.Results[0].Params["type"])
				require.Equal(t, "mail", r.Results[0].Params["traits.namespace"])
				require.Equal(t, "Test Mail Events", r.Results[0].Params["traits.name"])
				require.Contains(t, r.Results[0].Params, "id")
				require.Contains(t, r.Results[0].Params, "messageId")
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

			info := app.Mail.Add(cfg)
			pool := safe.NewPool(context.Background())
			app.Start(pool)
			defer pool.Stop()

			h := info.Handler(monitor.NewMail(), enginetest.NewEngine(), app.Events)

			tc.test(t, h, app)
		})
	}
}

func sendMail(h runtime.MailHandler, from, to, subject, body string) {
	ctx := context.Background()
	ctx = smtp.NewClientContext(ctx, "")

	rr := smtptest.NewRecorder()
	h.ServeSMTP(rr, smtp.NewMailRequest(from, ctx))

	h.ServeSMTP(rr, smtp.NewRcptRequest(to, ctx))

	h.ServeSMTP(rr, smtp.NewDataRequest(&smtp.Message{
		From:    []smtp.Address{{Address: from}},
		To:      []smtp.Address{{Address: to}},
		Date:    time.Now(),
		Subject: subject,
		Body:    body,
	}, ctx))
}
