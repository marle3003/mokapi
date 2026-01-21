package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlagSlice(t *testing.T) {
	type config struct {
		Foo      []string
		Explodes []string `explode:"explode"`
	}

	testcases := []struct {
		name string
		cmd  func() *Command
		args []string
		test func(t *testing.T, cmd *Command, args []string, err error)
	}{
		{
			name: "no default",
			cmd: func() *Command {
				c := &Command{Name: "foo", Config: &config{}}
				c.Flags().StringSlice("foo", nil, false, FlagDoc{})
				return c
			},
			args: []string{"--foo a b c"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b", "c"}, cmd.Flags().GetStringSlice("foo"))
				require.Equal(t, &config{Foo: []string{"a", "b", "c"}}, cmd.Config)
			},
		},
		{
			name: "with default",
			cmd: func() *Command {
				c := &Command{Name: "foo", Config: &config{}}
				c.Flags().StringSlice("foo", []string{"zzz"}, false, FlagDoc{})
				return c
			},
			args: []string{"--foo a b c"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b", "c"}, cmd.Flags().GetStringSlice("foo"))
				require.Equal(t, &config{Foo: []string{"a", "b", "c"}}, cmd.Config)
			},
		},
		{
			name: "value should be split",
			cmd: func() *Command {
				c := &Command{Name: "foo", Config: &config{}}
				c.Flags().StringSlice("foo", []string{"zzz"}, false, FlagDoc{})
				return c
			},
			args: []string{"--foo", "a,b,c"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b", "c"}, cmd.Flags().GetStringSlice("foo"))
				require.Equal(t, &config{Foo: []string{"a", "b", "c"}}, cmd.Config)
			},
		},
		{
			name: "explode",
			cmd: func() *Command {
				c := &Command{Name: "foo", Config: &config{}}
				c.Flags().StringSlice("foo", nil, true, FlagDoc{})
				return c
			},
			args: []string{"--foo", "a", "--foo", "b", "--foo", "c"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b", "c"}, cmd.Flags().GetStringSlice("foo"))
				require.Equal(t, &config{Foo: []string{"a", "b", "c"}}, cmd.Config)
			},
		},
		{
			name: "explode tag",
			cmd: func() *Command {
				c := &Command{Name: "foo", Config: &config{}}
				c.Flags().StringSlice("explode", nil, true, FlagDoc{})
				return c
			},
			args: []string{"--explode", "a", "--explode", "b", "--explode", "c"},
			test: func(t *testing.T, cmd *Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b", "c"}, cmd.Flags().GetStringSlice("explode"))
				require.Equal(t, &config{Explodes: []string{"a", "b", "c"}}, cmd.Config)
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
