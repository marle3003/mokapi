package mail

import (
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestConfig_Patch(t *testing.T) {
	mustCompile := func(s string) *RuleExpr {
		r, _ := regexp.Compile(s)
		return NewRuleExpr(r)
	}

	testcases := []struct {
		name    string
		configs []*Config
		test    func(t *testing.T, result *Config)
	}{
		{
			name: "description and version",
			configs: []*Config{
				{},
				{Info: Info{
					Description: "foo",
					Version:     "2.0",
				}},
			},
			test: func(t *testing.T, result *Config) {
				require.Equal(t, "foo", result.Info.Description)
				require.Equal(t, "2.0", result.Info.Version)
			},
		},
		{
			name: "add rule",
			configs: []*Config{
				{},
				{Rules: []Rule{{Name: "foo", Sender: mustCompile(".*")}}},
			},
			test: func(t *testing.T, result *Config) {
				require.Len(t, result.Rules, 1)
				require.Equal(t, ".*", result.Rules[0].Sender.String())
			},
		},
		{
			name: "patch rule",
			configs: []*Config{
				{Rules: []Rule{{Name: "foo"}}},
				{Rules: []Rule{{Name: "foo", Sender: mustCompile(".*")}}},
			},
			test: func(t *testing.T, result *Config) {
				require.Len(t, result.Rules, 1)
				require.Equal(t, ".*", result.Rules[0].Sender.String())
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
