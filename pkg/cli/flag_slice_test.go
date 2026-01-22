package cli_test

import (
	"mokapi/pkg/cli"
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
		cmd  func() *cli.Command
		args []string
		test func(t *testing.T, cmd *cli.Command, args []string, err error)
	}{
		{
			name: "no default",
			cmd: func() *cli.Command {
				c := &cli.Command{Name: "foo", Config: &config{}}
				c.Flags().StringSlice("foo", nil, false, cli.FlagDoc{})
				return c
			},
			args: []string{"--foo a b c"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b", "c"}, cmd.Flags().GetStringSlice("foo"))
				require.Equal(t, &config{Foo: []string{"a", "b", "c"}}, cmd.Config)
			},
		},
		{
			name: "with default",
			cmd: func() *cli.Command {
				c := &cli.Command{Name: "foo", Config: &config{}}
				c.Flags().StringSlice("foo", []string{"zzz"}, false, cli.FlagDoc{})
				return c
			},
			args: []string{"--foo a b c"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b", "c"}, cmd.Flags().GetStringSlice("foo"))
				require.Equal(t, &config{Foo: []string{"a", "b", "c"}}, cmd.Config)
			},
		},
		{
			name: "value should be split",
			cmd: func() *cli.Command {
				c := &cli.Command{Name: "foo", Config: &config{}}
				c.Flags().StringSlice("foo", []string{"zzz"}, false, cli.FlagDoc{})
				return c
			},
			args: []string{"--foo", "a,b,c"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b", "c"}, cmd.Flags().GetStringSlice("foo"))
				require.Equal(t, &config{Foo: []string{"a", "b", "c"}}, cmd.Config)
			},
		},
		{
			name: "explode",
			cmd: func() *cli.Command {
				c := &cli.Command{Name: "foo", Config: &config{}}
				c.Flags().StringSlice("foo", nil, true, cli.FlagDoc{})
				return c
			},
			args: []string{"--foo", "a", "--foo", "b", "--foo", "c"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b", "c"}, cmd.Flags().GetStringSlice("foo"))
				require.Equal(t, &config{Foo: []string{"a", "b", "c"}}, cmd.Config)
			},
		},
		{
			name: "explode tag",
			cmd: func() *cli.Command {
				c := &cli.Command{Name: "foo", Config: &config{}}
				c.Flags().StringSlice("explode", nil, true, cli.FlagDoc{})
				return c
			},
			args: []string{"--explode", "a", "--explode", "b", "--explode", "c"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b", "c"}, cmd.Flags().GetStringSlice("explode"))
				require.Equal(t, &config{Explodes: []string{"a", "b", "c"}}, cmd.Config)
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
