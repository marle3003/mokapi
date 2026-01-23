package flags_test

import (
	"mokapi/config/static"
	"mokapi/pkg/cli"
	"mokapi/pkg/cli/clitest"
	"mokapi/pkg/cmd/mokapi"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoot_Providers_Http(t *testing.T) {
	testcases := []struct {
		name string
		cmd  *cli.Command
		test func(t *testing.T, cfg *static.Config, flags *cli.FlagSet)
	}{
		{
			name: "--providers-http",
			cmd: func() *cli.Command {
				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{"--providers-http", "url=https://foo.bar/file.yaml,proxy=https://proxy.example.com"})
				return cmd
			}(),
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"https://foo.bar/file.yaml"}, cfg.Providers.Http.Urls)
				require.Equal(t, "https://proxy.example.com", cfg.Providers.Http.Proxy)
			},
		},
		{
			name: "set urls with single url using YAML file",
			cmd: func() *cli.Command {
				cmd := mokapi.NewCmdMokapi()
				cmd.SetArgs([]string{})
				cli.SetFileReader(&clitest.TestFileReader{Files: map[string][]byte{
					"/etc/foo/mokapi.yaml": []byte("providers:\n  http:\n    urls: https://foo.bar/file.yaml"),
				}})
				t.Cleanup(func() {
					cli.SetFileReader(&cli.FileReader{})
				})
				cmd.SetConfigFile("/etc/foo/mokapi.yaml")

				return cmd
			}(),
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"https://foo.bar/file.yaml"}, cfg.Providers.Http.Urls)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
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
