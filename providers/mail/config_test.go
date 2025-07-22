package mail_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
  - name: alice@mokapi.io
    username: alice
    password: mokapi
`,
			test: func(t *testing.T, c *mail.Config) {
				require.Equal(t, "alice@mokapi.io", c.Mailboxes[0].Name)
				require.Equal(t, "alice", c.Mailboxes[0].Username)
				require.Equal(t, "mokapi", c.Mailboxes[0].Password)
				require.Len(t, c.Mailboxes[0].Folders, 0)
			},
		},
		{
			name: "mailbox with folder",
			data: `
mail: 1.0
mailboxes:
  - name: alice@mokapi.io
    username: alice
    password: mokapi
    folders: 
      - name: inbox
        flags: [\HasNoChildren]
`,
			test: func(t *testing.T, c *mail.Config) {
				require.Equal(t, "alice@mokapi.io", c.Mailboxes[0].Name)
				require.Equal(t, "alice", c.Mailboxes[0].Username)
				require.Equal(t, "mokapi", c.Mailboxes[0].Password)
				require.Equal(t, "inbox", c.Mailboxes[0].Folders[0].Name)
				require.Equal(t, []string{"\\HasNoChildren"}, c.Mailboxes[0].Folders[0].Flags)
				require.Len(t, c.Mailboxes[0].Folders[0].Folders, 0)
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
  - sender: .*@foo.bar
    action: allow
`,
			test: func(t *testing.T, c *mail.Config) {
				require.Len(t, c.Servers, 1)
				require.Equal(t, "localhost:8025", c.Servers["smtps"].Host)
				require.Equal(t, "smtps", c.Servers["smtps"].Protocol)
				require.Len(t, c.Rules, 1)
				require.Equal(t, ".*@foo.bar", c.Rules[0].Sender.String())
				require.Equal(t, mail.Allow, c.Rules[0].Action)
			},
		},
		{
			name: "custom response rule",
			data: `
smtp: 1.0
rules:
  - recipient: .*@foo.bar
    action: deny
    rejectResponse:
      statusCode: 550
      enhancedStatusCode: 5.7.1
      text: foobar
`,
			test: func(t *testing.T, c *mail.Config) {
				require.Len(t, c.Rules, 1)
				require.Equal(t, ".*@foo.bar", c.Rules[0].Recipient.String())
				require.Equal(t, mail.Deny, c.Rules[0].Action)
				require.NotNil(t, c.Rules[0].RejectResponse)
				require.Equal(t, smtp.StatusCode(550), c.Rules[0].RejectResponse.StatusCode)
				require.Equal(t, smtp.EnhancedStatusCode{5, 7, 1}, c.Rules[0].RejectResponse.EnhancedStatusCode)
				require.Equal(t, "foobar", c.Rules[0].RejectResponse.Text)
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
"mailboxes": [
  { "name": "alice@mokapi.io",
    "username": "alice",
    "password": "mokapi"
  }
]}`,
			test: func(t *testing.T, c *mail.Config) {
				require.Equal(t, "alice@mokapi.io", c.Mailboxes[0].Name)
				require.Equal(t, "alice", c.Mailboxes[0].Username)
				require.Equal(t, "mokapi", c.Mailboxes[0].Password)
				require.Len(t, c.Mailboxes[0].Folders, 0)
			},
		},
		{
			name: "mailbox with folder",
			data: `{
"mail": "1.0",
"mailboxes": [
  { "name": "alice@mokapi.io",
    "username": "alice",
    "password": "mokapi",
    "folders": [
      { "name": "inbox",
        "flags": ["\\HasNoChildren"]
      }
    ]
  }
]}`,
			test: func(t *testing.T, c *mail.Config) {
				require.Equal(t, "alice@mokapi.io", c.Mailboxes[0].Name)
				require.Equal(t, "alice", c.Mailboxes[0].Username)
				require.Equal(t, "mokapi", c.Mailboxes[0].Password)
				require.Equal(t, "inbox", c.Mailboxes[0].Folders[0].Name)
				require.Equal(t, []string{"\\HasNoChildren"}, c.Mailboxes[0].Folders[0].Flags)
				require.Len(t, c.Mailboxes[0].Folders[0].Folders, 0)
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
"rules": [
  { "sender": ".*@foo.bar",
    "action": "allow"
  }]
}`,
			test: func(t *testing.T, c *mail.Config) {
				require.Len(t, c.Servers, 1)
				require.Equal(t, "localhost:8025", c.Servers["smtps"].Host)
				require.Equal(t, "smtps", c.Servers["smtps"].Protocol)
				require.Len(t, c.Rules, 1)
				require.Equal(t, ".*@foo.bar", c.Rules[0].Sender.String())
				require.Equal(t, mail.Allow, c.Rules[0].Action)
			},
		},
		{
			name: "custom response rule",
			data: `{
"smtp": "1.0",
"rules": [
  { "recipient": ".*@foo.bar",
    "action": "deny",
    "rejectResponse": {
      "statusCode": 550,
      "enhancedStatusCode": "5.7.1",
      "text": "foobar"
    }
  }]
}`,
			test: func(t *testing.T, c *mail.Config) {
				require.Len(t, c.Rules, 1)
				require.Equal(t, ".*@foo.bar", c.Rules[0].Recipient.String())
				require.Equal(t, mail.Deny, c.Rules[0].Action)
				require.NotNil(t, c.Rules[0].RejectResponse)
				require.Equal(t, smtp.StatusCode(550), c.Rules[0].RejectResponse.StatusCode)
				require.Equal(t, smtp.EnhancedStatusCode{5, 7, 1}, c.Rules[0].RejectResponse.EnhancedStatusCode)
				require.Equal(t, "foobar", c.Rules[0].RejectResponse.Text)
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
