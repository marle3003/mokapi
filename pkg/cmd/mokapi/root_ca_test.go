package mokapi_test

import (
	"mokapi/config/static"
	"mokapi/pkg/cli"
	"mokapi/pkg/cmd/mokapi"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootCaFlags(t *testing.T) {
	testcases := []struct {
		name string
		cli  []string
		env  map[string]string
		file string
	}{
		{
			name: "cli",
			cli:  []string{"--root-ca-cert", "cert.pem", "--root-ca-key", "key.pem"},
		},
		{
			name: "env",
			env:  map[string]string{"MOKAPI_ROOT_CA_CERT": "cert.pem", "MOKAPI_ROOT_CA_KEY": "key.pem"},
		},
		{
			name: "file",
			file: "rootCaCert: cert.pem\nrootCaKey: key.pem\n",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			args := append([]string{}, tc.cli...)
			if tc.file != "" {
				f := filepath.Join(t.TempDir(), "mokapi.yaml")
				require.NoError(t, os.WriteFile(f, []byte(tc.file), 0644))
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

			require.Equal(t, "cert.pem", cfg.RootCaCert.String())
			require.Equal(t, "key.pem", cfg.RootCaKey.String())
		})
	}
}

func TestRootCaFlagDocs(t *testing.T) {
	cmd := mokapi.NewCmdMokapi()

	cert := cmd.Flags().Lookup("root-ca-cert")
	require.NotNil(t, cert)
	require.Contains(t, strings.ToLower(cert.Short), "certificate")
	require.NotContains(t, strings.ToLower(cert.Short), "private key")
	for _, code := range cert.Examples[0].Codes {
		require.NotContains(t, code.Source, "root-ca-key", code.Title)
	}

	key := cmd.Flags().Lookup("root-ca-key")
	require.NotNil(t, key)
	require.Contains(t, strings.ToLower(key.Short), "private key")
	for _, code := range key.Examples[0].Codes {
		require.NotContains(t, code.Source, "root-ca-cert", code.Title)
	}
}
