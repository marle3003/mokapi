package static_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/decoders"
	"mokapi/config/static"
	"os"
	"testing"
)

func TestGitConfig(t *testing.T) {
	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "assign with =",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, `--ConfigFile=foo`)

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.ConfigFile)
			},
		},
		{
			name: "assign without =",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, `--ConfigFile`, "foo")

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.ConfigFile)
			},
		},
		{
			name: "help",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, `--help`)

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)
				require.Equal(t, true, cfg.Help)
			},
		},
		{
			name: "json",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, `--providers.file={"filename":"foo.yaml","directory":"foo", "skipPrefix":["_"]}`)

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)
				require.Equal(t, []string{"foo.yaml"}, cfg.Providers.File.Filenames)
				require.Equal(t, []string{"foo"}, cfg.Providers.File.Directories)
				require.Equal(t, []string{"_"}, cfg.Providers.File.SkipPrefix)
			},
		},
		{
			name: "shorthand object",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, `--providers.file`, "filename=foo.yaml")

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)
				require.Equal(t, []string{"foo.yaml"}, cfg.Providers.File.Filenames)
			},
		},
		{
			name: "args",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--providers.git.repositories[0].url=https://github.com/PATH-TO/REPOSITORY?ref=branch-name")
				os.Args = append(os.Args, "--providers.git.repositories[0].pullInterval=5m")
				os.Args = append(os.Args, "--providers.git.repositories[1].pullInterval=5h")

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)
				require.Equal(t, "https://github.com/PATH-TO/REPOSITORY?ref=branch-name", cfg.Providers.Git.Repositories[0].Url)
				require.Equal(t, "5m", cfg.Providers.Git.Repositories[0].PullInterval)
				require.Equal(t, "", cfg.Providers.Git.Repositories[1].Url)
				require.Equal(t, "5h", cfg.Providers.Git.Repositories[1].PullInterval)
			},
		},
		{
			name: "shorthand array",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, `--providers.git.repositories`, "url=foo,pullInterval=5m url=bar,pullInterval=5h")

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)
				require.Len(t, cfg.Providers.Git.Repositories, 2)
				require.Equal(t, "foo", cfg.Providers.Git.Repositories[0].Url)
				require.Equal(t, "5m", cfg.Providers.Git.Repositories[0].PullInterval)
				require.Equal(t, "bar", cfg.Providers.Git.Repositories[1].Url)
				require.Equal(t, "5h", cfg.Providers.Git.Repositories[1].PullInterval)
			},
		},
		{
			name: "explode with json",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, `--providers.git.repository={"url":"https://github.com/PATH-TO/REPOSITORY?ref=branch-name","pullInterval":"5m"}`)

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)
				require.Equal(t, "https://github.com/PATH-TO/REPOSITORY?ref=branch-name", cfg.Providers.Git.Repositories[0].Url)
				require.Equal(t, "5m", cfg.Providers.Git.Repositories[0].PullInterval)
			},
		},
		{
			name: "env var",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				err := os.Setenv("MOKAPI_Providers_GIT_Repositories[0]_Url", "https://github.com/PATH-TO/REPOSITORY")
				require.NoError(t, err)
				defer os.Unsetenv("MOKAPI_Providers_GIT_Repositories[0]_Url")
				err = os.Setenv("MOKAPI_Providers_GIT_Repositories[0]_PullInterval", "3m")
				require.NoError(t, err)
				defer os.Unsetenv("MOKAPI_Providers_GIT_Repositories[0]_PullInterval")

				cfg := static.Config{}
				err = decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Equal(t, "https://github.com/PATH-TO/REPOSITORY", cfg.Providers.Git.Repositories[0].Url)
				require.Equal(t, "3m", cfg.Providers.Git.Repositories[0].PullInterval)
			},
		},
		{
			name: "file provider include",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--Providers.file.include", `mokapi/**/*.json mokapi/**/*.yaml "foo bar/**/*.yaml"`)

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Equal(t, []string{"mokapi/**/*.json", "mokapi/**/*.yaml", "foo bar/**/*.yaml"}, cfg.Providers.File.Include)
			},
		},
		{
			name: "file provider include with space",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--Providers.file.include", `"C:\Documents and Settings\" C:\Work"`)

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Equal(t, []string{"C:\\Documents and Settings\\", "C:\\Work"}, cfg.Providers.File.Include)
			},
		},
		{
			name: "file provider include twice",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--Providers.file.include", "foo", "--Providers.file.include", "bar")

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Equal(t, []string{"foo", "bar"}, cfg.Providers.File.Include)
			},
		},
		{
			name: "file provider include overwrite",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--Providers.file.include", "foo", "--Providers.file.include[0]", "bar")

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Equal(t, []string{"bar"}, cfg.Providers.File.Include)
			},
		},
		{
			name: "git provider set url",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--Providers.git.Url", `foo`)

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Equal(t, []string{"foo"}, cfg.Providers.Git.Urls)
			},
		},
		{
			name: "git provider set urls",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--Providers.git.Urls", `foo`)

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Equal(t, []string{"foo"}, cfg.Providers.Git.Urls)
			},
		},
		{
			name: "http provider set url",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--Providers.Http.Url", `foo`)

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Equal(t, []string{"foo"}, cfg.Providers.Http.Urls)
			},
		},
		{
			name: "http provider set urls using explode",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--Providers.Http.Url", `foo`)
				os.Args = append(os.Args, "--Providers.Http.Url", `bar`)

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Equal(t, []string{"foo", "bar"}, cfg.Providers.Http.Urls)
			},
		},
		{
			name: "http provider set urls",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--Providers.Http.Urls", `foo bar`)

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Equal(t, []string{"foo", "bar"}, cfg.Providers.Http.Urls)
			},
		},
		{
			name: "http provider",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--Providers.Http", `urls=foo bar,pollInterval=5s,pollTimeout=30s,proxy=bar,tlsSkipVerify=true`)

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Equal(t, []string{"foo", "bar"}, cfg.Providers.Http.Urls)
				require.Equal(t, "5s", cfg.Providers.Http.PollInterval)
				require.Equal(t, "30s", cfg.Providers.Http.PollTimeout)
				require.Equal(t, true, cfg.Providers.Http.TlsSkipVerify)
				require.Equal(t, "bar", cfg.Providers.Http.Proxy)
			},
		},
		{
			name: "config",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--config", `{"openapi": "3.0"}`)

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Len(t, cfg.Configs, 1)
				require.Equal(t, "{\"openapi\": \"3.0\"}", cfg.Configs[0])
			},
		},
		{
			name: "configfile",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe", "--configfile", "foo.json")

				read := func(path string) ([]byte, error) {
					return []byte(`{"configs": [ { "openapi": "3.0", "info": { "name": "foo" } } ]}`), nil
				}

				cfg := static.Config{}
				err := decoders.Load([]decoders.ConfigDecoder{decoders.NewFileDecoder(read), decoders.NewFlagDecoder()}, &cfg)
				require.NoError(t, err)

				require.Len(t, cfg.Configs, 1)
				require.Equal(t, `{ "openapi": "3.0", "info": { "name": "foo" } }`, cfg.Configs[0])
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			os.Args = nil
			tc.test(t)
		})
	}
}
