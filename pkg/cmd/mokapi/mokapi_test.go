package mokapi_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/config/static"
	"mokapi/pkg/cli"
	"mokapi/pkg/cmd/mokapi"
	"testing"
)

func TestMain_Flags(t *testing.T) {
	testcases := []struct {
		name string
		args []string
		test func(t *testing.T, cfg *static.Config)
	}{
		{
			name: "default",
			args: []string{},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "info", cfg.Log.Level)
				require.Equal(t, "text", cfg.Log.Format)
				require.Equal(t, "8080", cfg.Api.Port)
				require.Equal(t, true, cfg.Api.Dashboard)
				require.Equal(t, []string{"_"}, cfg.Providers.File.SkipPrefix)
				require.Equal(t, int64(100), cfg.Event.Store["default"].Size)
				require.Equal(t, "0.85", cfg.DataGen.OptionalProperties)
			},
		},
		{
			name: "--providers-file-filenames",
			args: []string{"--providers-file-filenames", "foo.yaml"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "--providers-file-filenames comma separated",
			args: []string{"--providers-file-filenames", "foo.yaml,bar.yaml"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo.yaml", "bar.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "--providers-file-filenames twice",
			args: []string{"--providers-file-filenames", "foo.yaml", "--providers-file-filenames", "bar.yaml"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo.yaml", "bar.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "--providers-file-filename",
			args: []string{"--providers-file-filename", "foo.yaml"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"foo.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "--event-store",
			args: []string{"--event-store", "foo={\"size\":250}"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, static.Store{Size: 250}, cfg.Event.Store["foo"])
			},
		},
		{
			name: "--event-store-foo-size",
			args: []string{"--event-store-foo-size", "250"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, int64(250), cfg.Event.Store["foo"].Size)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := mokapi.NewCmdMokapi(context.Background())
			cmd.SetArgs(tc.args)

			cfg := static.NewConfig()
			cmd.Run = func(cmd *cli.Command, args []string) error {
				cfg = cmd.Config.(*static.Config)
				return nil
			}
			err := cmd.Execute()
			require.NoError(t, err)

			tc.test(t, cfg)
		})
	}
}
