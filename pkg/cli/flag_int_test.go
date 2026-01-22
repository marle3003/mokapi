package cli_test

import (
	"mokapi/pkg/cli"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlagInt(t *testing.T) {
	type config struct {
		Foo int
		Bar int64
		Any interface{}
	}

	testcases := []struct {
		name string
		cmd  func() *cli.Command
		args []string
		test func(t *testing.T, cmd *cli.Command, args []string, err error)
	}{
		{
			name: "int value",
			cmd: func() *cli.Command {
				c := &cli.Command{Name: "foo", Config: &config{}}
				c.Flags().Int("foo", 0, cli.FlagDoc{})
				return c
			},
			args: []string{"--foo 12"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, cmd.Flags().GetInt("foo"))
				require.Equal(t, &config{Foo: 12}, cmd.Config)
			},
		},
		{
			name: "float value",
			cmd: func() *cli.Command {
				c := &cli.Command{Name: "foo", Config: &config{}}
				c.Flags().Int("foo", 0, cli.FlagDoc{})
				return c
			},
			args: []string{"--foo 12.4"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.EqualError(t, err, "failed to set flag foo: parsing 12.4: invalid syntax")
			},
		},
		{
			name: "bind to int64",
			cmd: func() *cli.Command {
				c := &cli.Command{Name: "foo", Config: &config{}}
				c.Flags().Int("bar", 0, cli.FlagDoc{})
				return c
			},
			args: []string{"--bar 12"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, cmd.Flags().GetInt("bar"))
				require.Equal(t, &config{Bar: 12}, cmd.Config)
			},
		},
		{
			name: "bind to interface{}",
			cmd: func() *cli.Command {
				c := &cli.Command{Name: "foo", Config: &config{}}
				c.Flags().Int("any", 0, cli.FlagDoc{})
				return c
			},
			args: []string{"--any 12"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, cmd.Flags().GetInt("any"))
				require.Equal(t, &config{Any: 12}, cmd.Config)
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
