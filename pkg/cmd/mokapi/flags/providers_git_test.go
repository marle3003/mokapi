package flags_test

import (
	"mokapi/config/static"
	"mokapi/pkg/cli"
	"mokapi/pkg/cli/clitest"
	"mokapi/pkg/cmd/mokapi"
	"os"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

func TestRoot_Providers_Git(t *testing.T) {
	testcases := []struct {
		name string
		cmd  func(t *testing.T) *cli.Command
		test func(t *testing.T, cfg *static.Config, flags *cli.FlagSet)
	}{
		{
			name: "--providers-git-repositories",
			cmd: func(t *testing.T) *cli.Command {
				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{"--providers-git-repositories url=https://github.com/foo/foo.git,include=*.json url=https://github.com/bar/bar.git,include=*.yaml"})
				return cmd
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []static.GitRepo{
					{Url: "https://github.com/foo/foo.git", Include: []string{"*.json"}},
					{Url: "https://github.com/bar/bar.git", Include: []string{"*.yaml"}},
				}, cfg.Providers.Git.Repositories)
			},
		},
		{
			name: "env variable using shorthand syntax",
			cmd: func(t *testing.T) *cli.Command {
				key := "MOKAPI_PROVIDERS_GIT"
				err := os.Setenv(key, "pullInterval=10s,tempDir=/tempdir")
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key)
				})

				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{})
				return cmd
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, "10s", cfg.Providers.Git.PullInterval)
				require.Equal(t, "/tempdir", cfg.Providers.Git.TempDir)
			},
		},
		{
			name: "index url",
			cmd: func(t *testing.T) *cli.Command {
				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{"--providers-git-repositories[0]-url https://github.com/foo/foo.git"})
				return cmd
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []static.GitRepo{
					{Url: "https://github.com/foo/foo.git"},
				}, cfg.Providers.Git.Repositories)
			},
		},
		{
			name: "github private key from env variable and cli flags",
			cmd: func(t *testing.T) *cli.Command {
				key1 := "MOKAPI_Providers_Git_Repositories_0_Auth_GitHub_PrivateKey"
				err := os.Setenv(key1, `-----BEGIN RSA PRIVATE KEY-----
MIIBOQIBAAJAXWRPQyGlEY+SXz8Uslhe+MLjTgWd8lf/nA0hgCm9JFKC1tq1S73c
Q9naClNXsMqY7pwPt1bSY8jYRqHHbdoUvwIDAQABAkAfJkz1pCwtfkig8iZSEf2j
VUWBiYgUA9vizdJlsAZBLceLrdk8RZF2YOYCWHrpUtZVea37dzZJe99Dr53K0UZx
AiEAtyHQBGoCVHfzPM//a+4tv2ba3tx9at+3uzGR86YNMzcCIQCCjWHcLW/+sQTW
OXeXRrtxqHPp28ir8AVYuNX0nT1+uQIgJm158PMtufvRlpkux78a6mby1oD98Ecx
jp5AOhhF/NECICyHsQN69CJ5mt6/R01wMOt5u9/eubn76rbyhPgk0h7xAiEAjn6m
EmLwkIYD9VnZfp9+2UoWSh0qZiTIHyNwFpJH78o=
-----END RSA PRIVATE KEY-----

`)
				require.NoError(t, err)
				key2 := "MOKAPI_Providers_Git_Repositories_1_Auth_GitHub_PrivateKey"
				err = os.Setenv(key2, `-----BEGIN RSA PRIVATE KEY-----
MIIBOAIBAAJARsF2wfXtjllRR8nnz8+CLULn0bqgZtYktJB2BdcB5bw6OYmmDVCc
TeTC3VXZATdSqNA6WDWCkSVinC05uYEOEwIDAQABAkArUAaYmSkAeKCO54Pl7Ert
1gT+l9XU3cW+WqhEzuc0cC4Eiqe9phpdiQXNosI60a8YyeyBUjCtQGFwbJ1Kl8Hh
AiEAioOWu1s5nbB6ioOXdhbW4Ov5xfI62TYJNxdz656/njsCIQCCxRfwRVfDcC0h
hvuOpFzvZ870deo1/OD8j4U8jG+aCQIgXeU55qO+eODLEN6Ha+urmikc1kyQC/KP
aKMjV5PzfUUCIHX2s4yEERJ1K9EVwfE/5bH1E+TERb3j21UZZphjGv15AiBBs0w5
WRuPspPXIAHPKrjEHkUsgDZHW/V0fJWbIjJarw==
-----END RSA PRIVATE KEY-----
`)
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key1)
					_ = os.Unsetenv(key2)
				})

				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{"--providers-git-repositories url=https://github.com/foo/foo.git url=https://github.com/bar/bar.git"})
				return cmd
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Len(t, cfg.Providers.Git.Repositories, 2)

				require.Equal(t, "https://github.com/foo/foo.git", cfg.Providers.Git.Repositories[0].Url)
				require.NotNil(t, cfg.Providers.Git.Repositories[0].Auth)
				require.True(t, strings.HasPrefix(cfg.Providers.Git.Repositories[0].Auth.GitHub.PrivateKey.String(), "-----BEGIN RSA PRIVATE KEY-----"))
				_, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.Providers.Git.Repositories[0].Auth.GitHub.PrivateKey))
				require.NoError(t, err)

				require.Equal(t, "https://github.com/foo/foo.git", cfg.Providers.Git.Repositories[0].Url)
				require.True(t, strings.HasPrefix(cfg.Providers.Git.Repositories[1].Auth.GitHub.PrivateKey.String(), "-----BEGIN RSA PRIVATE KEY-----"))
				_, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.Providers.Git.Repositories[1].Auth.GitHub.PrivateKey))
				require.NoError(t, err)
			},
		},
		{
			name: "github private key from env variable and config file",
			cmd: func(t *testing.T) *cli.Command {
				key1 := "MOKAPI_Providers_Git_Repositories_0_Auth_GitHub_PrivateKey"
				err := os.Setenv(key1, `-----BEGIN RSA PRIVATE KEY-----
MIIBOQIBAAJAXWRPQyGlEY+SXz8Uslhe+MLjTgWd8lf/nA0hgCm9JFKC1tq1S73c
Q9naClNXsMqY7pwPt1bSY8jYRqHHbdoUvwIDAQABAkAfJkz1pCwtfkig8iZSEf2j
VUWBiYgUA9vizdJlsAZBLceLrdk8RZF2YOYCWHrpUtZVea37dzZJe99Dr53K0UZx
AiEAtyHQBGoCVHfzPM//a+4tv2ba3tx9at+3uzGR86YNMzcCIQCCjWHcLW/+sQTW
OXeXRrtxqHPp28ir8AVYuNX0nT1+uQIgJm158PMtufvRlpkux78a6mby1oD98Ecx
jp5AOhhF/NECICyHsQN69CJ5mt6/R01wMOt5u9/eubn76rbyhPgk0h7xAiEAjn6m
EmLwkIYD9VnZfp9+2UoWSh0qZiTIHyNwFpJH78o=
-----END RSA PRIVATE KEY-----

`)
				require.NoError(t, err)
				key2 := "MOKAPI_Providers_Git_Repositories_1_Auth_GitHub_PrivateKey"
				err = os.Setenv(key2, `-----BEGIN RSA PRIVATE KEY-----
MIIBOAIBAAJARsF2wfXtjllRR8nnz8+CLULn0bqgZtYktJB2BdcB5bw6OYmmDVCc
TeTC3VXZATdSqNA6WDWCkSVinC05uYEOEwIDAQABAkArUAaYmSkAeKCO54Pl7Ert
1gT+l9XU3cW+WqhEzuc0cC4Eiqe9phpdiQXNosI60a8YyeyBUjCtQGFwbJ1Kl8Hh
AiEAioOWu1s5nbB6ioOXdhbW4Ov5xfI62TYJNxdz656/njsCIQCCxRfwRVfDcC0h
hvuOpFzvZ870deo1/OD8j4U8jG+aCQIgXeU55qO+eODLEN6Ha+urmikc1kyQC/KP
aKMjV5PzfUUCIHX2s4yEERJ1K9EVwfE/5bH1E+TERb3j21UZZphjGv15AiBBs0w5
WRuPspPXIAHPKrjEHkUsgDZHW/V0fJWbIjJarw==
-----END RSA PRIVATE KEY-----
`)
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key1)
					_ = os.Unsetenv(key2)
				})

				cmd := mokapi.NewCmdMokapi()

				cli.SetFileReader(&clitest.TestFileReader{Files: map[string][]byte{
					"/etc/foo/mokapi.yaml": []byte(`
providers:
  git:
    repositories:
      - url: https://github.com/foo/foo.git
        auth:
          github:
            appId: 1
            installationId: 123456
      - url: https://github.com/bar/bar.git
        auth:
          github:
            appId: 2
            installationId: 823242
`),
				}})
				t.Cleanup(func() {
					cli.SetFileReader(&cli.FileReader{})
				})
				cmd.SetConfigFile("/etc/foo/mokapi.yaml")
				cmd.SetArgs([]string{})

				return cmd
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Len(t, cfg.Providers.Git.Repositories, 2)

				require.Equal(t, "https://github.com/foo/foo.git", cfg.Providers.Git.Repositories[0].Url)
				require.NotNil(t, cfg.Providers.Git.Repositories[0].Auth)
				require.True(t, strings.HasPrefix(cfg.Providers.Git.Repositories[0].Auth.GitHub.PrivateKey.String(), "-----BEGIN RSA PRIVATE KEY-----"))
				_, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.Providers.Git.Repositories[0].Auth.GitHub.PrivateKey))
				require.NoError(t, err)

				require.Equal(t, "https://github.com/foo/foo.git", cfg.Providers.Git.Repositories[0].Url)
				require.True(t, strings.HasPrefix(cfg.Providers.Git.Repositories[1].Auth.GitHub.PrivateKey.String(), "-----BEGIN RSA PRIVATE KEY-----"))
				_, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.Providers.Git.Repositories[1].Auth.GitHub.PrivateKey))
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				cli.SetFileReader(&cli.FileReader{})
			}()

			cmd := tc.cmd(t)
			var cfg *static.Config
			cmd.Run = func(cmd *cli.Command, args []string) error {
				cfg = cmd.Config.(*static.Config)
				return nil
			}
			err := cmd.Execute()
			require.NoError(t, err)

			tc.test(t, cfg, cmd.Flags())
		})
	}
}
