package mail

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// The limit must be the same for a config file as documented in docs/mail/overview.md:
// 100 mails when maxInboxMails is not set, unlimited when it is explicitly 0.
func TestStore_InboxLimitFromConfigFile(t *testing.T) {
	parsers := map[string]func(data string, c *Config) error{
		"yaml": func(data string, c *Config) error { return yaml.Unmarshal([]byte(data), c) },
		"json": func(data string, c *Config) error { return json.Unmarshal([]byte(data), c) },
	}
	testdata := map[string]map[string]struct {
		data     string
		expected int
	}{
		"yaml": {
			"not set":        {data: "settings:\n  autoCreateMailbox: true\n", expected: DefaultMaxInboxMails},
			"no settings":    {data: "info:\n  title: foo\n", expected: DefaultMaxInboxMails},
			"explicit 0":     {data: "settings:\n  maxInboxMails: 0\n", expected: 150},
			"explicit limit": {data: "settings:\n  maxInboxMails: 7\n", expected: 7},
		},
		"json": {
			"not set":        {data: `{"settings":{"autoCreateMailbox":true}}`, expected: DefaultMaxInboxMails},
			"no settings":    {data: `{"info":{"title":"foo"}}`, expected: DefaultMaxInboxMails},
			"explicit 0":     {data: `{"settings":{"maxInboxMails":0}}`, expected: 150},
			"explicit limit": {data: `{"settings":{"maxInboxMails":7}}`, expected: 7},
		},
	}

	for format, parse := range parsers {
		for name, tc := range testdata[format] {
			t.Run(format+" "+name, func(t *testing.T) {
				c := &Config{}
				require.NoError(t, parse(tc.data, c))
				s := NewStore(c)
				require.NoError(t, s.EnsureMailbox("foo@bar", c.Settings))
				mb := mustMailbox(t, s, "foo@bar")

				for i := 0; i < 150; i++ {
					mb.Append(newMessage("m"))
				}

				require.Equal(t, tc.expected, len(mb.Folders["INBOX"].Messages))
			})
		}
	}
}
