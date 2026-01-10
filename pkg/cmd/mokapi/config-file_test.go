package mokapi_test

import (
	"mokapi/config/static"
	"mokapi/pkg/cli"
	"mokapi/pkg/cmd/mokapi"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type testFileReader struct {
	files map[string][]byte
}

func (t *testFileReader) ReadFile(name string) ([]byte, error) {
	if b, ok := t.files[name]; ok {
		return b, nil
	}
	// if test is executed on windows we get second path
	name = strings.ReplaceAll(name, "\\", "/")
	if b, ok := t.files[name]; ok {
		return b, nil
	}
	return nil, os.ErrNotExist
}

func (t *testFileReader) FileExists(name string) bool {
	if _, ok := t.files[name]; ok {
		return true
	}
	name = strings.ReplaceAll(name, "\\", "/")
	_, ok := t.files[name]
	return ok
}

func TestFileDecoder_Decode(t *testing.T) {
	newCmd := func(args []string) *cli.Command {
		c := mokapi.NewCmdMokapi()
		c.SetArgs(args)
		c.Run = func(cmd *cli.Command, args []string) error {
			return nil
		}
		return c
	}

	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "--config-file",
			test: func(t *testing.T) {
				path := createTempFile(t, "test.yml", "")

				c := newCmd([]string{"--config-file", path})
				err := c.Execute()
				cfg := c.Config.(*static.Config)
				require.NoError(t, err)
				require.Equal(t, path, cfg.ConfigFile)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				cli.SetFileReader(&cli.FileReader{})
			}()

			tc.test(t)
		})
	}
}
