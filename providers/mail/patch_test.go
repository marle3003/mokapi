package mail_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/mail"
	"testing"
)

func TestConfig_Patch(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*mail.Config
		test    func(t *testing.T, result *mail.Config)
	}{
		{
			name: "description",
			configs: []*mail.Config{
				{},
				{Info: mail.Info{
					Description: "foo",
				}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Equal(t, "foo", result.Info.Description)
			},
		},
		{
			name: "autoCreateMailbox",
			configs: []*mail.Config{
				{Settings: &mail.Settings{AutoCreateMailbox: true}},
				{},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.True(t, result.Settings.AutoCreateMailbox)
			},
		},
		{
			name: "autoCreateMailbox",
			configs: []*mail.Config{
				{Settings: &mail.Settings{AutoCreateMailbox: true}},
				{Settings: &mail.Settings{AutoCreateMailbox: false}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.False(t, result.Settings.AutoCreateMailbox)
			},
		},
		{
			name: "add rule",
			configs: []*mail.Config{
				{},
				{Rules: map[string]*mail.Rule{"foo": {Sender: mail.NewRuleExpr(".*")}}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Len(t, result.Rules, 1)
				require.Equal(t, ".*", result.Rules["foo"].Sender.String())
			},
		},
		{
			name: "patch rule",
			configs: []*mail.Config{
				{Rules: map[string]*mail.Rule{"foo": {}}},
				{Rules: map[string]*mail.Rule{"foo": {Sender: mail.NewRuleExpr(".*")}}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Len(t, result.Rules, 1)
				require.Equal(t, ".*", result.Rules["foo"].Sender.String())
			},
		},
		{
			name: "update mailbox",
			configs: []*mail.Config{
				{
					Mailboxes: map[string]*mail.MailboxConfig{
						"foo@example.com": {},
					},
				},
				{
					Mailboxes: map[string]*mail.MailboxConfig{
						"foo@example.com": {
							Password: "secret",
							Folders: map[string]*mail.FolderConfig{
								"foo": {},
							},
						},
					},
				},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Len(t, result.Mailboxes, 1)
				require.Equal(t, "secret", result.Mailboxes["foo@example.com"].Password)
				require.Equal(t, map[string]*mail.FolderConfig{
					"foo": {},
				}, result.Mailboxes["foo@example.com"].Folders)
			},
		},
		{
			name: "update mailbox folder",
			configs: []*mail.Config{
				{
					Mailboxes: map[string]*mail.MailboxConfig{
						"foo@example.com": {
							Folders: map[string]*mail.FolderConfig{
								"foo": {Flags: []string{"foo"}},
							},
						},
					},
				},
				{
					Mailboxes: map[string]*mail.MailboxConfig{
						"foo@example.com": {
							Password: "secret",
							Folders: map[string]*mail.FolderConfig{
								"foo": {Flags: []string{"foo", "bar"}},
							},
						},
					},
				},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Len(t, result.Mailboxes, 1)
				require.Equal(t, "secret", result.Mailboxes["foo@example.com"].Password)
				require.Equal(t, map[string]*mail.FolderConfig{
					"foo": {Flags: []string{"foo", "bar"}},
				}, result.Mailboxes["foo@example.com"].Folders)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				c.Patch(p)
			}
			tc.test(t, c)
		})
	}
}

func TestConfig_Patch_DoesNotAliasPatch(t *testing.T) {
	newPatch := func() *mail.Config {
		return &mail.Config{
			Info:     mail.Info{Contact: &mail.Contact{Name: "Alice"}},
			Servers:  map[string]*mail.Server{"smtp": {Host: ":25", Protocol: "smtp"}},
			Settings: &mail.Settings{MaxRecipients: 5},
			Mailboxes: map[string]*mail.MailboxConfig{
				"alice@foo.bar": {Folders: map[string]*mail.FolderConfig{"INBOX": {Flags: []string{"a"}}}},
			},
			Rules: map[string]*mail.Rule{"r": {Sender: mail.NewRuleExpr(".*"), Action: mail.Deny}},
		}
	}
	modify := func(c *mail.Config) {
		c.Info.Contact.Name = "changed"
		for _, s := range c.Servers {
			s.Host = "changed"
		}
		c.Settings.MaxRecipients = 99
		for _, mb := range c.Mailboxes {
			mb.Username = "changed"
			for _, f := range mb.Folders {
				if len(f.Flags) > 0 {
					f.Flags[0] = "changed"
				}
				f.Folders = map[string]*mail.FolderConfig{"New": {}}
			}
		}
		for _, r := range c.Rules {
			r.Action = mail.Allow
		}
	}

	t.Run("into an empty config", func(t *testing.T) {
		p := newPatch()
		c := &mail.Config{}
		c.Patch(p)
		modify(c)
		require.Equal(t, newPatch(), p)
	})

	t.Run("into a config with other entries", func(t *testing.T) {
		p := newPatch()
		c := &mail.Config{
			Info:      mail.Info{Contact: &mail.Contact{}},
			Servers:   map[string]*mail.Server{"imap": {}},
			Settings:  &mail.Settings{},
			Mailboxes: map[string]*mail.MailboxConfig{"bob@foo.bar": {}},
			Rules:     map[string]*mail.Rule{"other": {}},
		}
		c.Patch(p)
		modify(c)
		require.Equal(t, newPatch(), p)
	})

	t.Run("new folder in an existing mailbox", func(t *testing.T) {
		p := newPatch()
		c := &mail.Config{Mailboxes: map[string]*mail.MailboxConfig{"alice@foo.bar": {Folders: map[string]*mail.FolderConfig{"Trash": {}}}}}
		c.Patch(p)
		modify(c)
		require.Equal(t, newPatch(), p)
	})

	t.Run("new folder in an existing mailbox without folders", func(t *testing.T) {
		p := newPatch()
		c := &mail.Config{Mailboxes: map[string]*mail.MailboxConfig{"alice@foo.bar": {}}}
		c.Patch(p)
		modify(c)
		require.Equal(t, newPatch(), p)
	})
}

func TestConfig_Patch_ParsedConfigStaysUnchanged(t *testing.T) {
	first := newCloneFixture()
	second := &mail.Config{
		Servers:  map[string]*mail.Server{"smtp": {Host: ":2525"}, "imap": {Host: ":143", Protocol: "imap"}},
		Settings: &mail.Settings{MaxRecipients: 1},
		Mailboxes: map[string]*mail.MailboxConfig{
			"alice@foo.bar": {Password: "secret", Folders: map[string]*mail.FolderConfig{"INBOX": {Flags: []string{"x"}, Folders: map[string]*mail.FolderConfig{"Sub": {Flags: []string{"y"}}}}}},
		},
		Rules: map[string]*mail.Rule{"deny": {Sender: mail.NewRuleExpr("a")}},
	}
	secondBefore := second.Clone()

	merged := first.Clone()
	merged.Patch(second)

	require.Equal(t, newCloneFixture(), first)
	require.Equal(t, secondBefore, second)
	require.Equal(t, ":2525", merged.Servers["smtp"].Host)
	require.Equal(t, ":25", first.Servers["smtp"].Host)
}

func TestConfig_Patch_Servers(t *testing.T) {
	c := &mail.Config{Servers: map[string]*mail.Server{"smtp": {Host: ":25", Protocol: "smtp", Description: "first"}}}
	c.Patch(&mail.Config{Servers: map[string]*mail.Server{
		"smtp": {Host: ":2525"},
		"imap": {Host: ":143", Protocol: "imap"},
	}})

	require.Equal(t, map[string]*mail.Server{
		"smtp": {Host: ":2525", Protocol: "smtp", Description: "first"},
		"imap": {Host: ":143", Protocol: "imap"},
	}, c.Servers)
}

func TestConfig_Patch_Fields(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*mail.Config
		test    func(t *testing.T, result *mail.Config)
	}{
		{
			name: "maxInboxMails",
			configs: []*mail.Config{
				{Settings: &mail.Settings{MaxInboxMails: 10}},
				{Settings: &mail.Settings{MaxInboxMails: 3}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Equal(t, 3, result.Settings.MaxInboxMails)
			},
		},
		{
			name: "allowUnknownSenders",
			configs: []*mail.Config{
				{Settings: &mail.Settings{AllowUnknownSenders: true}},
				{Settings: &mail.Settings{AllowUnknownSenders: false}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.False(t, result.Settings.AllowUnknownSenders)
			},
		},
		{
			name: "version",
			configs: []*mail.Config{
				{Info: mail.Info{Version: "1.0"}},
				{Info: mail.Info{Version: "2.0"}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Equal(t, "2.0", result.Info.Version)
			},
		},
		{
			name: "empty version keeps first",
			configs: []*mail.Config{
				{Info: mail.Info{Version: "1.0"}},
				{},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Equal(t, "1.0", result.Info.Version)
			},
		},
		{
			name: "contact",
			configs: []*mail.Config{
				{Info: mail.Info{Contact: &mail.Contact{Name: "first"}}},
				{Info: mail.Info{Contact: &mail.Contact{Name: "second"}}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Equal(t, "second", result.Info.Contact.Name)
			},
		},
		{
			name: "no contact keeps first",
			configs: []*mail.Config{
				{Info: mail.Info{Contact: &mail.Contact{Name: "first"}}},
				{},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Equal(t, "first", result.Info.Contact.Name)
			},
		},
		{
			name: "mailbox description",
			configs: []*mail.Config{
				{Mailboxes: map[string]*mail.MailboxConfig{"foo@example.com": {Description: "first"}}},
				{Mailboxes: map[string]*mail.MailboxConfig{"foo@example.com": {Description: "second"}}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Equal(t, "second", result.Mailboxes["foo@example.com"].Description)
			},
		},
		{
			name: "mailbox without folders gets folder of second file",
			configs: []*mail.Config{
				{Mailboxes: map[string]*mail.MailboxConfig{"foo@example.com": {}}},
				{Mailboxes: map[string]*mail.MailboxConfig{"foo@example.com": {Folders: map[string]*mail.FolderConfig{"Trash": {}}}}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Contains(t, result.Mailboxes["foo@example.com"].Folders, "Trash")
			},
		},
		{
			name: "nested folder without folder map",
			configs: []*mail.Config{
				{Mailboxes: map[string]*mail.MailboxConfig{"foo@example.com": {Folders: map[string]*mail.FolderConfig{"INBOX": {}}}}},
				{Mailboxes: map[string]*mail.MailboxConfig{"foo@example.com": {Folders: map[string]*mail.FolderConfig{"INBOX": {Folders: map[string]*mail.FolderConfig{"Sub": {}}}}}}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Contains(t, result.Mailboxes["foo@example.com"].Folders["INBOX"].Folders, "Sub")
			},
		},
		{
			name: "rule action",
			configs: []*mail.Config{
				{Rules: map[string]*mail.Rule{"r": {Action: mail.Allow}}},
				{Rules: map[string]*mail.Rule{"r": {Action: mail.Deny}}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Equal(t, mail.Deny, result.Rules["r"].Action)
			},
		},
		{
			name: "rule description",
			configs: []*mail.Config{
				{Rules: map[string]*mail.Rule{"r": {Description: "foo"}}},
				{Rules: map[string]*mail.Rule{"r": {Description: "bar"}}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Equal(t, "bar", result.Rules["r"].Description)
			},
		},
		{
			name: "rule without action and description keeps original",
			configs: []*mail.Config{
				{Rules: map[string]*mail.Rule{"r": {Action: mail.Deny, Description: "foo"}}},
				{Rules: map[string]*mail.Rule{"r": {Sender: mail.NewRuleExpr(".*")}}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Equal(t, mail.Deny, result.Rules["r"].Action)
				require.Equal(t, "foo", result.Rules["r"].Description)
			},
		},
		{
			name: "nil entries are skipped",
			configs: []*mail.Config{
				{
					Mailboxes: map[string]*mail.MailboxConfig{"foo@example.com": {Folders: map[string]*mail.FolderConfig{"INBOX": {}}}},
					Rules:     map[string]*mail.Rule{"r": {}},
				},
				{
					Mailboxes: map[string]*mail.MailboxConfig{"foo@example.com": nil},
					Rules:     map[string]*mail.Rule{"r": nil},
				},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Contains(t, result.Mailboxes["foo@example.com"].Folders, "INBOX")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				c.Patch(p)
			}
			tc.test(t, c)
		})
	}
}
