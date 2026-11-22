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

func TestHelp_Subcommand(t *testing.T) {
	newRoot := func() *cli.Command {
		root := &cli.Command{
			Name:  "app",
			Short: "Root description",
			Use:   "app [flags]",
		}
		root.Flags().String("root-flag", "", cli.FlagDoc{Short: "Only for root"})

		child := &cli.Command{
			Name:    "child",
			Short:   "Child short",
			Long:    "Child long description",
			Use:     "app child [flags]",
			Example: "  app child --count 2",
			Run:     func(cmd *cli.Command, args []string) error { return nil },
		}
		child.Flags().Int("count", 1, cli.FlagDoc{Short: "Number of items"})
		root.Commands = append(root.Commands, child)
		return root
	}

	t.Run("child --help shows help of the child", func(t *testing.T) {
		root := newRoot()
		root.SetArgs([]string{"child", "--help"})
		var sb strings.Builder
		root.SetOutput(&sb)

		require.NoError(t, root.Execute())

		out := sb.String()
		require.Contains(t, out, "Child long description")
		require.Contains(t, out, "app child [flags]")
		require.Contains(t, out, "--count")
		require.Contains(t, out, "Number of items")
		require.Contains(t, out, "\nExamples:\n  app child --count 2\n")
		require.NotContains(t, out, "Commands:")
		require.NotContains(t, out, "--root-flag")
		require.NotContains(t, out, "Root description")
	})

	t.Run("child -h shows help of the child", func(t *testing.T) {
		root := newRoot()
		root.SetArgs([]string{"child", "-h"})
		var sb strings.Builder
		root.SetOutput(&sb)

		require.NoError(t, root.Execute())

		require.Contains(t, sb.String(), "--count")
		require.NotContains(t, sb.String(), "--root-flag")
	})

	t.Run("root --help lists commands", func(t *testing.T) {
		root := newRoot()
		root.SetArgs([]string{"--help"})
		var sb strings.Builder
		root.SetOutput(&sb)

		require.NoError(t, root.Execute())

		out := sb.String()
		require.Contains(t, out, "\nCommands:\n  child  Child short\n")
		require.Contains(t, out, "--root-flag")
	})

	t.Run("child keeps its own output writer", func(t *testing.T) {
		root := newRoot()
		root.SetArgs([]string{"child", "--help"})
		var rootOut, childOut strings.Builder
		root.SetOutput(&rootOut)
		root.Commands[0].SetOutput(&childOut)

		require.NoError(t, root.Execute())

		require.Empty(t, rootOut.String())
		require.Contains(t, childOut.String(), "--count")
	})

	t.Run("example without children has no Commands section", func(t *testing.T) {
		cmd := &cli.Command{Name: "app", Example: "\n  app run\n\n"}
		cmd.SetArgs([]string{"--help"})
		var sb strings.Builder
		cmd.SetOutput(&sb)

		require.NoError(t, cmd.Execute())

		require.Contains(t, sb.String(), "\nExamples:\n  app run\n")
		require.NotContains(t, sb.String(), "Commands:")
	})
}
