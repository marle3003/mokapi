package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommand(t *testing.T) {
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
				c := &Command{Name: "foo"}
				c.Flags().Bool("foo", false, "")
				return c
			},
			args: []string{"--foo"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cmd.Name)
				require.Equal(t, true, cmd.Flags().GetBool("foo"))
			},
		},
		{
			name: "-f",
			cmd: func() *Command {
				c := &Command{Name: "foo"}
				c.Flags().BoolShort("foo", "f", false, "")
				return c
			},
			args: []string{"-f"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cmd.Name)
				require.Equal(t, true, cmd.Flags().GetBool("foo"))
			},
		},
		{
			name: "bind to config",
			cmd: func() *Command {
				c := &Command{Config: &config{}}
				c.Flags().Bool("flag", false, "")
				return c
			},
			args: []string{"--flag"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, true, cmd.Flags().GetBool("flag"))
				require.Equal(t, true, cmd.Config.(*config).Flag)
			},
		},
		{
			name: "--count",
			cmd: func() *Command {
				c := &Command{Config: &config{}}
				c.Flags().Int("count", 12, "")
				return c
			},
			args: []string{"--count", "10"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, 10, cmd.Flags().GetInt("count"))
			},
		},
		{
			name: "--count default",
			cmd: func() *Command {
				c := &Command{Config: &config{}}
				c.Flags().Int("count", 12, "")
				return c
			},
			args: []string{},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, cmd.Flags().GetInt("count"))
			},
		},
		{
			name: "--skip-prefix",
			cmd: func() *Command {
				c := &Command{Config: &config{}}
				c.Flags().StringSlice("skip-prefix", []string{"_"}, "", false)
				return c
			},
			args: []string{"--skip-prefix", "_", "foo_"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"_", "foo_"}, cmd.Flags().GetStringSlice("skip-prefix"))
				require.Equal(t, []string{"_", "foo_"}, cmd.Config.(*config).SkipPrefix)
			},
		},
		{
			name: "--skip-prefix default",
			cmd: func() *Command {
				c := &Command{Config: &config{}}
				c.Flags().StringSlice("skip-prefix", []string{"_"}, "", false)
				return c
			},
			args: []string{},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"_"}, cmd.Flags().GetStringSlice("skip-prefix"))
				require.Equal(t, []string{"_"}, cmd.Config.(*config).SkipPrefix)
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
