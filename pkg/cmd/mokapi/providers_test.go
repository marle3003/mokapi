package mokapi_test

import (
	"mokapi/config/static"
	"mokapi/pkg/cli"
	"mokapi/pkg/cmd/mokapi"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoot_Providers_File(t *testing.T) {
	testcases := []struct {
		name string
		args func(t *testing.T) []string
		test func(t *testing.T, cfg *static.Config, flags *cli.FlagSet)
	}{
		{
			name: "filename from cli",
			args: func(t *testing.T) []string {
				return []string{"--providers-file-filename", "foo.yaml", "--providers-file-filename", "bar.yaml"}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"foo.yaml", "bar.yaml"}, cfg.Providers.File.Filenames)
				require.Equal(t, []string{"foo.yaml", "bar.yaml"}, flags.GetStringSlice("providers-file-filename"))
			},
		},
		{
			name: "filename from cli one parameter",
			args: func(t *testing.T) []string {
				return []string{"--providers-file-filename", "foo.yaml,bar.yaml"}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"foo.yaml", "bar.yaml"}, cfg.Providers.File.Filenames)
				require.Equal(t, []string{"foo.yaml,bar.yaml"}, flags.GetStringSlice("providers-file-filename"))
			},
		},
		{
			name: "filename from env var",
			args: func(t *testing.T) []string {
				key := "MOKAPI_PROVIDERS_FILE_FILENAME"
				err := os.Setenv(key, "foo.yaml")
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key)
				})

				return []string{}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"foo.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "filename from env var is not split",
			args: func(t *testing.T) []string {
				key := "MOKAPI_PROVIDERS_FILE_FILENAME"
				err := os.Setenv(key, "foo.yaml,bar.yaml")
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key)
				})

				return []string{}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"foo.yaml", "bar.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "filenames from env var",
			args: func(t *testing.T) []string {
				key := "MOKAPI_PROVIDERS_FILE_FILENAMES"
				err := os.Setenv(key, "foo.yaml,bar.yaml")
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key)
				})

				return []string{}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"foo.yaml", "bar.yaml"}, cfg.Providers.File.Filenames)
				require.Equal(t, []string{"foo.yaml", "bar.yaml"}, flags.GetStringSlice("providers-file-filenames"))
			},
		},
		{
			name: "filenames from config file",
			args: func(t *testing.T) []string {
				temp := t.TempDir()
				f := path.Join(temp, "cfg.json")
				err := os.WriteFile(f, []byte(`{
"providers": {
  "file": {
    "filenames": ["foo.yaml"]
  }
}}`), 0644)
				require.NoError(t, err)

				return []string{"--config-file", f}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"foo.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "directory from env var",
			args: func(t *testing.T) []string {
				key := "MOKAPI_PROVIDERS_FILE_DIRECTORIES"
				err := os.Setenv(key, "/foo,/bar")
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key)
				})

				return []string{}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"/foo", "/bar"}, cfg.Providers.File.Directories)
			},
		},
		{
			name: "directories from config file",
			args: func(t *testing.T) []string {
				temp := t.TempDir()
				f := path.Join(temp, "cfg.json")
				err := os.WriteFile(f, []byte(`{
"providers": {
  "file": {
    "directories": ["/foo", "/bar"]
  }
}}`), 0644)
				require.NoError(t, err)

				return []string{"--config-file", f}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"/foo", "/bar"}, cfg.Providers.File.Directories)
			},
		},
		{
			name: "SkipPrefix from cli",
			args: func(t *testing.T) []string {
				return []string{"--providers-file-skip-prefix", "foo_,bar_"}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"foo_", "bar_"}, cfg.Providers.File.SkipPrefix)
			},
		},
		{
			name: "SkipPrefix from env var",
			args: func(t *testing.T) []string {
				key := "MOKAPI_PROVIDERS_FILE_SKIP_PREFIX"
				err := os.Setenv(key, "foo_,bar_")
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key)
				})

				return []string{}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"foo_", "bar_"}, cfg.Providers.File.SkipPrefix)
			},
		},
		{
			name: "SkipPrefix from config file",
			args: func(t *testing.T) []string {
				temp := t.TempDir()
				f := path.Join(temp, "cfg.json")
				err := os.WriteFile(f, []byte(`{
"providers": {
  "file": {
    "skipPrefix": ["/foo", "/bar"]
  }
}}`), 0644)
				require.NoError(t, err)

				return []string{"--config-file", f}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"/foo", "/bar"}, cfg.Providers.File.SkipPrefix)
			},
		},
		{
			name: "skipPrefix single element appends to default value",
			args: func(t *testing.T) []string {
				return []string{"--providers-file-skip-prefix", "foo"}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Len(t, cfg.Providers.File.SkipPrefix, 1)
				require.Contains(t, cfg.Providers.File.SkipPrefix, "foo")
			},
		},
		{
			name: "Include from cli",
			args: func(t *testing.T) []string {
				return []string{"--providers-file-include", "foo_,bar_"}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"foo_", "bar_"}, cfg.Providers.File.Include)
			},
		},
		{
			name: "Include from env var",
			args: func(t *testing.T) []string {
				key := "MOKAPI_PROVIDERS_FILE_INCLUDE"
				err := os.Setenv(key, "foo_,bar_")
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key)
				})

				return []string{}
			},
			test: func(t *testing.T, cfg *static.Config, flags *cli.FlagSet) {
				require.Equal(t, []string{"foo_", "bar_"}, cfg.Providers.File.Include)
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

func TestRoot_Providers_Git(t *testing.T) {
	testcases := []struct {
		name string
		args func(t *testing.T) []string
		test func(t *testing.T, cfg *static.Config)
	}{
		{
			name: "url from cli",
			args: func(t *testing.T) []string {
				return []string{"--providers-git-url", "https://foo.git", "--providers-git-url", "https://bar.git"}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"https://foo.git", "https://bar.git"}, cfg.Providers.Git.Urls)
			},
		},
		{
			name: "url from env var",
			args: func(t *testing.T) []string {
				key := "MOKAPI_PROVIDERS_GIT_URLS"
				err := os.Setenv(key, "https://foo.git,https://bar.git")
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key)
				})

				return []string{}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"https://foo.git", "https://bar.git"}, cfg.Providers.Git.Urls)
			},
		},
		{
			name: "filenames from env var",
			args: func(t *testing.T) []string {
				key := "MOKAPI_PROVIDERS_GIT_URLS"
				err := os.Setenv(key, "https://foo.git,https://bar.git")
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key)
				})

				return []string{}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []string{"https://foo.git", "https://bar.git"}, cfg.Providers.Git.Urls)
			},
		},
		{
			name: "PullInterval from env var",
			args: func(t *testing.T) []string {
				key := "MOKAPI_PROVIDERS_GIT_PULL_INTERVAL"
				err := os.Setenv(key, "12m")
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key)
				})

				return []string{}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "12m", cfg.Providers.Git.PullInterval)
			},
		},
		{
			name: "PullInterval from config file",
			args: func(t *testing.T) []string {
				temp := t.TempDir()
				f := path.Join(temp, "cfg.json")
				err := os.WriteFile(f, []byte(`{
"providers": {
  "git": {
    "pullInterval": "12m"
  }
}}`), 0644)
				require.NoError(t, err)

				return []string{"--config-file", f}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "12m", cfg.Providers.Git.PullInterval)
			},
		},
		{
			name: "PullInterval from cli",
			args: func(t *testing.T) []string {
				return []string{"--providers-git-pull-interval", "12m"}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "12m", cfg.Providers.Git.PullInterval)
			},
		},
		{
			name: "TempDir from env var",
			args: func(t *testing.T) []string {
				key := "MOKAPI_PROVIDERS_GIT_TEMP_DIR"
				err := os.Setenv(key, "/foo")
				require.NoError(t, err)
				t.Cleanup(func() {
					_ = os.Unsetenv(key)
				})

				return []string{}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "/foo", cfg.Providers.Git.TempDir)
			},
		},
		{
			name: "TempDir from config file",
			args: func(t *testing.T) []string {
				temp := t.TempDir()
				f := path.Join(temp, "cfg.json")
				err := os.WriteFile(f, []byte(`{
"providers": {
  "git": {
    "tempDir": "/foo"
  }
}}`), 0644)
				require.NoError(t, err)

				return []string{"--config-file", f}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, "/foo", cfg.Providers.Git.TempDir)
			},
		},
		{
			name: "Repository from cli",
			args: func(t *testing.T) []string {
				return []string{"--providers-git-repository", "url=foo,pullInterval=5m", "--providers-git-repository", "url=bar,pullInterval=5h"}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.GitRepo{
					{Url: "foo", PullInterval: "5m"},
					{Url: "bar", PullInterval: "5h"},
				}, cfg.Providers.Git.Repositories)
			},
		},
		{
			name: "Repositories from cli",
			args: func(t *testing.T) []string {
				return []string{"--providers-git-repositories", "url=foo,pullInterval=5m url=bar,pullInterval=5h"}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.GitRepo{
					{Url: "foo", PullInterval: "5m"},
					{Url: "bar", PullInterval: "5h"},
				}, cfg.Providers.Git.Repositories)
			},
		},
		{
			name: "Repositories with index and using shorthand",
			args: func(t *testing.T) []string {
				return []string{"--providers-git-repositories[0]", "url=foo,pullInterval=5m"}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.GitRepo{
					{Url: "foo", PullInterval: "5m"},
				}, cfg.Providers.Git.Repositories)
			},
		},
		{
			name: "set repository's URL on index",
			args: func(t *testing.T) []string {
				return []string{"--providers-git-repositories[0]-url", "foo"}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.GitRepo{
					{Url: "foo"},
				}, cfg.Providers.Git.Repositories)
			},
		},
		{
			name: "set repository's Files on index",
			args: func(t *testing.T) []string {
				return []string{"--providers-git-repositories[0]-files", "foo bar"}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.GitRepo{
					{Files: []string{"foo", "bar"}},
				}, cfg.Providers.Git.Repositories)
			},
		},
		{
			name: "set repository's Files on index using explode",
			args: func(t *testing.T) []string {
				return []string{"--providers-git-repositories[0]-file", "foo"}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.GitRepo{
					{Files: []string{"foo"}},
				}, cfg.Providers.Git.Repositories)
			},
		},
		{
			name: "set repository's include on index",
			args: func(t *testing.T) []string {
				return []string{"--providers-git-repositories[0]-include", "foo bar"}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.GitRepo{
					{Include: []string{"foo", "bar"}},
				}, cfg.Providers.Git.Repositories)
			},
		},
		{
			name: "set repository's auth on index using shorthand",
			args: func(t *testing.T) []string {
				return []string{"--providers-git-repositories[0]-auth-github", "appId=123,installationId=456,privateKey=key"}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, &static.GitHubAuth{
					AppId:          123,
					InstallationId: 456,
					PrivateKey:     "key",
				}, cfg.Providers.Git.Repositories[0].Auth.GitHub)
			},
		},
		{
			name: "auth github in config file",
			args: func(t *testing.T) []string {
				temp := t.TempDir()
				f := path.Join(temp, "cfg.yaml")
				err := os.WriteFile(f, []byte(`
providers:
  git:
    repositories:
      - url: foo
        auth:
          github:
            appId: 1001042
`), 0644)
				require.NoError(t, err)

				return []string{"--config-file", f}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, &static.GitHubAuth{
					AppId: 1001042,
				}, cfg.Providers.Git.Repositories[0].Auth.GitHub)
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

			tc.test(t, cfg)
		})
	}
}

func TestRoot_Providers_Npm(t *testing.T) {
	testcases := []struct {
		name string
		args func(t *testing.T) []string
		test func(t *testing.T, cfg *static.Config)
	}{
		{
			name: "npm packages from file",
			args: func(t *testing.T) []string {
				temp := t.TempDir()
				f := path.Join(temp, "cfg.json")
				err := os.WriteFile(f, []byte(`{
"providers": {
  "npm": {
    "packages": [
		{
			"name": "foo",
			"files": ["dist/foo.json"]
		}
	]
  }
}}`), 0644)
				require.NoError(t, err)

				return []string{"--config-file", f}
			},
			test: func(t *testing.T, cfg *static.Config) {
				require.Equal(t, []static.NpmPackage{
					{
						Name:    "foo",
						Files:   []string{"dist/foo.json"},
						Include: []string(nil),
					},
				}, cfg.Providers.Npm.Packages)
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

			tc.test(t, cfg)
		})
	}
}
