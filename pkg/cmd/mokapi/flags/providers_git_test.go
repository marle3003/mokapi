package flags_test

import (
	"mokapi/config/static"
	"mokapi/pkg/cli"
	"mokapi/pkg/cmd/mokapi"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoot_Providers_Git(t *testing.T) {
	testcases := []struct {
		name string
		cmd  *cli.Command
		test func(t *testing.T, cfg *static.Config, flags *cli.FlagSet)
	}{
		{
			name: "--providers-git-repositories",
			cmd: func() *cli.Command {
				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{"--providers-git-repositories url=https://github.com/foo/foo.git,include=*.json url=https://github.com/bar/bar.git,include=*.yaml"})
				return cmd
			}(),
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []static.GitRepo{
					{Url: "https://github.com/foo/foo.git", Include: []string{"*.json"}},
					{Url: "https://github.com/bar/bar.git", Include: []string{"*.yaml"}},
				}, cfg.Providers.Git.Repositories)
			},
		},
		{
			name: "env variable using shorthand syntax",
			cmd: func() *cli.Command {
				key := "MOKAPI_PROVIDERS_GIT"
				err := os.Setenv(key, "pullInterval=10s,tempDir=/tempdir")
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key)
				})

				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{})
				return cmd
			}(),
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, "10s", cfg.Providers.Git.PullInterval)
				require.Equal(t, "/tempdir", cfg.Providers.Git.TempDir)
			},
		},
		{
			name: "index url",
			cmd: func() *cli.Command {
				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{"--providers-git-repositories[0]-url https://github.com/foo/foo.git"})
				return cmd
			}(),
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []static.GitRepo{
					{Url: "https://github.com/foo/foo.git"},
				}, cfg.Providers.Git.Repositories)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				cli.SetFileReader(&cli.FileReader{})
			}()

			cmd := tc.cmd
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
