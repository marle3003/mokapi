package mail_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/mail"
	"mokapi/smtp"
	"testing"
)

func TestParseConfig_Yaml(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, c *mail.Config)
	}{
		{
			name: "empty",
			data: "",
			test: func(t *testing.T, c *mail.Config) {
				require.NotNil(t, c)
				require.Empty(t, c.Version)
			},
		},
		{
			name: "version",
			data: "mail: 1.0",
			test: func(t *testing.T, c *mail.Config) {
				require.Equal(t, "1.0", c.Version)
				require.True(t, c.Settings.AutoCreateMailbox)
			},
		},
		{
			name: "mailbox",
			data: `
mail: 1.0
mailboxes:
  alice@mokapi.io:
    username: alice
    password: mokapi
`,
			test: func(t *testing.T, c *mail.Config) {
				require.Contains(t, c.Mailboxes, "alice@mokapi.io")
				require.Equal(t, "alice", c.Mailboxes["alice@mokapi.io"].Username)
				require.Equal(t, "mokapi", c.Mailboxes["alice@mokapi.io"].Password)
				require.Len(t, c.Mailboxes["alice@mokapi.io"].Folders, 0)
			},
		},
		{
			name: "mailbox with folder",
			data: `
mail: 1.0
mailboxes:
  alice@mokapi.io:
    username: alice
    password: mokapi
    folders: 
      inbox:
        flags: [\HasNoChildren]
`,
			test: func(t *testing.T, c *mail.Config) {
				require.Contains(t, c.Mailboxes, "alice@mokapi.io")
				require.Equal(t, "alice", c.Mailboxes["alice@mokapi.io"].Username)
				require.Equal(t, "mokapi", c.Mailboxes["alice@mokapi.io"].Password)
				require.Contains(t, c.Mailboxes["alice@mokapi.io"].Folders, "inbox")
				require.Equal(t, []string{"\\HasNoChildren"}, c.Mailboxes["alice@mokapi.io"].Folders["inbox"].Flags)
				require.Len(t, c.Mailboxes["alice@mokapi.io"].Folders["inbox"].Folders, 0)
			},
		},
		{
			name: "server on all ip addresses",
			data: `
mail: 1.0
servers:
  smtp:
    host: :1234
    protocol: smtp
`,
			test: func(t *testing.T, c *mail.Config) {
				require.Len(t, c.Servers, 1)
				require.Equal(t, ":1234", c.Servers["smtp"].Host)
				require.Equal(t, "smtp", c.Servers["smtp"].Protocol)
			},
		},
		{
			name: "server with one rule",
			data: `
mail: 1.0
servers: 
  smtps:
    host: localhost:8025
    protocol: smtps
rules:
  sender:
    sender: .*@foo.bar
    action: allow
`,
			test: func(t *testing.T, c *mail.Config) {
				require.Len(t, c.Servers, 1)
				require.Equal(t, "localhost:8025", c.Servers["smtps"].Host)
				require.Equal(t, "smtps", c.Servers["smtps"].Protocol)
				require.Len(t, c.Rules, 1)
				require.Equal(t, ".*@foo.bar", c.Rules["sender"].Sender.String())
				require.Equal(t, mail.Allow, c.Rules["sender"].Action)
			},
		},
		{
			name: "custom response rule",
			data: `
smtp: 1.0
rules:
  recipients:
    recipient: .*@foo.bar
    action: deny
    rejectResponse:
      statusCode: 550
      enhancedStatusCode: 5.7.1
      message: foobar
`,
			test: func(t *testing.T, c *mail.Config) {
				require.Len(t, c.Rules, 1)
				require.Equal(t, ".*@foo.bar", c.Rules["recipients"].Recipient.String())
				require.Equal(t, mail.Deny, c.Rules["recipients"].Action)
				require.NotNil(t, c.Rules["recipients"].RejectResponse)
				require.Equal(t, smtp.StatusCode(550), c.Rules["recipients"].RejectResponse.StatusCode)
				require.Equal(t, smtp.EnhancedStatusCode{5, 7, 1}, c.Rules["recipients"].RejectResponse.EnhancedStatusCode)
				require.Equal(t, "foobar", c.Rules["recipients"].RejectResponse.Message)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := &mail.Config{}
			err := yaml.Unmarshal([]byte(tc.data), &c)
			require.NoError(t, err)
			tc.test(t, c)
		})
	}
}

