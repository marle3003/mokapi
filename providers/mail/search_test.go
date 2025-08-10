package mail_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/imap"
	"mokapi/providers/mail"
	"mokapi/runtime/events/eventstest"
	"mokapi/smtp"
	"testing"
	"time"
)

func TestImapHandler_Search(t *testing.T) {
	mustTime := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		return t
	}

	cfg := &mail.Config{
		Mailboxes: map[string]*mail.MailboxConfig{
			"alice@mokapi.io": {
				Username: "alice",
				Password: "foo",
			},
		},
	}

	addMail := func(s *mail.Store, m *mail.Mail) {
		mb := s.Mailboxes["alice@mokapi.io"]
		inbox := mb.Folders["INBOX"]
		inbox.Messages = append(inbox.Messages, m)
	}

	testcases := []struct {
		name string
		cfg  *mail.Config
		test func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context)
	}{
		{
			name: "Search Sequence Number: 1",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId:     uint32(10849681),
					Message: &smtp.Message{},
					Flags:   []imap.Flag{imap.FlagSeen},
				})

				set := &imap.IdSet{}
				set.AddId(1)
				res, err := h.Search(&imap.SearchRequest{
					Criteria: &imap.SearchCriteria{
						Seq: set,
					},
				}, ctx)
				require.NoError(t, err)
				r, ok := res.All.Nums()
				require.True(t, ok)
				require.Equal(t, []uint32{1}, r)
			},
		},
		{
			name: "Search Sequence Number: 1:*",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId:     uint32(10849681),
					Message: &smtp.Message{},
					Flags:   []imap.Flag{imap.FlagSeen},
				})
				addMail(s, &mail.Mail{
					UId:     uint32(10849682),
					Message: &smtp.Message{},
					Flags:   []imap.Flag{imap.FlagSeen},
				})

				set := &imap.IdSet{}
				set.AddRange(imap.SeqNum{Value: 1}, imap.SeqNum{Star: true})
				res, err := h.Search(&imap.SearchRequest{
					Criteria: &imap.SearchCriteria{
						Seq: set,
					},
				}, ctx)
				require.NoError(t, err)
				r, ok := res.All.Nums()
				require.True(t, ok)
				require.Equal(t, []uint32{1, 2}, r)
			},
		},
		{
			name: "Search Uid Number: 10849681",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId:     uint32(10849681),
					Message: &smtp.Message{},
					Flags:   []imap.Flag{imap.FlagSeen},
				})
				addMail(s, &mail.Mail{
					UId:     uint32(1084962),
					Message: &smtp.Message{},
					Flags:   []imap.Flag{imap.FlagSeen},
				})

				set := &imap.IdSet{}
				set.AddId(10849681)
				res, err := h.Search(&imap.SearchRequest{
					Criteria: &imap.SearchCriteria{
						UID: set,
					},
				}, ctx)
				require.NoError(t, err)
				r, ok := res.All.Nums()
				require.True(t, ok)
				require.Equal(t, []uint32{1}, r)
			},
		},
		{
			name: "UID Search Uid Number: 10849681",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId:     uint32(10849681),
					Message: &smtp.Message{},
					Flags:   []imap.Flag{imap.FlagSeen},
				})

				set := &imap.IdSet{}
				set.AddId(10849681)
				res, err := h.Search(&imap.SearchRequest{
					IsUid: true,
					Criteria: &imap.SearchCriteria{
						UID: set,
					},
				}, ctx)
				require.NoError(t, err)
				r, ok := res.All.Nums()
				require.True(t, ok)
				require.Equal(t, []uint32{10849681}, r)
			},
		},
		{
			name: "Search flag \\Answered",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId:     uint32(10849681),
					Message: &smtp.Message{},
					Flags:   []imap.Flag{imap.FlagAnswered},
				})
				addMail(s, &mail.Mail{
					UId:     uint32(10849682),
					Message: &smtp.Message{},
					Flags:   []imap.Flag{},
				})

				res, err := h.Search(&imap.SearchRequest{
					Criteria: &imap.SearchCriteria{
						Flag: []imap.Flag{imap.FlagAnswered},
					},
				}, ctx)
				require.NoError(t, err)
				r, ok := res.All.Nums()
				require.True(t, ok)
				require.Equal(t, []uint32{1}, r)
			},
		},
		{
			name: "Search has not flag \\Answered",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId:     uint32(10849681),
					Message: &smtp.Message{},
					Flags:   []imap.Flag{imap.FlagAnswered},
				})
				addMail(s, &mail.Mail{
					UId:     uint32(10849682),
					Message: &smtp.Message{},
					Flags:   []imap.Flag{},
				})

				res, err := h.Search(&imap.SearchRequest{
					Criteria: &imap.SearchCriteria{
						NotFlag: []imap.Flag{imap.FlagAnswered},
					},
				}, ctx)
				require.NoError(t, err)
				r, ok := res.All.Nums()
				require.True(t, ok)
				require.Equal(t, []uint32{2}, r)
			},
		},
		{
			name: "search from name",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId: uint32(10849681),
					Message: &smtp.Message{
						Headers: map[string]string{"From": "Bob <bob@mokapi.io>"},
					},
					Flags: []imap.Flag{imap.FlagAnswered},
				})
				addMail(s, &mail.Mail{
					UId:     uint32(10849682),
					Message: &smtp.Message{},
					Flags:   []imap.Flag{},
				})

				res, err := h.Search(&imap.SearchRequest{
					Criteria: &imap.SearchCriteria{
						Headers: []imap.HeaderCriteria{{Name: "From", Value: "bob"}},
					},
				}, ctx)
				require.NoError(t, err)
				r, ok := res.All.Nums()
				require.True(t, ok)
				require.Equal(t, []uint32{1}, r)
			},
		},
		{
			name: "search from name is case insensitive",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId: uint32(10849681),
					Message: &smtp.Message{
						Headers: map[string]string{"From": "Bob <support@mokapi.io>"},
					},
					Flags: []imap.Flag{imap.FlagAnswered},
				})
				addMail(s, &mail.Mail{
					UId:     uint32(10849682),
					Message: &smtp.Message{},
					Flags:   []imap.Flag{},
				})

				res, err := h.Search(&imap.SearchRequest{
					Criteria: &imap.SearchCriteria{
						Headers: []imap.HeaderCriteria{{Name: "From", Value: "bob"}},
					},
				}, ctx)
				require.NoError(t, err)
				r, ok := res.All.Nums()
				require.True(t, ok)
				require.Equal(t, []uint32{1}, r)
			},
		},
		{
			name: "search by before",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId:      uint32(10849681),
					Message:  &smtp.Message{},
					Flags:    []imap.Flag{imap.FlagAnswered},
					Received: mustTime("2025-07-27T13:01:30+00:00"),
				})
				addMail(s, &mail.Mail{
					UId:      uint32(10849682),
					Message:  &smtp.Message{},
					Flags:    []imap.Flag{},
					Received: mustTime("2025-07-22T13:01:30+00:00"),
				})

				before, _ := time.Parse(imap.SearchDateLayout, "27-Jul-2025")
				res, err := h.Search(&imap.SearchRequest{
					Criteria: &imap.SearchCriteria{
						Before: before,
					},
				}, ctx)
				require.NoError(t, err)
				r, ok := res.All.Nums()
				require.True(t, ok)
				require.Equal(t, []uint32{2}, r)
			},
		},
		{
			name: "search by sent before",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId: uint32(10849681),
					Message: &smtp.Message{
						Date: mustTime("2025-07-27T13:01:30+00:00"),
					},
					Flags: []imap.Flag{imap.FlagAnswered},
				})
				addMail(s, &mail.Mail{
					UId: uint32(10849682),
					Message: &smtp.Message{
						Date: mustTime("2025-07-22T13:01:30+00:00"),
					},
					Flags: []imap.Flag{},
				})

				before, _ := time.Parse(imap.SearchDateLayout, "27-Jul-2025")
				res, err := h.Search(&imap.SearchRequest{
					Criteria: &imap.SearchCriteria{
						SentBefore: before,
					},
				}, ctx)
				require.NoError(t, err)
				r, ok := res.All.Nums()
				require.True(t, ok)
				require.Equal(t, []uint32{2}, r)
			},
		},
		{
			name: "search by body",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId: uint32(10849681),
					Message: &smtp.Message{
						Body: "Hello World",
					},
					Flags: []imap.Flag{imap.FlagAnswered},
				})
				addMail(s, &mail.Mail{
					UId: uint32(10849682),
					Message: &smtp.Message{
						Body: "Foo",
					},
					Flags: []imap.Flag{},
				})

				res, err := h.Search(&imap.SearchRequest{
					Criteria: &imap.SearchCriteria{
						Body: []string{"hello"},
					},
				}, ctx)
				require.NoError(t, err)
				r, ok := res.All.Nums()
				require.True(t, ok)
				require.Equal(t, []uint32{1}, r)
			},
		},
		{
			name: "search not body",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId: uint32(10849681),
					Message: &smtp.Message{
						Body: "Hello World",
					},
					Flags: []imap.Flag{imap.FlagAnswered},
				})
				addMail(s, &mail.Mail{
					UId: uint32(10849682),
					Message: &smtp.Message{
						Body: "Foo",
					},
					Flags: []imap.Flag{},
				})

				res, err := h.Search(&imap.SearchRequest{
					Criteria: &imap.SearchCriteria{
						Not: []imap.SearchCriteria{
							{Body: []string{"hello"}},
						},
					},
				}, ctx)
				require.NoError(t, err)
				r, ok := res.All.Nums()
				require.True(t, ok)
				require.Equal(t, []uint32{2}, r)
			},
		},
		{
			name: "search or",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId: uint32(10849681),
					Message: &smtp.Message{
						Body: "Hello World",
					},
					Flags: []imap.Flag{imap.FlagAnswered},
				})
				addMail(s, &mail.Mail{
					UId: uint32(10849682),
					Message: &smtp.Message{
						Body: "Foo",
					},
					Flags: []imap.Flag{},
				})

				res, err := h.Search(&imap.SearchRequest{
					Criteria: &imap.SearchCriteria{
						Or: [][2]imap.SearchCriteria{
							{
								{Body: []string{"hello"}},
								{Body: []string{"foo"}},
							},
						},
					},
				}, ctx)
				require.NoError(t, err)
				r, ok := res.All.Nums()
				require.True(t, ok)
				require.Equal(t, []uint32{1, 2}, r)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := imap.NewClientContext(context.Background(), "127.0.0.1:84793")
			s := mail.NewStore(tc.cfg)
			h := mail.NewHandler(tc.cfg, s, enginetest.NewEngine(), &eventstest.Handler{})
			tc.test(t, h, s, ctx)
		})
	}
}
