package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlagBool(t *testing.T) {
	type config struct {
		Flag       bool
		Name       string
		SkipPrefix []string `flag:"skip-prefix"`
	}

	testcases := []struct {
		name string
		cmd  func() *Command
		args []string
		test func(t *testing.T, cmd *Command, args []string, err error)
	}{
		{
			name: "--foo",
			cmd: func() *Command {
				c := &Command{}
				c.Flags().Bool("foo", false, FlagDoc{})
				return c
			},
			args: []string{"--foo"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, true, cmd.Flags().GetBool("foo"))
			},
		},
		{
			name: "default is used",
			cmd: func() *Command {
				c := &Command{}
				c.Flags().Bool("foo", true, FlagDoc{})
				return c
			},
			args: []string{""},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, true, cmd.Flags().GetBool("foo"))
			},
		},
		{
			name: "--no-foo",
			cmd: func() *Command {
				c := &Command{}
				c.Flags().Bool("foo", true, FlagDoc{})
				return c
			},
			args: []string{"--no-foo"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, false, cmd.Flags().GetBool("foo"))
			},
		},
		{
			name: "--no-foo false",
			cmd: func() *Command {
				c := &Command{}
				c.Flags().Bool("foo", true, FlagDoc{})
				return c
			},
			args: []string{"--no-foo false"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, true, cmd.Flags().GetBool("foo"))
			},
		},
		{
			name: "--no-foo=false",
			cmd: func() *Command {
				c := &Command{}
				c.Flags().Bool("foo", true, FlagDoc{})
				return c
			},
			args: []string{"--no-foo=false"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, true, cmd.Flags().GetBool("foo"))
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var cmd *Command
			var args []string
			root := tc.cmd()
			root.SetArgs(tc.args)
			root.Run = func(c *Command, a []string) error {
				cmd = c
				args = a
				return nil
			}

			err := root.Execute()

			tc.test(t, cmd, args, err)
		})
	}
}
