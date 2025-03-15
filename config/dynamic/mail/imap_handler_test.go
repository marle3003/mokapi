package mail_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/mail"
	"mokapi/engine/enginetest"
	"mokapi/imap"
	"testing"
)

func TestImapHandler(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *mail.Config
		test func(t *testing.T, h *mail.Handler, ctx context.Context)
	}{
		{
			name: "Login successfully",
			cfg: &mail.Config{
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, ctx context.Context) {
				err := h.Login("alice", "foo", ctx)
				require.NoError(t, err)
			},
		},
		{
			name: "Login failed",
			cfg: &mail.Config{
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, ctx context.Context) {
				err := h.Login("alice", "bar", ctx)
				require.EqualError(t, err, "invalid credentials")
			},
		},
		{
			name: "Select Inbox",
			cfg: &mail.Config{
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				r, err := h.Select("Inbox", ctx)
				require.NoError(t, err)
				require.Len(t, r.Flags, 5)
			},
		},
		{
			name: "Select invalid mailbox",
			cfg: &mail.Config{
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Foo", ctx)
				require.EqualError(t, err, "mailbox not found")
			},
		},
		{
			name: "Unselect a mailbox",
			cfg: &mail.Config{
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("inbox", ctx)
				require.NoError(t, err)
				err = h.Unselect(ctx)
				require.NoError(t, err)
				c := imap.ClientFromContext(ctx)
				require.Equal(t, "", c.Session["selected"])
			},
		},
		{
			name: "List *",
			cfg: &mail.Config{
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				r, err := h.List("", "*", nil, ctx)
				require.NoError(t, err)
				require.Len(t, r, 1)
				require.Equal(t, imap.ListEntry{
					Flags:     nil,
					Delimiter: "/",
					Name:      "INBOX",
				}, r[0])
			},
		},
		//{
		//	name: "List foo",
		//	cfg: &mail.Config{
		//		Mailboxes: []mail.MailboxConfig{
		//			{
		//				Name:     "alice@mokapi.io",
		//				Username: "alice",
		//				Password: "foo",
		//			},
		//		},
		//	},
		//	test: func(t *testing.T, h *mail.Handler, ctx context.Context) {
		//		_ = h.Login("alice", "foo", ctx)
		//		r, err := h.List("", "foo", nil, ctx)
		//		require.NoError(t, err)
		//		require.Len(t, r, 0)
		//	},
		//},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := imap.NewClientContext(context.Background(), "127.0.0.1:84793")
			s := mail.NewStore(tc.cfg)
			h := mail.NewHandler(tc.cfg, s, enginetest.NewEngine())
			tc.test(t, h, ctx)
		})
	}
}
