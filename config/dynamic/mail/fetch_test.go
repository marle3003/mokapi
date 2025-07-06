package mail_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/mail"
	"mokapi/engine/enginetest"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/smtp"
	"testing"
	"time"
)

func TestImapHandler_Fetch(t *testing.T) {
	mustTime := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		return t
	}

	cfg := &mail.Config{
		Mailboxes: []mail.MailboxConfig{
			{
				Name:     "alice@mokapi.io",
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
			name: "BODY.PEEK[HEADER]",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId: uint32(10849681),
					Message: &smtp.Message{
						MessageId:   "abc123@mokapi.io",
						From:        []smtp.Address{{"Alice", "alice@mokapi.io"}},
						To:          []smtp.Address{{"Bob", "bob@mokapi.io"}},
						Subject:     "Hello Bob",
						ContentType: "text/html",
						Time:        mustTime("2025-07-06T13:01:30+00:00"),
						Size:        182,
						Body:        "<h1>Hello Bob</h1>",
					},
					Flags: []imap.Flag{imap.FlagSeen},
				})

				r := &imaptest.FetchRecorder{}
				err = h.Fetch(&imap.FetchRequest{
					Sequence: imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
					Options: imap.FetchOptions{
						Body: []imap.FetchBodySection{imap.FetchBodySection{
							Specifier: "header",
							Peek:      true,
						}},
					},
				}, r, ctx)
				require.NoError(t, err)
				require.Len(t, r.Messages, 1)
				require.Len(t, r.Messages[0].Body, 1)
				require.Equal(t, "header", r.Messages[0].Body[0].Section.Specifier)
				require.Equal(t, map[string]string{
					"message-id":   "abc123@mokapi.io",
					"from":         "Alice <alice@mokapi.io>",
					"to":           "Bob <bob@mokapi.io>",
					"date":         "06-Jul-2025 13:01:30 +0000",
					"subject":      "Hello Bob",
					"content-type": "text/html",
				}, r.Messages[0].Body[0].Headers)
			},
		},
		{
			name: "FULL",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId: uint32(10849681),
					Message: &smtp.Message{
						MessageId:   "abc123@mokapi.io",
						From:        []smtp.Address{{"Alice", "alice@mokapi.io"}},
						To:          []smtp.Address{{"Bob", "bob@mokapi.io"}},
						Subject:     "Hello Bob",
						ContentType: "text/html",
						Time:        mustTime("2025-07-06T13:01:30+00:00"),
						Size:        182,
						Body:        "<h1>Hello Bob</h1>",
					},
					Flags: []imap.Flag{imap.FlagSeen},
				})

				r := &imaptest.FetchRecorder{}
				err = h.Fetch(&imap.FetchRequest{
					Sequence: imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
					Options: imap.FetchOptions{
						Flags:         true,
						InternalDate:  true,
						RFC822Size:    true,
						Envelope:      true,
						BodyStructure: true,
					},
				}, r, ctx)
				require.NoError(t, err)
				require.Len(t, r.Messages, 1)
				require.Equal(t, "2025-07-06T13:01:30Z", r.Messages[0].InternalDate.Format(time.RFC3339))
				require.Equal(t, uint32(182), r.Messages[0].Size)
				require.Equal(t, []imap.Flag{imap.FlagSeen}, r.Messages[0].Flags)
				require.Equal(t, mustTime("2025-07-06T13:01:30+00:00"), r.Messages[0].Envelope.Date)
				require.Equal(t, "Hello Bob", r.Messages[0].Envelope.Subject)
				require.Equal(t, []imap.Address{{Name: "Alice", Mailbox: "alice", Host: "mokapi.io"}}, r.Messages[0].Envelope.From)
				require.Equal(t, []imap.Address{{Name: "Bob", Mailbox: "bob", Host: "mokapi.io"}}, r.Messages[0].Envelope.To)
				require.Equal(t, "abc123@mokapi.io", r.Messages[0].Envelope.MessageId)
				require.Equal(t, "text", r.Messages[0].BodyStructure.Type)
				require.Equal(t, "html", r.Messages[0].BodyStructure.Subtype)
				require.Equal(t, uint32(182), r.Messages[0].BodyStructure.Size)
			},
		},
		{
			name: "body structure with multipart",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId: uint32(10849681),
					Message: &smtp.Message{
						ContentType: "multipart/mixed",
						Body:        "hello",
						Attachments: []smtp.Attachment{
							{
								ContentType: "text/plain; charset=utf-8",
								Disposition: "",
								Data:        []byte("hello"),
								Header: map[string]string{
									"Content-Transfer-Encoding": "7bit",
								},
							},
							{
								ContentType: "image/png; charset=utf-8",
								Disposition: "attachment; name=cat.png",
								Data:        []byte("hello"),
								ContentId:   "foo",
								Header: map[string]string{
									"Content-Transfer-Encoding": "base64",
								},
							},
						},
					},
					Flags: []imap.Flag{imap.FlagSeen},
				})

				r := &imaptest.FetchRecorder{}
				err = h.Fetch(&imap.FetchRequest{
					Sequence: imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
					Options: imap.FetchOptions{
						BodyStructure: true,
					},
				}, r, ctx)
				require.NoError(t, err)
				require.Len(t, r.Messages, 1)
				require.Equal(t, "mixed", r.Messages[0].BodyStructure.Subtype)

				require.Equal(t, "text", r.Messages[0].BodyStructure.Parts[0].Type)
				require.Equal(t, "plain", r.Messages[0].BodyStructure.Parts[0].Subtype)
				require.Equal(t, "", r.Messages[0].BodyStructure.Parts[0].Disposition)
				require.Equal(t, uint32(5), r.Messages[0].BodyStructure.Parts[0].Size)
				require.Nil(t, r.Messages[0].BodyStructure.Parts[0].ContentId)
				require.Equal(t, "7bit", r.Messages[0].BodyStructure.Parts[0].Encoding)

				require.Equal(t, "image", r.Messages[0].BodyStructure.Parts[1].Type)
				require.Equal(t, "png", r.Messages[0].BodyStructure.Parts[1].Subtype)
				require.Equal(t, "attachment; name=cat.png", r.Messages[0].BodyStructure.Parts[1].Disposition)
				require.Equal(t, uint32(5), r.Messages[0].BodyStructure.Parts[1].Size)
				require.Equal(t, "foo", *r.Messages[0].BodyStructure.Parts[1].ContentId)
				require.Equal(t, "base64", r.Messages[0].BodyStructure.Parts[1].Encoding)
			},
		},
		{
			name: "BODY[1.TEXT]<0.10>",
			cfg:  cfg,
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				addMail(s, &mail.Mail{
					UId: uint32(10849681),
					Message: &smtp.Message{
						Body: "Lorem ipsum dolor sit amet",
					},
					Flags: []imap.Flag{imap.FlagSeen},
				})

				r := &imaptest.FetchRecorder{}
				err = h.Fetch(&imap.FetchRequest{
					Sequence: imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
					Options: imap.FetchOptions{
						Body: []imap.FetchBodySection{{
							Specifier: "text",
							Parts:     []int{1},
							Partially: &imap.BodyPart{Offset: uint32(0), Limit: uint32(10)},
						},
						},
					},
				}, r, ctx)
				require.NoError(t, err)
				require.Len(t, r.Messages, 1)
				require.Len(t, r.Messages[0].Body, 1)
				require.Equal(t, "text", r.Messages[0].Body[0].Section.Specifier)
				require.Equal(t, "Lorem ipsum dolor sit amet", r.Messages[0].Body[0].Body)
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
			h := mail.NewHandler(tc.cfg, s, enginetest.NewEngine())
			tc.test(t, h, s, ctx)
		})
	}
}
