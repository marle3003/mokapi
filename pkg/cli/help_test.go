package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelp(t *testing.T) {
	testcases := []struct {
		name string
		cmd  *Command
		test func(t *testing.T, out string)
	}{
		{
			name: "no flags",
			cmd: func() *Command {
				c := &Command{}
				return c
			}(),
			test: func(t *testing.T, out string) {
				require.Equal(t, "\nFlags:\n --help   Show help information and exit\n", out)
			},
		},
		{
			name: "Long description",
			cmd: func() *Command {
				c := &Command{Long: "Long Description"}
				return c
			}(),
			test: func(t *testing.T, out string) {
				require.Equal(t, "\n\nLong Description\n\nFlags:\n --help   Show help information and exit\n", out)
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
