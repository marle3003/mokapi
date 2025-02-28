package directory

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/try"
	"testing"
)

func TestConfig_Schema(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		reader dynamic.Reader
		test   func(t *testing.T, cfg *Config, err error)
	}{
		{
			name:  "attributeTypes",
			input: `{ "files": [ "./config.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/config.ldif": {Raw: []byte("dn: \nsubschemaSubentry: cn=schema\n\ndn: cn=schema\nattributeTypes: ( 2.5.4.3 NAME 'cn' DESC 'Common Name' EQUALITY caseIgnoreMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.15 SINGLE-VALUE )")},
			}},
			test: func(t *testing.T, cfg *Config, err error) {
				require.NotNil(t, cfg.Schema)
				require.Contains(t, cfg.Schema.AttributeTypes, "cn")
				require.Equal(t, "caseIgnoreMatch", cfg.Schema.AttributeTypes["cn"].Equality)
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
			if err != nil {
				tc.test(t, c, err)
			} else {
				err = c.Parse(&dynamic.Config{Data: c, Info: dynamic.ConfigInfo{Url: try.MustUrl("file:/foo.yml")}}, tc.reader)
				tc.test(t, c, err)
			}

			tc.test(t, c, err)
		})
	}
}
