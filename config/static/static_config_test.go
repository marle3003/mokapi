package static

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/decoders"
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

				cfg := Config{}
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

				cfg := Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)
				require.Equal(t, "foo", cfg.ConfigFile)
			},
		},
		{
			name: "json",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, `--providers.file={"filename":"foo.yaml","directory":"foo", "skipPrefix":["_"]}`)

				cfg := Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)
				require.Equal(t, "foo.yaml", cfg.Providers.File.Filename)
				require.Equal(t, "foo", cfg.Providers.File.Directory)
				require.Equal(t, []string{"_"}, cfg.Providers.File.SkipPrefix)
			},
		},
		{
			name: "shorthand object",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, `--providers.file`, "filename=foo.yaml")

				cfg := Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)
				require.Equal(t, "foo.yaml", cfg.Providers.File.Filename)
			},
		},
		{
			name: "args",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--providers.git.repositories[0].url=https://github.com/PATH-TO/REPOSITORY?ref=branch-name")
				os.Args = append(os.Args, "--providers.git.repositories[0].pullInterval=5m")
				os.Args = append(os.Args, "--providers.git.repositories[1].pullInterval=5h")

				cfg := Config{}
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

				cfg := Config{}
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

				cfg := Config{}
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

				cfg := Config{}
				err = decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Equal(t, "https://github.com/PATH-TO/REPOSITORY", cfg.Providers.Git.Repositories[0].Url)
				require.Equal(t, "3m", cfg.Providers.Git.Repositories[0].PullInterval)
			},
		},
		{
			name: "config",
			test: func(t *testing.T) {
				os.Args = append(os.Args, "mokapi.exe")
				os.Args = append(os.Args, "--config", `{"openapi": "3.0"}`)

				cfg := Config{}
				err := decoders.Load([]decoders.ConfigDecoder{&decoders.FlagDecoder{}}, &cfg)
				require.NoError(t, err)

				require.Len(t, cfg.Configs, 1)
				require.Equal(t, "{\"openapi\": \"3.0\"}", cfg.Configs[0])
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
