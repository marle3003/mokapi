package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigKey(t *testing.T) {
	type store struct{ Size int }
	type auth struct {
		GitHub *struct {
			AppId int `yaml:"appId"`
		} `yaml:"gitHub,github"`
	}
	type repo struct {
		Url   string
		Files []string `explode:"file"`
		Auth  *auth
	}
	type config struct {
		Log struct {
			Level string
		}
		Repos      []repo `explode:"repo"`
		Store      map[string]store
		RootCaCert string `yaml:"rootCaCert" name:"root-ca-cert"`
		Names      []string
		Hidden     string `yaml:"-" flag:"hidden-flag"`
		Help       bool   `yaml:"-" aliases:"h"`
	}

	testcases := []struct {
		name   string
		config any
		flag   string
		key    string
	}{
		{name: "nested field name is lower first", config: &config{}, flag: "log-level", key: "log.level"},
		{name: "name tag with yaml tag", config: &config{}, flag: "root-ca-cert", key: "rootCaCert"},
		{name: "list item", config: &config{}, flag: "repos[<index>]-url", key: "repos[].url"},
		{name: "list item of list item written with a dash index", config: &config{}, flag: "repos-0-url", key: "repos[].url"},
		{name: "explode field", config: &config{}, flag: "repos[<index>]-file", key: "repos[].files"},
		{name: "list itself", config: &config{}, flag: "names", key: "names"},
		{name: "list replaced by explode flag", config: &config{}, flag: "repo", key: "repos"},
		{name: "yaml name is first part of tag", config: &config{}, flag: "repos[<index>]-auth-github-appid", key: "repos[].auth.gitHub.appId"},
		{name: "map key placeholder", config: &config{}, flag: "store-<name>-size", key: "store.<name>.size"},
		{name: "map key literal", config: &config{}, flag: "store-default-size", key: "store.default.size"},
		{name: "yaml dash", config: &config{}, flag: "hidden-flag", key: "-"},
		{name: "alias with yaml dash", config: &config{}, flag: "h", key: "-"},
		{name: "unknown flag", config: &config{}, flag: "unknown", key: "-"},
		{name: "no config", config: nil, flag: "log-level", key: "-"},
		{name: "config that is not a struct", config: "foo", flag: "log-level", key: "-"},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.key, configKey(tc.config, tc.flag))
		})
	}
}
