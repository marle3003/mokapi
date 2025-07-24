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
				{Rules: []mail.Rule{{Name: "foo", Sender: mail.NewRuleExpr(".*")}}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Len(t, result.Rules, 1)
				require.Equal(t, ".*", result.Rules[0].Sender.String())
			},
		},
		{
			name: "patch rule",
			configs: []*mail.Config{
				{Rules: []mail.Rule{{Name: "foo"}}},
				{Rules: []mail.Rule{{Name: "foo", Sender: mail.NewRuleExpr(".*")}}},
			},
			test: func(t *testing.T, result *mail.Config) {
				require.Len(t, result.Rules, 1)
				require.Equal(t, ".*", result.Rules[0].Sender.String())
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
