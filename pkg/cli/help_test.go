package cli_test

import (
	"mokapi/pkg/cli"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelp(t *testing.T) {
	testcases := []struct {
		name string
		cmd  *cli.Command
		test func(t *testing.T, out string)
	}{
		{
			name: "no flags",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				return c
			}(),
			test: func(t *testing.T, out string) {
				require.Equal(t, "\nFlags:\n  -h, --help   Show help information and exit\n", out)
			},
		},
		{
			name: "Long description",
			cmd: func() *cli.Command {
				c := &cli.Command{Long: "Long Description"}
				return c
			}(),
			test: func(t *testing.T, out string) {
				require.Equal(t, "\n\nLong Description\n\nFlags:\n  -h, --help   Show help information and exit\n", out)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := tc.cmd
			root.SetArgs([]string{"--help"})
			var sb strings.Builder
			root.SetOutput(&sb)

			err := root.Execute()
			require.NoError(t, err)

			tc.test(t, sb.String())
		})
	}
}
