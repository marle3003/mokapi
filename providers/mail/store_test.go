package mail

import (
	"mokapi/imap"
	"mokapi/smtp"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustMailbox(t *testing.T, s *Store, name string) *Mailbox {
	t.Helper()
	mb, ok := s.Mailbox(name)
	require.True(t, ok, "mailbox %s", name)
	return mb
}

func newMessage(id string) *smtp.Message {
	return &smtp.Message{MessageId: id}
}

func TestStore_Update(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*Config
		test    func(s *Store, t *testing.T)
	}{
		{
			name: "add mailbox",
			configs: []*Config{
				{
					Mailboxes: map[string]*MailboxConfig{
						"foo@example.com": {},
					},
				},
				{
					Mailboxes: map[string]*MailboxConfig{
						"foo@example.com": {},
						"bar@example.com": {},
					},
				},
			},
			test: func(s *Store, t *testing.T) {
				require.Len(t, s.Mailboxes(), 2)
			},
		},
		{
			name: "update mailbox",
			configs: []*Config{
				{
					Mailboxes: map[string]*MailboxConfig{
						"foo@example.com": {},
					},
				},
				{
					Mailboxes: map[string]*MailboxConfig{
						"foo@example.com": {
							Username:    "foo",
							Password:    "secret",
							Description: "changed",
							Folders: map[string]*FolderConfig{
								"Trash": {},
							},
						},
					},
				},
			},
			test: func(s *Store, t *testing.T) {
				require.Len(t, s.Mailboxes(), 1)
				mb := mustMailbox(t, s, "foo@example.com")
				username, password := mb.Credentials()
				require.Equal(t, "foo", username)
				require.Equal(t, "secret", password)
				require.Equal(t, "changed", mb.Description())
				require.Contains(t, mb.Folders, "Trash")
			},
		},
		{
			name: "nil mailbox config",
			configs: []*Config{
				{Mailboxes: map[string]*MailboxConfig{"foo@example.com": nil}},
				{Mailboxes: map[string]*MailboxConfig{"foo@example.com": nil}},
			},
			test: func(s *Store, t *testing.T) {
				require.Len(t, s.Mailboxes(), 1)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := NewStore(tc.configs[0])
			for _, cfg := range tc.configs[1:] {
				s.Update(cfg)
			}
			tc.test(s, t)
		})
	}
}

func TestStore_Update_KeepsMailsAndFolders(t *testing.T) {
	cfg := &Config{Mailboxes: map[string]*MailboxConfig{
		"alice@foo.bar": {Folders: map[string]*FolderConfig{
			"INBOX": {Folders: map[string]*FolderConfig{"Sub": {Flags: []string{"old"}}}},
		}},
	}}
	s := NewStore(cfg)
	mb := mustMailbox(t, s, "alice@foo.bar")
	mb.Append(newMessage("inbox"))
	sub := mb.Folders["INBOX"].Folders["Sub"]
	sub.Append(newMessage("sub"))
	inbox := mb.Folders["INBOX"]

	s.Update(&Config{Mailboxes: map[string]*MailboxConfig{
		"alice@foo.bar": {Folders: map[string]*FolderConfig{
			"INBOX": {Folders: map[string]*FolderConfig{"Sub": {Flags: []string{"new"}}, "Other": {}}},
			"Trash": {},
		}},
	}})

	same := mustMailbox(t, s, "alice@foo.bar")
	require.Same(t, mb, same)
	require.Same(t, inbox, same.Folders["INBOX"])
	require.Same(t, sub, same.Folders["INBOX"].Folders["Sub"])
	require.Len(t, inbox.Messages, 1)
	require.Len(t, sub.Messages, 1)
	require.Equal(t, []imap.MailboxFlags{"new"}, sub.Flags)
	require.Contains(t, inbox.Folders, "Other")
	require.Contains(t, same.Folders, "Trash")
}

func TestStore_Update_Subscriptions(t *testing.T) {
	newSubscribedStore := func(t *testing.T) (*Store, *Folder) {
		s := NewStore(&Config{Mailboxes: map[string]*MailboxConfig{
			"alice@foo.bar": {Folders: map[string]*FolderConfig{"Archive": {Flags: []string{"old"}}}},
		}})
		f := mustMailbox(t, s, "alice@foo.bar").Folders["Archive"]
		f.AddFlag(imap.Subscribed)
		return s, f
	}

	t.Run("keeps subscription and replaces other flags", func(t *testing.T) {
		s, f := newSubscribedStore(t)

		s.Update(&Config{Mailboxes: map[string]*MailboxConfig{
			"alice@foo.bar": {Folders: map[string]*FolderConfig{"Archive": {Flags: []string{"new"}}}},
		}})

		require.ElementsMatch(t, []imap.MailboxFlags{"new", imap.Subscribed}, f.Flags)
	})

	t.Run("subscribed flag in file is not duplicated", func(t *testing.T) {
		s, f := newSubscribedStore(t)

		s.Update(&Config{Mailboxes: map[string]*MailboxConfig{
			"alice@foo.bar": {Folders: map[string]*FolderConfig{"Archive": {Flags: []string{`\Subscribed`}}}},
		}})

		require.Equal(t, []imap.MailboxFlags{imap.Subscribed}, f.Flags)
	})

	t.Run("keeps subscription of nested folder", func(t *testing.T) {
		cfg := func() *Config {
			return &Config{Mailboxes: map[string]*MailboxConfig{
				"alice@foo.bar": {Folders: map[string]*FolderConfig{
					"Archive": {Folders: map[string]*FolderConfig{"2025": {}}},
				}},
			}}
		}
		s := NewStore(cfg())
		f := mustMailbox(t, s, "alice@foo.bar").Folders["Archive"].Folders["2025"]
		f.AddFlag(imap.Subscribed)

		s.Update(cfg())

		require.Equal(t, []imap.MailboxFlags{imap.Subscribed}, f.Flags)
	})
}

