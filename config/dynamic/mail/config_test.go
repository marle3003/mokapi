package mail_test

import (
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/mail"
	"mokapi/smtp"
	"testing"
)

func TestParseConfig(t *testing.T) {
	testcases := []struct {
		name string
		data string
		test func(t *testing.T, c *mail.Config)
	}{
		{
			name: "empty",
			data: "",
			test: func(t *testing.T, c *mail.Config) {
			},
		},
		{
			name: "server with one rule",
			data: `
smtp: 1.0
server: smtps://localhost:8025
rules:
  - sender: .*@foo.bar
    action: allow
`,
			test: func(t *testing.T, c *mail.Config) {
				require.Equal(t, "smtps://localhost:8025", c.Server)
				require.Len(t, c.Rules, 1)
				require.Equal(t, ".*@foo.bar", c.Rules[0].Sender.String())
				require.Equal(t, mail.Allow, c.Rules[0].Action)
			},
		},
		{
			name: "server with custom response rule",
			data: `
smtp: 1.0
server: smtps://localhost:8025
rules:
  - recipient: .*@foo.bar
    action: deny
    rejectResponse:
      statusCode: 550
      enhancedStatusCode: 5.7.1
      text: foobar
`,
			test: func(t *testing.T, c *mail.Config) {
				require.Equal(t, "smtps://localhost:8025", c.Server)
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
