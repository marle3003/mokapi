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
		f    func(t *testing.T)
	}{
		{
			name: "args",
			f: func(t *testing.T) {
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
			name: "env var",
			f: func(t *testing.T) {
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
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			os.Args = nil
			tc.f(t)
		})
	}
}
