package flags_test

import (
	"mokapi/config/static"
	"mokapi/pkg/cli"
	"mokapi/pkg/cmd/mokapi"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoot_Health(t *testing.T) {
	testcases := []struct {
		name string
		cmd  func(t *testing.T) *cli.Command
		test func(t *testing.T, cfg *static.Config, flags *cli.FlagSet)
	}{
		{
			name: "port",
			cmd: func(t *testing.T) *cli.Command {
				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{"--health-port", "8081"})
				return cmd
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, 8081, cfg.Health.Port)
			},
		},
		{
			name: "path",
			cmd: func(t *testing.T) *cli.Command {
				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{"--health-path", "/health/live"})
				return cmd
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, "/health/live", cfg.Health.Path)
			},
		},
		{
			name: "disabled",
			cmd: func(t *testing.T) *cli.Command {
				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{"--health-enabled", "false"})
				return cmd
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, false, cfg.Health.Enabled)
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
