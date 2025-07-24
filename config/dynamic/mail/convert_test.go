package mail

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/mail"
	"mokapi/smtp"
	"regexp"
	"testing"
)

func TestConvert(t *testing.T) {
	testcases := []struct {
		name string
		cfg  *Config
		test func(*testing.T, *mail.Config)
	}{
		{
			name: "empty",
			cfg:  &Config{},
			test: func(t *testing.T, c *mail.Config) {
				require.NotNil(t, c)
				require.Equal(t, "1.0", c.Version)
			},
		},
		{
			name: "info",
			cfg:  &Config{Info: Info{Name: "name", Description: "description"}},
			test: func(t *testing.T, c *mail.Config) {
				require.NotNil(t, c)
				require.Equal(t, "name", c.Info.Name)
				require.Equal(t, "description", c.Info.Description)
			},
		},
		{
			name: "smtp server",
			cfg:  &Config{Server: "smtp://mokapi.io"},
			test: func(t *testing.T, c *mail.Config) {
				require.NotNil(t, c)
				require.Contains(t, c.Servers, "smtp://mokapi.io")
				server := c.Servers["smtp://mokapi.io"]
				require.Equal(t, "mokapi.io", server.Host)
				require.Equal(t, "smtp", server.Protocol)
			},
		},
		{
			name: "imap server",
			cfg:  &Config{Server: "imap://mokapi.io"},
			test: func(t *testing.T, c *mail.Config) {
				require.NotNil(t, c)
				require.Contains(t, c.Servers, "imap://mokapi.io")
				server := c.Servers["imap://mokapi.io"]
				require.Equal(t, "mokapi.io", server.Host)
				require.Equal(t, "imap", server.Protocol)
			},
		},
		{
			name: "smtp server using list",
			cfg: &Config{Servers: []Server{
				{Url: "smtp://mokapi.io", Description: "foobar"},
			}},
			test: func(t *testing.T, c *mail.Config) {
				require.NotNil(t, c)
				require.Contains(t, c.Servers, "smtp://mokapi.io")
				server := c.Servers["smtp://mokapi.io"]
				require.Equal(t, "mokapi.io", server.Host)
				require.Equal(t, "smtp", server.Protocol)
				require.Equal(t, "foobar", server.Description)
			},
		},
		{
			name: "imap and smtp server using list",
			cfg: &Config{Servers: []Server{
				{Url: "imap://mokapi.io", Description: "foobar"},
				{Url: "smtp://mokapi.io", Description: "yuh"},
			}},
			test: func(t *testing.T, c *mail.Config) {
				require.NotNil(t, c)
				require.Len(t, c.Servers, 2)
				require.Contains(t, c.Servers, "imap://mokapi.io")
				server := c.Servers["smtp://mokapi.io"]
				require.Equal(t, "mokapi.io", server.Host)
				require.Equal(t, "smtp", server.Protocol)
				require.Equal(t, "yuh", server.Description)

				server = c.Servers["imap://mokapi.io"]
				require.Equal(t, "mokapi.io", server.Host)
				require.Equal(t, "imap", server.Protocol)
				require.Equal(t, "foobar", server.Description)
			},
		},
		{
			name: "mailbox",
			cfg: &Config{
				Mailboxes: []MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "password",
						Folders:  nil,
					},
				},
			},
			test: func(t *testing.T, c *mail.Config) {
				require.NotNil(t, c)
				require.Len(t, c.Mailboxes, 1)
				mb := c.Mailboxes["alice@mokapi.io"]
				require.Equal(t, "alice", mb.Username)
				require.Equal(t, "password", mb.Password)
				require.Len(t, mb.Folders, 0)
			},
		},
		{
			name: "mailbox with folder",
			cfg: &Config{
				Mailboxes: []MailboxConfig{
					{
						Name:     "alice@mokapi.io",
						Username: "alice",
						Password: "password",
						Folders: []FolderConfig{
							{
								Name:  "folder",
								Flags: []string{"\\foo"},
							},
						},
					},
				},
			},
			test: func(t *testing.T, c *mail.Config) {
				require.NotNil(t, c)
				require.Len(t, c.Mailboxes, 1)
				mb := c.Mailboxes["alice@mokapi.io"]
				require.Equal(t, "alice", mb.Username)
				require.Equal(t, "password", mb.Password)
				require.Len(t, mb.Folders, 1)
				require.Equal(t, []string{"\\foo"}, mb.Folders["folder"].Flags)
				require.Len(t, mb.Folders["folder"].Folders, 0)
			},
		},
		{
			name: "rules",
			cfg: &Config{
				Rules: []Rule{
					{
						Name:      "foo",
						Sender:    &RuleExpr{expr: regexp.MustCompile("alice@mokapi.io")},
						Recipient: &RuleExpr{expr: regexp.MustCompile("bob@mokapi.io")},
						Subject:   &RuleExpr{expr: regexp.MustCompile("Hello.*")},
						Body:      &RuleExpr{expr: regexp.MustCompile(".*")},
						Action:    RuleAction("deny"),
						RejectResponse: &RejectResponse{
							StatusCode:         550,
							EnhancedStatusCode: smtp.EnhancedStatusCode{5, 7, 1},
							Text:               "foobar",
						},
					},
				},
			},
			test: func(t *testing.T, c *mail.Config) {
				require.NotNil(t, c)
				require.Len(t, c.Rules, 1)
				r := c.Rules[0]
				require.Equal(t, "foo", r.Name)
				require.Equal(t, mail.NewRuleExpr("alice@mokapi.io"), r.Sender)
				require.Equal(t, mail.NewRuleExpr("bob@mokapi.io"), r.Recipient)
				require.Equal(t, mail.NewRuleExpr("Hello.*"), r.Subject)
				require.Equal(t, mail.NewRuleExpr(".*"), r.Body)
				require.Equal(t, mail.RuleAction("deny"), r.Action)
				require.Equal(t, &mail.RejectResponse{
					StatusCode:         550,
					EnhancedStatusCode: smtp.EnhancedStatusCode{5, 7, 1},
					Text:               "foobar",
				}, r.RejectResponse)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.test(t, tc.cfg.Convert())
		})
	}
}
