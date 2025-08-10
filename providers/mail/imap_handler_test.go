package mail_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"mokapi/providers/mail"
	"mokapi/runtime/events/eventstest"
	"mokapi/smtp"
	"testing"
	"time"
)

func TestImapHandler(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *mail.Config
		test func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context)
	}{
		{
			name: "Login successfully",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				err := h.Login("alice", "foo", ctx)
				require.NoError(t, err)
			},
		},
		{
			name: "Login failed",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				err := h.Login("alice", "bar", ctx)
				require.EqualError(t, err, "invalid credentials")
			},
		},
		{
			name: "Select Inbox",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				r, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)
				require.Len(t, r.Flags, 5)
			},
		},
		{
			name: "Select Inbox/foo",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
						Folders: map[string]*mail.FolderConfig{
							"inbox": {
								Folders: map[string]*mail.FolderConfig{
									"foo": {},
								},
							},
						},
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				r, err := h.Select("Inbox/foo", false, ctx)
				require.NoError(t, err)
				require.Len(t, r.Flags, 5)
			},
		},
		{
			name: "Select invalid mailbox",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Foo", false, ctx)
				require.EqualError(t, err, "mailbox not found")
			},
		},
		{
			name: "Unselect a mailbox",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("inbox", false, ctx)
				require.NoError(t, err)
				err = h.Unselect(ctx)
				require.NoError(t, err)
				c := imap.ClientFromContext(ctx)
				require.Equal(t, "", c.Session["selected"])
			},
		},
		{
			name: "Status Inbox",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)

				mb := s.Mailboxes["alice@mokapi.io"]
				mb.EnsureInbox()
				inbox := mb.Folders["INBOX"]
				inbox.Messages = append(inbox.Messages, &mail.Mail{})

				r, err := h.Status(&imap.StatusRequest{Mailbox: "inbox"}, ctx)
				require.NoError(t, err)
				require.Equal(t, uint32(1), r.Messages)
			},
		},
		{
			name: "List *",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
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
		{
			name: "List * with nested folders",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
						Folders: map[string]*mail.FolderConfig{
							"foo": {
								Folders: map[string]*mail.FolderConfig{
									"bar": {},
								},
							},
						},
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				r, err := h.List("", "*", nil, ctx)
				require.NoError(t, err)
				require.Len(t, r, 3)
				require.Equal(t, imap.ListEntry{
					Flags:     nil,
					Delimiter: "/",
					Name:      "INBOX",
				}, r[0])
				require.Equal(t, imap.ListEntry{
					Flags:     nil,
					Delimiter: "/",
					Name:      "foo",
				}, r[1])
				require.Equal(t, imap.ListEntry{
					Flags:     nil,
					Delimiter: "/",
					Name:      "foo/bar",
				}, r[2])
			},
		},
		{
			name: "List inbox should be first",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
						Folders: map[string]*mail.FolderConfig{
							"ABC": {},
						},
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				r, err := h.List("", "*", nil, ctx)
				require.NoError(t, err)
				require.Len(t, r, 2)
				require.Equal(t, []imap.ListEntry{
					{
						Flags:     nil,
						Delimiter: "/",
						Name:      "INBOX",
					},
					{
						Flags:     nil,
						Delimiter: "/",
						Name:      "ABC",
					},
				}, r)
			},
		},
		{
			name: "List foo",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
						Folders: map[string]*mail.FolderConfig{
							"foo": {
								Flags: []string{"foo", "bar"},
							},
						},
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				r, err := h.List("", "foo", nil, ctx)
				require.NoError(t, err)
				require.Len(t, r, 1)
				require.Equal(t, "foo", r[0].Name)
				require.Equal(t, []imap.MailboxFlags{"foo", "bar"}, r[0].Flags)
			},
		},
		{
			name: "List pattern Archive/*/foo",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
						Folders: map[string]*mail.FolderConfig{
							"Archive": {
								Folders: map[string]*mail.FolderConfig{
									"2025": {
										Folders: map[string]*mail.FolderConfig{
											"foo": {},
										},
									},
									"2026": {
										Folders: map[string]*mail.FolderConfig{
											"bar": {},
										},
									},
								},
							},
						},
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				r, err := h.List("", "Archive/*/foo", nil, ctx)
				require.NoError(t, err)
				require.Len(t, r, 1)
				require.Equal(t, "Archive/2025/foo", r[0].Name)
			},
		},
		{
			name: "List Archive %",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
						Folders: map[string]*mail.FolderConfig{
							"Archive": {
								Folders: map[string]*mail.FolderConfig{
									"2025": {
										Folders: map[string]*mail.FolderConfig{
											"foo": {},
										},
									},
									"2026": {
										Folders: map[string]*mail.FolderConfig{
											"bar": {},
										},
									},
								},
							},
						},
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				r, err := h.List("Archive", "%", nil, ctx)
				require.NoError(t, err)
				require.Equal(t, []imap.ListEntry{
					{Delimiter: "/", Flags: nil, Name: "Archive/2025"},
					{Delimiter: "/", Flags: nil, Name: "Archive/2026"},
				}, r)
			},
		},
		{
			name: "List Archive/%",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
						Folders: map[string]*mail.FolderConfig{
							"Archive": {
								Folders: map[string]*mail.FolderConfig{
									"2025": {

										Folders: map[string]*mail.FolderConfig{
											"foo": {},
										},
									},
									"2026": {
										Folders: map[string]*mail.FolderConfig{
											"bar": {},
										},
									},
								},
							},
						},
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				r, err := h.List("", "Archive/%", nil, ctx)
				require.NoError(t, err)
				require.Equal(t, []imap.ListEntry{
					{Delimiter: "/", Flags: nil, Name: "Archive/2025"},
					{Delimiter: "/", Flags: nil, Name: "Archive/2026"},
				}, r)
			},
		},
		{
			name: "Add deleted flag to message",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				inbox := mb.Folders["INBOX"]
				inbox.Messages = append(inbox.Messages, &mail.Mail{})

				r := &imaptest.FetchRecorder{}
				err = h.Store(&imap.StoreRequest{
					Sequence: imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
					Action:   "add",
					Flags:    []imap.Flag{imap.FlagDeleted},
				}, r, ctx)
				require.NoError(t, err)
				require.Len(t, r.Messages, 1)
				require.Equal(t, []imap.Flag{imap.FlagDeleted}, r.Messages[0].Flags)
				require.Equal(t, []imap.Flag{imap.FlagDeleted}, inbox.Messages[0].Flags)
			},
		},
		{
			name: "Add flag should not result in duplicate entries",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				inbox := mb.Folders["INBOX"]
				inbox.Messages = append(inbox.Messages, &mail.Mail{
					Flags: []imap.Flag{imap.FlagDeleted},
				})

				r := &imaptest.FetchRecorder{}
				err = h.Store(&imap.StoreRequest{
					Sequence: imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
					Action:   "add",
					Flags:    []imap.Flag{imap.FlagDeleted},
				}, r, ctx)
				require.NoError(t, err)
				require.Len(t, r.Messages, 1)
				require.Equal(t, []imap.Flag{imap.FlagDeleted}, r.Messages[0].Flags)
				require.Equal(t, []imap.Flag{imap.FlagDeleted}, inbox.Messages[0].Flags)
			},
		},
		{
			name: "Remove flag",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				inbox := mb.Folders["INBOX"]
				inbox.Messages = append(inbox.Messages, &mail.Mail{
					Flags: []imap.Flag{imap.FlagDeleted},
				})

				r := &imaptest.FetchRecorder{}
				err = h.Store(&imap.StoreRequest{
					Sequence: imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
					Action:   "remove",
					Flags:    []imap.Flag{imap.FlagDeleted},
				}, r, ctx)
				require.NoError(t, err)
				require.Len(t, r.Messages, 1)
				require.Empty(t, r.Messages[0].Flags)
				require.Empty(t, inbox.Messages[0].Flags)
			},
		},
		{
			name: "Replace flag",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				inbox := mb.Folders["INBOX"]
				inbox.Messages = append(inbox.Messages, &mail.Mail{
					Flags: []imap.Flag{imap.FlagDeleted},
				})

				r := &imaptest.FetchRecorder{}
				err = h.Store(&imap.StoreRequest{
					Sequence: imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
					Action:   "replace",
					Flags:    []imap.Flag{imap.FlagSeen},
				}, r, ctx)
				require.NoError(t, err)
				require.Len(t, r.Messages, 1)
				require.Equal(t, []imap.Flag{imap.FlagSeen}, r.Messages[0].Flags)
				require.Equal(t, []imap.Flag{imap.FlagSeen}, inbox.Messages[0].Flags)
			},
		},
		{
			name: "Add deleted flag to message using UID",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				inbox := mb.Folders["INBOX"]
				inbox.Messages = append(inbox.Messages, &mail.Mail{
					UId: uint32(143),
				})

				r := &imaptest.FetchRecorder{}
				err = h.Store(&imap.StoreRequest{
					Sequence: imap.IdSet{Ids: []imap.Set{imap.IdNum(143)}, IsUid: true},
					Action:   "add",
					Flags:    []imap.Flag{imap.FlagDeleted},
				}, r, ctx)
				require.NoError(t, err)
				require.Len(t, r.Messages, 1)
				require.Equal(t, uint32(143), r.Messages[0].Uid)
				require.Equal(t, []imap.Flag{imap.FlagDeleted}, r.Messages[0].Flags)
				require.Equal(t, []imap.Flag{imap.FlagDeleted}, inbox.Messages[0].Flags)
			},
		},
		{
			name: "Expunge",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				inbox := mb.Folders["INBOX"]
				inbox.Messages = append(inbox.Messages,
					&mail.Mail{
						Flags: []imap.Flag{imap.FlagDeleted},
					},
					&mail.Mail{},
				)

				r := &imaptest.ExpungeRecorder{}
				err = h.Expunge(nil, r, ctx)
				require.NoError(t, err)
				require.Equal(t, []uint32{1}, r.Ids)
				require.Len(t, inbox.Messages, 1)
			},
		},
		{
			name: "Expunge 1",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				inbox := mb.Folders["INBOX"]
				inbox.Messages = append(inbox.Messages,
					&mail.Mail{
						Flags: []imap.Flag{imap.FlagDeleted},
					},
					&mail.Mail{
						Flags: []imap.Flag{imap.FlagDeleted},
					},
				)

				r := &imaptest.ExpungeRecorder{}
				err = h.Expunge(&imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}}, r, ctx)
				require.NoError(t, err)
				require.Equal(t, []uint32{1}, r.Ids)
				require.Len(t, inbox.Messages, 1)
			},
		},
		{
			name: "UID Expunge",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				inbox := mb.Folders["INBOX"]
				inbox.Messages = append(inbox.Messages,
					&mail.Mail{
						UId:   uint32(143),
						Flags: []imap.Flag{imap.FlagDeleted},
					},
					&mail.Mail{
						UId:   uint32(144),
						Flags: []imap.Flag{imap.FlagDeleted},
					},
				)

				r := &imaptest.ExpungeRecorder{}
				err = h.Expunge(&imap.IdSet{Ids: []imap.Set{imap.IdNum(143)}, IsUid: true}, r, ctx)
				require.NoError(t, err)
				require.Equal(t, []uint32{143}, r.Ids)
				require.Len(t, inbox.Messages, 1)
			},
		},
		{
			name: "Create folder",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				err := h.Login("alice", "foo", ctx)
				require.NoError(t, err)

				err = h.Create("foo", &imap.CreateOptions{}, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				require.Contains(t, mb.Folders, "foo")
			},
		},
		{
			name: "Create folder with flags",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				err := h.Login("alice", "foo", ctx)
				require.NoError(t, err)

				err = h.Create("foo", &imap.CreateOptions{Flags: []imap.MailboxFlags{"foo"}}, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				require.Contains(t, mb.Folders, "foo")
				require.Equal(t, []imap.MailboxFlags{"foo"}, mb.Folders["foo"].Flags)
			},
		},
		{
			name: "Create folder below INBOX",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				err := h.Login("alice", "foo", ctx)
				require.NoError(t, err)

				err = h.Create("inbox/foo", &imap.CreateOptions{}, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				require.Contains(t, mb.Folders["INBOX"].Folders, "foo")
			},
		},
		{
			name: "Move",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				inbox := mb.Folders["INBOX"]
				inbox.Messages = append(inbox.Messages,
					&mail.Mail{UId: uint32(111)},
				)
				mb.AddFolder(&mail.Folder{
					Name: "foo",
				})

				r := &imaptest.MoveRecorder{}
				err = h.Move(
					&imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
					"foo",
					r,
					ctx)
				require.NoError(t, err)
				require.Greater(t, r.Copy.UIDValidity, uint32(0))
				require.Equal(t, imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}}, r.Copy.SourceUIDs)
				require.Equal(t, imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}}, r.Copy.DestUIDs)
				require.Len(t, r.Ids, 1)

				require.Len(t, inbox.Messages, 0)
				require.Len(t, mb.Folders["foo"].Messages, 1)
			},
		},
		{
			name: "UID Move",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				inbox := mb.Folders["INBOX"]
				inbox.Messages = append(inbox.Messages,
					&mail.Mail{UId: uint32(111)},
				)
				mb.AddFolder(&mail.Folder{
					Name: "foo",
				})

				r := &imaptest.MoveRecorder{}
				err = h.Move(
					&imap.IdSet{Ids: []imap.Set{imap.IdNum(111)}, IsUid: true},
					"foo",
					r,
					ctx)
				require.NoError(t, err)
				require.Greater(t, r.Copy.UIDValidity, uint32(0))
				require.Equal(t, imap.IdSet{Ids: []imap.Set{imap.IdNum(111)}, IsUid: true}, r.Copy.SourceUIDs)

				f := mb.Folders["foo"]
				require.Equal(t, imap.IdSet{Ids: []imap.Set{imap.IdNum(f.Messages[0].UId)}, IsUid: true}, r.Copy.DestUIDs)
				require.Len(t, r.Ids, 1)

				require.Len(t, inbox.Messages, 0)
				require.Len(t, mb.Folders["foo"].Messages, 1)
			},
		},
		{
			name: "Copy",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				inbox := mb.Folders["INBOX"]
				inbox.Messages = append(inbox.Messages,
					&mail.Mail{UId: uint32(111)},
				)
				mb.AddFolder(&mail.Folder{
					Name: "foo",
				})

				r := &imaptest.CopyRecorder{}
				err = h.Copy(
					&imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
					"foo",
					r,
					ctx)
				require.NoError(t, err)
				require.Greater(t, r.Copy.UIDValidity, uint32(0))
				require.Equal(t, imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}}, r.Copy.SourceUIDs)
				require.Equal(t, imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}}, r.Copy.DestUIDs)
				require.Len(t, inbox.Messages, 1)
				require.Equal(t, uint32(111), inbox.Messages[0].UId)
				require.Len(t, mb.Folders["foo"].Messages, 1)
			},
		},
		{
			name: "UID Copy",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				inbox := mb.Folders["INBOX"]
				inbox.Messages = append(inbox.Messages,
					&mail.Mail{UId: uint32(111)},
				)
				mb.AddFolder(&mail.Folder{
					Name: "foo",
				})

				r := &imaptest.MoveRecorder{}
				err = h.Move(
					&imap.IdSet{Ids: []imap.Set{imap.IdNum(111)}, IsUid: true},
					"foo",
					r,
					ctx)
				require.NoError(t, err)
				require.Greater(t, r.Copy.UIDValidity, uint32(0))
				require.Equal(t, imap.IdSet{Ids: []imap.Set{imap.IdNum(111)}, IsUid: true}, r.Copy.SourceUIDs)

				f := mb.Folders["foo"]
				require.Equal(t, imap.IdSet{Ids: []imap.Set{imap.IdNum(f.Messages[0].UId)}, IsUid: true}, r.Copy.DestUIDs)
				require.Len(t, r.Ids, 1)
			},
		},
		{
			name: "DElETE foo",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				mb.AddFolder(&mail.Folder{Name: "foo"})

				err = h.Delete("foo", ctx)
				require.NoError(t, err)

				f := mb.Folders["foo"]
				require.Nil(t, f)
			},
		},
		{
			name: "DElETE foo not found",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				err = h.Delete("foo", ctx)
				require.EqualError(t, err, "mailbox \"foo\" not found")
			},
		},
		{
			name: "DElETE foo but has children",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				mb.AddFolder(&mail.Folder{Name: "foo", Folders: map[string]*mail.Folder{"bar": {Name: "bar"}}})

				err = h.Delete("foo", ctx)
				require.EqualError(t, err, "name \"foo\" has inferior hierarchical names")
			},
		},
		{
			name: "DElETE foo/bar",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				mb.AddFolder(&mail.Folder{Name: "foo", Folders: map[string]*mail.Folder{"bar": {Name: "bar"}}})

				err = h.Delete("foo/bar", ctx)
				require.NoError(t, err)

				f := mb.Folders["foo"]
				require.Len(t, f.Folders, 0)
			},
		},
		{
			name: "DElETE INBOX",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				err = h.Delete("INBOX", ctx)
				require.EqualError(t, err, "INBOX cannot be deleted")
			},
		},
		{
			name: "RENAME foo to bar",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				mb.AddFolder(&mail.Folder{Name: "foo"})

				err = h.Rename("foo", "bar", ctx)
				require.NoError(t, err)

				f := mb.Folders["foo"]
				require.Nil(t, f)
				f = mb.Folders["bar"]
				require.NotNil(t, f)
				require.Equal(t, "bar", f.Name)
			},
		},
		{
			name: "RENAME move foo into bar",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				mb.AddFolder(&mail.Folder{Name: "foo"})
				mb.AddFolder(&mail.Folder{Name: "bar"})

				err = h.Rename("foo", "bar/foo", ctx)
				require.NoError(t, err)

				f := mb.Folders["foo"]
				require.Nil(t, f)
				f = mb.Folders["bar"].Folders["foo"]
				require.NotNil(t, f)
				require.Equal(t, "foo", f.Name)
			},
		},
		{
			name: "RENAME INBOX",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				mb.Folders["INBOX"].Messages = append(mb.Folders["INBOX"].Messages, &mail.Mail{})

				err = h.Rename("INBOX", "old", ctx)
				require.NoError(t, err)

				f := mb.Folders["INBOX"]
				require.NotNil(t, f)
				require.Equal(t, "INBOX", f.Name)
				require.Len(t, f.Messages, 0)
				f = mb.Folders["old"]
				require.NotNil(t, f)
				require.Equal(t, "old", f.Name)
				require.Len(t, f.Messages, 1)
			},
		},
		{
			name: "APPEND SENT",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
						Folders: map[string]*mail.FolderConfig{
							"Sent": {},
						},
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				mb.Folders["INBOX"].Messages = append(mb.Folders["INBOX"].Messages, &mail.Mail{})

				err = h.Append("Sent", &smtp.Message{MessageId: "1"}, imap.AppendOptions{}, ctx)
				require.NoError(t, err)

				f := mb.Select("Sent")
				require.Len(t, f.Messages, 1)
				require.Equal(t, "1", f.Messages[0].MessageId)
				require.Equal(t, []imap.Flag{imap.FlagRecent}, f.Messages[0].Flags)
			},
		},
		{
			name: "APPEND SENT with flags",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
						Folders: map[string]*mail.FolderConfig{
							"Sent": {},
						},
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				mb.Folders["INBOX"].Messages = append(mb.Folders["INBOX"].Messages, &mail.Mail{})

				err = h.Append("Sent", &smtp.Message{MessageId: "1"}, imap.AppendOptions{Flags: []imap.Flag{imap.FlagSeen}}, ctx)
				require.NoError(t, err)

				f := mb.Select("Sent")
				require.Len(t, f.Messages, 1)
				require.Equal(t, "1", f.Messages[0].MessageId)
				require.Equal(t, []imap.Flag{imap.FlagSeen}, f.Messages[0].Flags)
			},
		},
		{
			name: "IDLE append",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				done := make(chan struct{})
				r := &imaptest.UpdateRecorder{}
				err = h.Idle(r, done, ctx)
				require.NoError(t, err)

				err = h.Append("Inbox", &smtp.Message{MessageId: "1"}, imap.AppendOptions{Flags: []imap.Flag{imap.FlagSeen}}, ctx)
				require.NoError(t, err)

				require.Len(t, r.Messages, 1)
				require.Equal(t, []any{uint32(1)}, r.Messages[0])

				close(done)
				time.Sleep(time.Millisecond * 100)

				err = h.Append("Inbox", &smtp.Message{MessageId: "2"}, imap.AppendOptions{Flags: []imap.Flag{imap.FlagSeen}}, ctx)
				require.NoError(t, err)

				require.Len(t, r.Messages, 1)
			},
		},
		{
			name: "IDLE change message flag",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				err = h.Append("Inbox", &smtp.Message{MessageId: "1"}, imap.AppendOptions{Flags: []imap.Flag{imap.FlagSeen}}, ctx)
				require.NoError(t, err)

				done := make(chan struct{})
				r := &imaptest.UpdateRecorder{}
				err = h.Idle(r, done, ctx)
				require.NoError(t, err)

				err = h.Store(&imap.StoreRequest{
					Sequence: imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
					Action:   "add",
					Flags:    []imap.Flag{imap.FlagSeen},
				}, &imaptest.FetchRecorder{}, ctx)

				require.Len(t, r.Messages, 1)
				require.Equal(t, []any{uint32(1), []imap.Flag{imap.FlagSeen}}, r.Messages[0])

				close(done)
				time.Sleep(time.Millisecond * 100)

				err = h.Store(&imap.StoreRequest{
					Sequence: imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}},
					Action:   "add",
					Flags:    []imap.Flag{imap.FlagDeleted},
				}, &imaptest.FetchRecorder{}, ctx)

				require.Len(t, r.Messages, 1)
			},
		},
		{
			name: "IDLE expunge",
			cfg: &mail.Config{
				Mailboxes: map[string]*mail.MailboxConfig{
					"alice@mokapi.io": {
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", false, ctx)
				require.NoError(t, err)

				mb := s.Mailboxes["alice@mokapi.io"]
				mb.Folders["INBOX"].Messages = []*mail.Mail{
					{
						Message:  &smtp.Message{},
						UId:      1,
						Received: time.Time{},
					},
					{
						Message:  &smtp.Message{},
						UId:      2,
						Flags:    []imap.Flag{imap.FlagDeleted},
						Received: time.Time{},
					},
					{
						Message:  &smtp.Message{},
						UId:      3,
						Flags:    []imap.Flag{imap.FlagDeleted},
						Received: time.Time{},
					},
				}

				done := make(chan struct{})
				r := &imaptest.UpdateRecorder{}
				err = h.Idle(r, done, ctx)
				require.NoError(t, err)

				err = h.Expunge(&imap.IdSet{Ids: []imap.Set{imap.IdNum(2), imap.IdNum(3)}}, &imaptest.ExpungeRecorder{}, ctx)

				require.Len(t, r.Messages, 2)
				require.Equal(t, []any{uint32(2)}, r.Messages[0])
				require.Equal(t, []any{uint32(2)}, r.Messages[0])

				close(done)
				time.Sleep(time.Millisecond * 100)

				mb.Folders["INBOX"].Messages[0].Flags = []imap.Flag{imap.FlagDeleted}
				err = h.Expunge(&imap.IdSet{Ids: []imap.Set{imap.IdNum(1)}}, &imaptest.ExpungeRecorder{}, ctx)

				require.Len(t, r.Messages, 2)
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