func TestParseConfig_Json(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, c *mail.Config)
	}{
		{
			name: "version",
			data: `{"mail": "1.0"}`,
			test: func(t *testing.T, c *mail.Config) {
				require.Equal(t, "1.0", c.Version)
				require.True(t, c.Settings.AutoCreateMailbox)
			},
		},
		{
			name: "mailbox",
			data: `{
"mail": "1.0",
"mailboxes": {
  "alice@mokapi.io": {
    "username": "alice",
    "password": "mokapi"
  }
}}`,
			test: func(t *testing.T, c *mail.Config) {
				require.Contains(t, c.Mailboxes, "alice@mokapi.io")
				require.Equal(t, "alice", c.Mailboxes["alice@mokapi.io"].Username)
				require.Equal(t, "mokapi", c.Mailboxes["alice@mokapi.io"].Password)
				require.Len(t, c.Mailboxes["alice@mokapi.io"].Folders, 0)
			},
		},
		{
			name: "mailbox with folder",
			data: `{
"mail": "1.0",
"mailboxes": {
  "alice@mokapi.io": {
    "username": "alice",
    "password": "mokapi",
    "folders": {
      "inbox": {
        "flags": ["\\HasNoChildren"]
      }
    }
  }
}}`,
			test: func(t *testing.T, c *mail.Config) {
				require.Contains(t, c.Mailboxes, "alice@mokapi.io")
				require.Equal(t, "alice", c.Mailboxes["alice@mokapi.io"].Username)
				require.Equal(t, "mokapi", c.Mailboxes["alice@mokapi.io"].Password)
				require.Contains(t, c.Mailboxes["alice@mokapi.io"].Folders, "inbox")
				require.Equal(t, []string{"\\HasNoChildren"}, c.Mailboxes["alice@mokapi.io"].Folders["inbox"].Flags)
				require.Len(t, c.Mailboxes["alice@mokapi.io"].Folders["inbox"].Folders, 0)
			},
		},
		{
			name: "server with one rule",
			data: `{
"mail": "1.0",
"servers": {
  "smtps": {
    "host": "localhost:8025",
    "protocol": "smtps"
  }
},
"rules": {
  "sender": {
    "sender": ".*@foo.bar",
    "action": "allow"
  }}
}`,
			test: func(t *testing.T, c *mail.Config) {
				require.Len(t, c.Servers, 1)
				require.Equal(t, "localhost:8025", c.Servers["smtps"].Host)
				require.Equal(t, "smtps", c.Servers["smtps"].Protocol)
				require.Len(t, c.Rules, 1)
				require.Equal(t, ".*@foo.bar", c.Rules["sender"].Sender.String())
				require.Equal(t, mail.Allow, c.Rules["sender"].Action)
			},
		},
		{
			name: "custom response rule",
			data: `{
"smtp": "1.0",
"rules": {
  "recipients": { 
    "recipient": ".*@foo.bar",
    "action": "deny",
    "rejectResponse": {
      "statusCode": 550,
      "enhancedStatusCode": "5.7.1",
      "message": "foobar"
    }
  }}
}`,
			test: func(t *testing.T, c *mail.Config) {
				require.Len(t, c.Rules, 1)
				require.Equal(t, ".*@foo.bar", c.Rules["recipients"].Recipient.String())
				require.Equal(t, mail.Deny, c.Rules["recipients"].Action)
				require.NotNil(t, c.Rules["recipients"].RejectResponse)
				require.Equal(t, smtp.StatusCode(550), c.Rules["recipients"].RejectResponse.StatusCode)
				require.Equal(t, smtp.EnhancedStatusCode{5, 7, 1}, c.Rules["recipients"].RejectResponse.EnhancedStatusCode)
				require.Equal(t, "foobar", c.Rules["recipients"].RejectResponse.Message)
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := &mail.Config{}
			err := json.Unmarshal([]byte(tc.data), &c)
			require.NoError(t, err)
			tc.test(t, c)
		})
	}
}

