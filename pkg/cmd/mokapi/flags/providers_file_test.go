package flags_test

import (
	"mokapi/config/static"
	"mokapi/pkg/cli"
	"mokapi/pkg/cmd/mokapi"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoot_Providers_File(t *testing.T) {
	testcases := []struct {
		name string
		cmd  func(t *testing.T) *cli.Command
		test func(t *testing.T, cfg *static.Config, flags *cli.FlagSet)
	}{
		{
			name: "directories index path",
			cmd: func(t *testing.T) *cli.Command {
				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{"--providers-file-directories[0]-path ./foo"})
				return cmd
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []static.FileConfig{
					{Path: "./foo"},
				}, cfg.Providers.File.Directories)
			},
		},
		{
			name: "directories index path",
			cmd: func(t *testing.T) *cli.Command {
				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{"--providers-file-directories[0]-path /foo"})
				return cmd
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []static.FileConfig{
					{Path: "/foo"},
				}, cfg.Providers.File.Directories)
			},
		},
		{
			name: "directories index include",
			cmd: func(t *testing.T) *cli.Command {
				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{"--providers-file-directories[0]-include *index.js"})
				return cmd
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []static.FileConfig{
					{Include: []string{"*index.js"}},
				}, cfg.Providers.File.Directories)
			},
		},
		{
			name: "directories index include two values",
			cmd: func(t *testing.T) *cli.Command {
				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{"--providers-file-directories[0]-include *index.js *foo.ts"})
				return cmd
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []static.FileConfig{
					{Include: []string{"*index.js", "*foo.ts"}},
				}, cfg.Providers.File.Directories)
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
