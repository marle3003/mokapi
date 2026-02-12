package flags_test

import (
	"mokapi/config/static"
	"mokapi/pkg/cli"
	"mokapi/pkg/cmd/mokapi"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoot_Providers_Npm(t *testing.T) {
	testcases := []struct {
		name string
		args func(t *testing.T) []string
		test func(t *testing.T, cfg *static.Config, flags *cli.FlagSet)
	}{
		{
			name: "provider shorthand",
			args: func(t *testing.T) []string {
				return []string{"--providers-npm", `package={"name": "foo-api"},globalFolders=/npm`}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, "foo-api", cfg.Providers.Npm.Packages[0].Name)
				require.Equal(t, "/npm", cfg.Providers.Npm.GlobalFolders[0])
			},
		},
		{
			name: "package shorthand",
			args: func(t *testing.T) []string {
				return []string{"--providers-npm-package", "name=foo-api,file=api.yaml"}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, "foo-api", cfg.Providers.Npm.Packages[0].Name)
				require.Equal(t, []string{"api.yaml"}, cfg.Providers.Npm.Packages[0].Files)
			},
		},
		{
			name: "packages shorthand",
			args: func(t *testing.T) []string {
				return []string{"--providers-npm-packages", "name=foo-api name=bar-api"}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, "foo-api", cfg.Providers.Npm.Packages[0].Name)
				require.Equal(t, "bar-api", cfg.Providers.Npm.Packages[1].Name)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := mokapi.NewCmdMokapi()
			cmd.SetArgs(tc.args(t))

			cfg := static.NewConfig()
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
