package cli_test

import (
	"mokapi/pkg/cli"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlagBool(t *testing.T) {
	testcases := []struct {
		name string
		cmd  func() *cli.Command
		args []string
		test func(t *testing.T, cmd *cli.Command, args []string, err error)
	}{
		{
			name: "--foo",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().Bool("foo", false, cli.FlagDoc{})
				return c
			},
			args: []string{"--foo"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, true, cmd.Flags().GetBool("foo"))
			},
		},
		{
			name: "default is used",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().Bool("foo", true, cli.FlagDoc{})
				return c
			},
			args: []string{""},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, true, cmd.Flags().GetBool("foo"))
			},
		},
		{
			name: "--no-foo",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().Bool("foo", true, cli.FlagDoc{})
				return c
			},
			args: []string{"--no-foo"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, false, cmd.Flags().GetBool("foo"))
			},
		},
		{
			name: "--no-foo false",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().Bool("foo", true, cli.FlagDoc{})
				return c
			},
			args: []string{"--no-foo false"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, true, cmd.Flags().GetBool("foo"))
			},
		},
		{
			name: "--no-foo=false",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().Bool("foo", true, cli.FlagDoc{})
				return c
			},
			args: []string{"--no-foo=false"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, true, cmd.Flags().GetBool("foo"))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var cmd *cli.Command
			var args []string
			root := tc.cmd()
			root.SetArgs(tc.args)
			root.Run = func(c *cli.Command, a []string) error {
				cmd = c
				args = a
				return nil
			}

			err := root.Execute()

			tc.test(t, cmd, args, err)
		})
	}
}