func TestStore_Update_FolderNotInConfigStays(t *testing.T) {
	s := NewStore(&Config{Mailboxes: map[string]*MailboxConfig{
		"alice@foo.bar": {Folders: map[string]*FolderConfig{"INBOX": {}, "Archive": {}}},
	}})
	mb := mustMailbox(t, s, "alice@foo.bar")
	mb.Folders["Archive"].Append(newMessage("archived"))

	s.Update(&Config{Mailboxes: map[string]*MailboxConfig{"alice@foo.bar": {Folders: map[string]*FolderConfig{"INBOX": {}}}}})

	require.Len(t, mustMailbox(t, s, "alice@foo.bar").Folders["Archive"].Messages, 1)
}

func TestStore_Update_Mailboxes(t *testing.T) {
	configured := &Config{Mailboxes: map[string]*MailboxConfig{"alice@foo.bar": {}, "bob@foo.bar": {}}}

	t.Run("removes only configured mailboxes", func(t *testing.T) {
		s := NewStore(configured)
		require.NoError(t, s.EnsureMailbox("auto@foo.bar", nil))

		s.Update(&Config{Mailboxes: map[string]*MailboxConfig{"alice@foo.bar": {}}})

		_, ok := s.Mailbox("bob@foo.bar")
		require.False(t, ok)
		_, ok = s.Mailbox("alice@foo.bar")
		require.True(t, ok)
		_, ok = s.Mailbox("auto@foo.bar")
		require.True(t, ok)
	})

	t.Run("adopts auto-created mailbox that appears in the config", func(t *testing.T) {
		s := NewStore(&Config{})
		require.NoError(t, s.EnsureMailbox("alice@foo.bar", nil))
		mb := mustMailbox(t, s, "alice@foo.bar")
		mb.Append(newMessage("auto"))

		s.Update(&Config{Mailboxes: map[string]*MailboxConfig{"alice@foo.bar": {Password: "secret"}}})
		require.Same(t, mb, mustMailbox(t, s, "alice@foo.bar"))
		require.Len(t, mb.Folders["INBOX"].Messages, 1)

		s.Update(&Config{})
		_, ok := s.Mailbox("alice@foo.bar")
		require.False(t, ok, "adopted mailbox follows the file")
	})

	t.Run("mailbox created by Update is not auto-created", func(t *testing.T) {
		s := NewStore(&Config{})
		s.Update(configured)
		s.Update(&Config{})
		require.Empty(t, s.Mailboxes())
	})

	t.Run("sorted by name", func(t *testing.T) {
		s := NewStore(&Config{Mailboxes: map[string]*MailboxConfig{"c": {}, "a": {}, "b": {}}})
		var names []string
		for _, mb := range s.Mailboxes() {
			names = append(names, mb.Name)
		}
		require.Equal(t, []string{"a", "b", "c"}, names)
	})
}

