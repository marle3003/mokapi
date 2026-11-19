package mokapi_test

import (
	"mokapi/config/static"
	"mokapi/pkg/cli"
	"mokapi/pkg/cmd/mokapi"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Precedence of configuration sources is CLI > environment variable > config file.
// A list from a higher source replaces the list of a lower source, whereas
// repeating a list flag within the same source accumulates its values.
func TestConfigSourcePrecedence(t *testing.T) {
	type source struct {
		cli  []string
		env  map[string]string
		file string
	}

	testcases := []struct {
		name   string
		source source
		test   func(t *testing.T, cfg *static.Config)
	}{
		// scalar: log level
		{
			name:   "log level: env only",
			source: source{env: map[string]string{"MOKAPI_LOG_LEVEL": "info"}},
			test:   func(t *testing.T, cfg *static.Config) { require.Equal(t, "info", cfg.Log.Level) },
		},
		{
			name:   "log level: cli over env",
			source: source{cli: []string{"--log-level", "debug"}, env: map[string]string{"MOKAPI_LOG_LEVEL": "info"}},
			test:   func(t *testing.T, cfg *static.Config) { require.Equal(t, "debug", cfg.Log.Level) },
		},
		{
			name:   "log level: cli over file",
			source: source{cli: []string{"--log-level", "debug"}, file: "log:\n  level: warn\n"},
			test:   func(t *testing.T, cfg *static.Config) { require.Equal(t, "debug", cfg.Log.Level) },
		},
		{
			name:   "log level: env over file",
			source: source{env: map[string]string{"MOKAPI_LOG_LEVEL": "info"}, file: "log:\n  level: warn\n"},
			test:   func(t *testing.T, cfg *static.Config) { require.Equal(t, "info", cfg.Log.Level) },
		},
		{
			name: "log level: cli over env over file",
			source: source{cli: []string{"--log-level", "debug"}, env: map[string]string{"MOKAPI_LOG_LEVEL": "info"},
				file: "log:\n  level: warn\n"},
			test: func(t *testing.T, cfg *static.Config) { require.Equal(t, "debug", cfg.Log.Level) },
		},

		// list with explode flag: providers-file-filename
		{
			name: "filename: cli over env",
			source: source{cli: []string{"--providers-file-filename", "b.yaml"},
				env: map[string]string{"MOKAPI_PROVIDERS_FILE_FILENAME": "a.yaml"}},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"b.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name:   "filename: cli over file",
			source: source{cli: []string{"--providers-file-filename", "b.yaml"}, file: "providers:\n  file:\n    filenames: [a.yaml]\n"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"b.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "filename: env over file",
			source: source{env: map[string]string{"MOKAPI_PROVIDERS_FILE_FILENAME": "b.yaml"},
				file: "providers:\n  file:\n    filenames: [a.yaml]\n"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"b.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "filename: cli over env over file",
			source: source{cli: []string{"--providers-file-filename", "c.yaml"},
				env:  map[string]string{"MOKAPI_PROVIDERS_FILE_FILENAME": "b.yaml"},
				file: "providers:\n  file:\n    filenames: [a.yaml]\n"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"c.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name:   "filename: cli over longer file list",
			source: source{cli: []string{"--providers-file-filename", "c.yaml"}, file: "providers:\n  file:\n    filenames: [a.yaml, b.yaml]\n"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"c.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "filename: env over longer file list",
			source: source{env: map[string]string{"MOKAPI_PROVIDERS_FILE_FILENAME": "c.yaml"},
				file: "providers:\n  file:\n    filenames: [a.yaml, b.yaml]\n"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"c.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name:   "filename: repeated cli flag yields both",
			source: source{cli: []string{"--providers-file-filename", "a.yaml", "--providers-file-filename", "b.yaml"}},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"a.yaml", "b.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "filename: repeated cli flag over env and file",
			source: source{cli: []string{"--providers-file-filename", "c.yaml", "--providers-file-filename", "d.yaml"},
				env:  map[string]string{"MOKAPI_PROVIDERS_FILE_FILENAME": "b.yaml"},
				file: "providers:\n  file:\n    filenames: [a.yaml]\n"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"c.yaml", "d.yaml"}, cfg.Providers.File.Filenames)
			},
		},

		// list without explode: providers-file-filenames (file has more items than the higher source)
		{
			name: "filenames: cli over env",
			source: source{cli: []string{"--providers-file-filenames", "c.yaml"},
				env: map[string]string{"MOKAPI_PROVIDERS_FILE_FILENAMES": "a.yaml,b.yaml"}},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"c.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name:   "filenames: cli over file",
			source: source{cli: []string{"--providers-file-filenames", "c.yaml"}, file: "providers:\n  file:\n    filenames: [a.yaml, b.yaml]\n"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"c.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "filenames: env over file",
			source: source{env: map[string]string{"MOKAPI_PROVIDERS_FILE_FILENAMES": "c.yaml"},
				file: "providers:\n  file:\n    filenames: [a.yaml, b.yaml]\n"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"c.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "filenames: cli over env over file",
			source: source{cli: []string{"--providers-file-filenames", "d.yaml"},
				env:  map[string]string{"MOKAPI_PROVIDERS_FILE_FILENAMES": "c.yaml"},
				file: "providers:\n  file:\n    filenames: [a.yaml, b.yaml]\n"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"d.yaml"}, cfg.Providers.File.Filenames)
			},
		},

		// list of structs: a higher source replaces the items, fields of the old items are not inherited
		{
			name: "repositories: cli over file",
			source: source{cli: []string{"--providers-git-repositories", "url=b"},
				file: "providers:\n  git:\n    repositories:\n      - url: a\n        pullInterval: 5m\n"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Len(t, cfg.Providers.Git.Repositories, 1)
				require.Equal(t, "b", cfg.Providers.Git.Repositories[0].Url)
				require.Empty(t, cfg.Providers.Git.Repositories[0].PullInterval)
			},
		},

		// an empty value does not clear the list of a lower source
		{
			name: "filenames: empty env keeps file",
			source: source{env: map[string]string{"MOKAPI_PROVIDERS_FILE_FILENAMES": ""},
				file: "providers:\n  file:\n    filenames: [a.yaml, b.yaml]\n"},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"a.yaml", "b.yaml"}, cfg.Providers.File.Filenames)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.source.env {
				t.Setenv(k, v)
			}
			// non-nil: a nil slice would make Execute read os.Args
			args := append([]string{}, tc.source.cli...)
			if tc.source.file != "" {
				f := filepath.Join(t.TempDir(), "mokapi.yaml")
				require.NoError(t, os.WriteFile(f, []byte(tc.source.file), 0644))
				args = append([]string{"--config-file", f}, args...)
			}

			cmd := mokapi.NewCmdMokapi()
			cmd.SetArgs(args)
			var cfg *static.Config
			cmd.Run = func(cmd *cli.Command, args []string) error {
				cfg = cmd.Config.(*static.Config)
				return nil
			}
			require.NoError(t, cmd.Execute())

			tc.test(t, cfg)
		})
	}
}
