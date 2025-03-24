package directory

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"testing"
)

func TestConfig(t *testing.T) {
	testcases := []struct {
		name  string
		input string
		test  func(t *testing.T, cfg *Config, err error)
	}{
		{
			name:  "host overwrites server address",
			input: `{"ldap": "1.0", "server": { "address": "192.168.0.1:389" }, "host": ":389" }`,
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
				require.Equal(t, ":389", cfg.Address)
			},
		},
		{
			name:  "set Root DSE by config",
			input: `{"ldap": "1.0", "info": { "name": "foo", "description": "bar" } }`,
			test: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
				root := cfg.Entries.Lookup("")
				require.Equal(t, []string{"foo"}, root.Attributes["dsServiceName"])
				require.Equal(t, []string{"bar"}, root.Attributes["description"])
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var c *Config
			err := json.Unmarshal([]byte(tc.input), &c)
			if err == nil {
				err = c.Parse(&dynamic.Config{Info: dynamictest.NewConfigInfo()}, &dynamictest.Reader{})
			}
			tc.test(t, c, err)
		})
	}
}