func TestParseConfig(t *testing.T) {
	testcases := []struct {
		name   string
		config *mail.Config
		test   func(t *testing.T, err error)
	}{
		{
			name:   "missing name",
			config: &mail.Config{},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "mail configuration missing title")
			},
		},
		{
			name: "invalid server name",
			config: &mail.Config{
				Info: mail.Info{Name: "Test Server"},
				Servers: map[string]*mail.Server{
					"invalid!": {},
				},
			},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "server name 'invalid!' does not match valid pattern")
			},
		},
		{
			name: "valid server name",
			config: &mail.Config{
				Info: mail.Info{Name: "Test Server"},
				Servers: map[string]*mail.Server{
					"valid": {Protocol: "smtps"},
				},
			},
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "missing protocol",
			config: &mail.Config{
				Info: mail.Info{Name: "Test Server"},
				Servers: map[string]*mail.Server{
					"valid": {},
				},
			},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "missing protocol for server 'valid'")
			},
		},
		{
			name: "mailbox without name",
			config: &mail.Config{
				Info:      mail.Info{Name: "Test Server"},
				Mailboxes: map[string]*mail.MailboxConfig{"": {}},
			},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "mailbox name is required")
			},
		},
		{
			name: "invalid rule name",
			config: &mail.Config{
				Info: mail.Info{Name: "Test Server"},
				Rules: map[string]*mail.Rule{
					"invalid!": {},
				},
			},
			test: func(t *testing.T, err error) {
				require.EqualError(t, err, "rule name 'invalid!' does not match valid pattern")
			},
		},
		{
			name: "valid rule name",
			config: &mail.Config{
				Info: mail.Info{Name: "Test Server"},
				Rules: map[string]*mail.Rule{
					"valid": {},
				},
			},
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "rule name from convert",
			config: &mail.Config{
				Info: mail.Info{Name: "Test Server"},
				Rules: map[string]*mail.Rule{
					"1": {},
				},
			},
			test: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Parse(&dynamic.Config{}, &dynamictest.Reader{})
			tc.test(t, err)
		})
	}
}

func newCloneFixture() *mail.Config {
	return &mail.Config{
		Version: "1.0",
		Info: mail.Info{
			Name:    "Test",
			Version: "1.2.3",
			Contact: &mail.Contact{Name: "Alice", Email: "alice@foo.bar"},
		},
		Servers:  map[string]*mail.Server{"smtp": {Host: ":25", Protocol: "smtp"}},
		Settings: &mail.Settings{MaxRecipients: 5, MaxInboxMails: 10},
		Mailboxes: map[string]*mail.MailboxConfig{
			"alice@foo.bar": {
				Username: "alice",
				Folders: map[string]*mail.FolderConfig{
					"INBOX": {
						Flags:   []string{"a"},
						Folders: map[string]*mail.FolderConfig{"Sub": {Flags: []string{"b"}}},
					},
				},
			},
		},
		Rules: map[string]*mail.Rule{
			"deny": {
				Name:           "deny",
				Sender:         mail.NewRuleExpr(".*"),
				Action:         mail.Deny,
				RejectResponse: &mail.RejectResponse{Message: "no"},
			},
		},
	}
}

func TestConfig_Clone(t *testing.T) {
	t.Run("equal to the original", func(t *testing.T) {
		c := newCloneFixture()
		require.Equal(t, c, c.Clone())
	})

	t.Run("nil", func(t *testing.T) {
		var c *mail.Config
		require.Nil(t, c.Clone())
	})

	t.Run("changing the clone does not change the original", func(t *testing.T) {
		c := newCloneFixture()
		clone := c.Clone()

		clone.Info.Contact.Name = "Bob"
		clone.Servers["smtp"].Host = ":2525"
		clone.Servers["imap"] = &mail.Server{}
		clone.Settings.MaxRecipients = 99
		mb := clone.Mailboxes["alice@foo.bar"]
		mb.Username = "changed"
		inbox := mb.Folders["INBOX"]
		inbox.Flags[0] = "changed"
		inbox.Folders["Sub"].Flags[0] = "changed"
		inbox.Folders["New"] = &mail.FolderConfig{}
		clone.Mailboxes["bob@foo.bar"] = &mail.MailboxConfig{}
		clone.Rules["deny"].Action = mail.Allow
		clone.Rules["other"] = &mail.Rule{}

		require.Equal(t, newCloneFixture(), c)
	})

	t.Run("nil values", func(t *testing.T) {
		c := &mail.Config{
			Servers:   map[string]*mail.Server{"smtp": nil},
			Mailboxes: map[string]*mail.MailboxConfig{"alice@foo.bar": nil, "bob@foo.bar": {Folders: map[string]*mail.FolderConfig{"INBOX": nil}}},
			Rules:     map[string]*mail.Rule{"r": nil},
		}
		require.Equal(t, c, c.Clone())
	})
}
