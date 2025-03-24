package mail_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/mail"
	"mokapi/engine/enginetest"
	"mokapi/imap"
	"mokapi/imap/imaptest"
	"testing"
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				r, err := h.Select("Inbox", ctx)
				require.NoError(t, err)
				require.Len(t, r.Flags, 5)
			},
		},
		{
			name: "Select Inbox/foo",
			cfg: &mail.Config{
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
						Folders: []mail.FolderConfig{
							{
								Name: "inbox",
								Folders: []mail.FolderConfig{
									{Name: "foo"},
								},
							},
						},
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				r, err := h.Select("Inbox/foo", ctx)
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
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
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
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
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
			name: "Status Inbox",
			cfg: &mail.Config{
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
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
			name: "List inbox should be first",
			cfg: &mail.Config{
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
						Folders: []mail.FolderConfig{
							{Name: "ABC"},
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
						Folders: []mail.FolderConfig{
							{
								Name:  "foo",
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
			name: "List Archive/* foo",
			cfg: &mail.Config{
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
						Folders: []mail.FolderConfig{
							{
								Name: "Archive",
								Folders: []mail.FolderConfig{
									{
										Name: "2025",
										Folders: []mail.FolderConfig{
											{
												Name: "foo",
											},
										},
									},
									{
										Name: "2026",
										Folders: []mail.FolderConfig{
											{
												Name: "bar",
											},
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
				r, err := h.List("Archive/*", "foo", nil, ctx)
				require.NoError(t, err)
				require.Len(t, r, 1)
				require.Equal(t, "foo", r[0].Name)
			},
		},
		{
			name: "List Archive %",
			cfg: &mail.Config{
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
						Folders: []mail.FolderConfig{
							{
								Name: "Archive",
								Folders: []mail.FolderConfig{
									{
										Name: "2025",
										Folders: []mail.FolderConfig{
											{
												Name: "foo",
											},
										},
									},
									{
										Name: "2026",
										Folders: []mail.FolderConfig{
											{
												Name: "bar",
											},
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
					{Delimiter: "/", Flags: nil, Name: "2025"},
					{Delimiter: "/", Flags: nil, Name: "2026"},
				}, r)
			},
		},
		{
			name: "List Archive/%",
			cfg: &mail.Config{
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
						Folders: []mail.FolderConfig{
							{
								Name: "Archive",
								Folders: []mail.FolderConfig{
									{
										Name: "2025",
										Folders: []mail.FolderConfig{
											{
												Name: "foo",
											},
										},
									},
									{
										Name: "2026",
										Folders: []mail.FolderConfig{
											{
												Name: "bar",
											},
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
					{Delimiter: "/", Flags: nil, Name: "2025"},
					{Delimiter: "/", Flags: nil, Name: "2026"},
				}, r)
			},
		},
		{
			name: "Add deleted flag to message",
			cfg: &mail.Config{
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", ctx)
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", ctx)
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", ctx)
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", ctx)
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", ctx)
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", ctx)
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", ctx)
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", ctx)
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", ctx)
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", ctx)
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", ctx)
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
				Mailboxes: []mail.MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "foo",
					},
				},
			},
			test: func(t *testing.T, h *mail.Handler, s *mail.Store, ctx context.Context) {
				_ = h.Login("alice", "foo", ctx)
				_, err := h.Select("Inbox", ctx)
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