func TestStore_Update_MaxInboxMails(t *testing.T) {
	newConfig := func(max int) *Config {
		return &Config{
			Settings:  &Settings{MaxInboxMails: max},
			Mailboxes: map[string]*MailboxConfig{"alice@foo.bar": {Folders: map[string]*FolderConfig{"INBOX": {}}}},
		}
	}

	t.Run("config built INBOX has the limit", func(t *testing.T) {
		s := NewStore(newConfig(2))
		mb := mustMailbox(t, s, "alice@foo.bar")
		for i := 0; i < 5; i++ {
			mb.Append(newMessage("m"))
		}
		require.Len(t, mb.Folders["INBOX"].Messages, 2)
	})

	t.Run("lowered limit trims on the next mail", func(t *testing.T) {
		s := NewStore(newConfig(10))
		mb := mustMailbox(t, s, "alice@foo.bar")
		for i := 0; i < 8; i++ {
			mb.Append(newMessage("m"))
		}
		require.Len(t, mb.Folders["INBOX"].Messages, 8)

		s.Update(newConfig(3))
		require.Len(t, mb.Folders["INBOX"].Messages, 8, "takes effect with the next mail")

		mb.Append(newMessage("m"))
		require.Len(t, mb.Folders["INBOX"].Messages, 3)
	})

	t.Run("raised limit keeps more mails", func(t *testing.T) {
		s := NewStore(newConfig(1))
		mb := mustMailbox(t, s, "alice@foo.bar")
		mb.Append(newMessage("m"))

		s.Update(newConfig(3))
		mb.Append(newMessage("m"))
		mb.Append(newMessage("m"))
		require.Len(t, mb.Folders["INBOX"].Messages, 3)
	})

	t.Run("INBOX created later uses the limit", func(t *testing.T) {
		s := NewStore(&Config{Settings: &Settings{MaxInboxMails: 5}, Mailboxes: map[string]*MailboxConfig{"alice@foo.bar": {}}})
		s.Update(&Config{Settings: &Settings{MaxInboxMails: 1}, Mailboxes: map[string]*MailboxConfig{"alice@foo.bar": {}}})
		mb := mustMailbox(t, s, "alice@foo.bar")
		mb.Append(newMessage("m"))
		mb.Append(newMessage("m"))
		require.Len(t, mb.Folders["INBOX"].Messages, 1)
	})
}

func TestStore_EnsureMailbox(t *testing.T) {
	t.Run("not allowed", func(t *testing.T) {
		s := NewStore(&Config{})
		require.EqualError(t, s.EnsureMailbox("foo@bar", &Settings{AutoCreateMailbox: false}), "mailbox can not be created")
		_, ok := s.Mailbox("foo@bar")
		require.False(t, ok)
	})

	t.Run("existing mailbox does not need the setting", func(t *testing.T) {
		s := NewStore(&Config{Mailboxes: map[string]*MailboxConfig{"foo@bar": {}}})
		require.NoError(t, s.EnsureMailbox("foo@bar", &Settings{AutoCreateMailbox: false}))
	})

	t.Run("default limit without settings", func(t *testing.T) {
		s := NewStore(&Config{})
		require.NoError(t, s.EnsureMailbox("foo@bar", nil))
		mb := mustMailbox(t, s, "foo@bar")
		for i := 0; i < DefaultMaxInboxMails+5; i++ {
			mb.Append(newMessage("m"))
		}
		require.Len(t, mb.Folders["INBOX"].Messages, DefaultMaxInboxMails)
	})
}

func TestStore_ConfigBuiltFolderCanAddFolder(t *testing.T) {
	s := NewStore(&Config{Mailboxes: map[string]*MailboxConfig{"alice@foo.bar": {Folders: map[string]*FolderConfig{"INBOX": {}}}}})
	mb := mustMailbox(t, s, "alice@foo.bar")

	require.NotPanics(t, func() {
		mb.Folders["INBOX"].AddFolder(&Folder{Name: "Sub"})
	})
	require.Contains(t, mb.Folders["INBOX"].Folders, "Sub")
}

func TestStore_GetMail(t *testing.T) {
	s := NewStore(&Config{Mailboxes: map[string]*MailboxConfig{"alice@foo.bar": {Folders: map[string]*FolderConfig{"INBOX": {}, "Sub": {Folders: map[string]*FolderConfig{"Deep": {}}}}}}})
	mb := mustMailbox(t, s, "alice@foo.bar")
	mb.Append(newMessage("inbox"))
	mb.Folders["Sub"].Folders["Deep"].Append(newMessage("deep"))

	require.Equal(t, "inbox", s.GetMail("inbox").MessageId)
	require.Equal(t, "deep", s.GetMail("deep").MessageId)
	require.Nil(t, s.GetMail("unknown"))
}

func TestStore_Nil(t *testing.T) {
	var s *Store
	mb, ok := s.Mailbox("foo")
	require.Nil(t, mb)
	require.False(t, ok)
	require.Empty(t, s.Mailboxes())
	require.Nil(t, s.GetMail("id"))
}

func TestMailbox_Accessors(t *testing.T) {
	s := NewStore(&Config{Mailboxes: map[string]*MailboxConfig{
		"alice@foo.bar": {Username: "alice", Password: "secret", Description: "Alice", Folders: map[string]*FolderConfig{"INBOX": {}, "Archive": {}}},
	}})
	mb := mustMailbox(t, s, "alice@foo.bar")
	username, password := mb.Credentials()
	require.Equal(t, "alice", username)
	require.Equal(t, "secret", password)
	require.Equal(t, "Alice", mb.Description())
	require.Equal(t, []string{"Archive", "INBOX"}, mb.FolderNames())

	mb.Append(newMessage("inbox"))
	mb.Folders["Archive"].Append(newMessage("archive"))
	require.Len(t, mb.Messages(""), 2)
	require.Empty(t, mb.Messages("INBOX"), "messages of the sub folders of INBOX")
}
