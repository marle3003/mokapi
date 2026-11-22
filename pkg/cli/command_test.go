package cli_test

import (
	"mokapi/pkg/cli"
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
		cmd  func() *cli.Command
		args []string
		test func(t *testing.T, cmd *cli.Command, args []string, err error)
	}{
		{
			name: "--foo",
			cmd: func() *cli.Command {
				c := &cli.Command{Name: "foo"}
				c.Flags().Bool("foo", false, cli.FlagDoc{})
				return c
			},
			args: []string{"--foo"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cmd.Name)
				require.Equal(t, true, cmd.Flags().GetBool("foo"))
			},
		},
		{
			name: "-f",
			cmd: func() *cli.Command {
				c := &cli.Command{Name: "foo"}
				c.Flags().BoolShort("foo", "f", false, cli.FlagDoc{})
				return c
			},
			args: []string{"-f"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cmd.Name)
				require.Equal(t, true, cmd.Flags().GetBool("foo"))
			},
		},
		{
			name: "bind to config",
			cmd: func() *cli.Command {
				c := &cli.Command{Config: &config{}}
				c.Flags().Bool("flag", false, cli.FlagDoc{})
				return c
			},
			args: []string{"--flag"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, true, cmd.Flags().GetBool("flag"))
				require.Equal(t, true, cmd.Config.(*config).Flag)
			},
		},
		{
			name: "--count",
			cmd: func() *cli.Command {
				c := &cli.Command{Config: &config{}}
				c.Flags().Int("count", 12, cli.FlagDoc{})
				return c
			},
			args: []string{"--count", "10"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, 10, cmd.Flags().GetInt("count"))
			},
		},
		{
			name: "--count default",
			cmd: func() *cli.Command {
				c := &cli.Command{Config: &config{}}
				c.Flags().Int("count", 12, cli.FlagDoc{})
				return c
			},
			args: []string{},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, 12, cmd.Flags().GetInt("count"))
			},
		},
		{
			name: "--skip-prefix",
			cmd: func() *cli.Command {
				c := &cli.Command{Config: &config{}}
				c.Flags().StringSlice("skip-prefix", []string{"_"}, false, cli.FlagDoc{})
				return c
			},
			args: []string{"--skip-prefix", "_", "foo_"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"_", "foo_"}, cmd.Flags().GetStringSlice("skip-prefix"))
				require.Equal(t, []string{"_", "foo_"}, cmd.Config.(*config).SkipPrefix)
			},
		},
		{
			name: "--skip-prefix default",
			cmd: func() *cli.Command {
				c := &cli.Command{Config: &config{SkipPrefix: []string{"_"}}}
				c.Flags().StringSlice("skip-prefix", []string{"_"}, false, cli.FlagDoc{})
				return c
			},
			args: []string{},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"_"}, cmd.Flags().GetStringSlice("skip-prefix"))
				require.Equal(t, []string{"_"}, cmd.Config.(*config).SkipPrefix)
			},
		},
		{
			name: "--skip-prefix=foo",
			cmd: func() *cli.Command {
				c := &cli.Command{Config: &config{}}
				c.Flags().StringSlice("skip-prefix", []string{"_"}, false, cli.FlagDoc{})
				return c
			},
			args: []string{"--skip-prefix=foo"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"foo"}, cmd.Flags().GetStringSlice("skip-prefix"))
				require.Equal(t, []string{"foo"}, cmd.Config.(*config).SkipPrefix)
			},
		},
		{
			name: "--no-flag",
			cmd: func() *cli.Command {
				c := &cli.Command{Config: &config{}}
				c.Flags().Bool("flag", true, cli.FlagDoc{})
				return c
			},
			args: []string{"--no-flag"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, false, cmd.Flags().GetBool("flag"))
				require.Equal(t, false, cmd.Config.(*config).Flag)
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

func TestCommand_PositionalArgsAfterFlag(t *testing.T) {
	testcases := []struct {
		name string
		cmd  func() *cli.Command
		args []string
		test func(t *testing.T, cmd *cli.Command, args []string, err error)
	}{
		{
			name: "string flag followed by positional argument",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().String("name", "", cli.FlagDoc{})
				return c
			},
			args: []string{"--name", "foo", "pos"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cmd.Flags().GetString("name"))
				require.Equal(t, []string{"pos"}, args)
			},
		},
		{
			name: "string flag followed by several positional arguments",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().String("name", "", cli.FlagDoc{})
				return c
			},
			args: []string{"--name", "foo", "a", "b"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cmd.Flags().GetString("name"))
				require.Equal(t, []string{"a", "b"}, args)
			},
		},
		{
			name: "shorthand string flag followed by positional argument",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().StringShort("name", "n", "", cli.FlagDoc{})
				return c
			},
			args: []string{"-n", "foo", "pos"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cmd.Flags().GetString("name"))
				require.Equal(t, []string{"pos"}, args)
			},
		},
		{
			name: "int flag followed by positional argument",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().Int("count", 1, cli.FlagDoc{})
				return c
			},
			args: []string{"--count", "2", "pos"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, 2, cmd.Flags().GetInt("count"))
				require.Equal(t, []string{"pos"}, args)
			},
		},
		{
			name: "float flag followed by positional argument",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().Float("ratio", 1, cli.FlagDoc{})
				return c
			},
			args: []string{"--ratio", "0.5", "pos"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, 0.5, cmd.Flags().GetFloat("ratio"))
				require.Equal(t, []string{"pos"}, args)
			},
		},
		{
			name: "bool flag without value followed by positional argument",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().Bool("flag", false, cli.FlagDoc{})
				return c
			},
			args: []string{"--flag", "pos"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, true, cmd.Flags().GetBool("flag"))
				require.Equal(t, []string{"pos"}, args)
			},
		},
		{
			name: "bool flag with explicit value followed by positional argument",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().Bool("flag", true, cli.FlagDoc{})
				return c
			},
			args: []string{"--flag", "false", "pos"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, false, cmd.Flags().GetBool("flag"))
				require.Equal(t, []string{"pos"}, args)
			},
		},
		{
			name: "list flag consumes following values (documents current behavior)",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().StringSlice("list", nil, false, cli.FlagDoc{})
				return c
			},
			args: []string{"--list", "a", "b"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b"}, cmd.Flags().GetStringSlice("list"))
				require.Empty(t, args)
			},
		},
		{
			name: "string flag with = followed by positional argument",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().String("name", "", cli.FlagDoc{})
				return c
			},
			args: []string{"--name=foo", "pos"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, "foo", cmd.Flags().GetString("name"))
				require.Equal(t, []string{"pos"}, args)
			},
		},
		{
			name: "bool flag with true followed by positional argument",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().Bool("flag", false, cli.FlagDoc{})
				return c
			},
			args: []string{"--flag", "true", "pos"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, true, cmd.Flags().GetBool("flag"))
				require.Equal(t, []string{"pos"}, args)
			},
		},
		{
			name: "--no-flag followed by positional argument",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().Bool("flag", true, cli.FlagDoc{})
				return c
			},
			args: []string{"--no-flag", "pos"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, false, cmd.Flags().GetBool("flag"))
				require.Equal(t, []string{"pos"}, args)
			},
		},
		{
			name: "shorthand bool flag followed by positional argument",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().BoolShort("flag", "f", false, cli.FlagDoc{})
				return c
			},
			args: []string{"-f", "pos"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, true, cmd.Flags().GetBool("flag"))
				require.Equal(t, []string{"pos"}, args)
			},
		},
		{
			name: "string flag at the end without value",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().String("name", "x", cli.FlagDoc{})
				return c
			},
			args: []string{"--name"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, "", cmd.Flags().GetString("name"))
				require.Empty(t, args)
			},
		},
		{
			name: "dynamic string flag followed by positional argument",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().DynamicString("foo[<index>]", cli.FlagDoc{})
				return c
			},
			args: []string{"--foo[0]", "bar", "pos"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, "bar", cmd.Flags().GetString("foo[0]"))
				require.Equal(t, []string{"pos"}, args)
			},
		},
		{
			name: "dynamic int flag followed by positional argument",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().DynamicInt("foo-<name>-size", cli.FlagDoc{})
				return c
			},
			args: []string{"--foo-bar-size", "5", "pos"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, 5, cmd.Flags().GetInt("foo-bar-size"))
				require.Equal(t, []string{"pos"}, args)
			},
		},
		{
			name: "dynamic list flag is greedy",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().DynamicStringSlice("foo[<index>]-list", false, cli.FlagDoc{})
				return c
			},
			args: []string{"--foo[0]-list", "a", "b"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b"}, cmd.Flags().GetStringSlice("foo[0]-list"))
				require.Empty(t, args)
			},
		},
		{
			name: "list flag followed by -- separator ends the list",
			cmd: func() *cli.Command {
				c := &cli.Command{}
				c.Flags().StringSlice("list", nil, false, cli.FlagDoc{})
				return c
			},
			args: []string{"--list", "a", "b", "--", "pos"},
			test: func(t *testing.T, cmd *cli.Command, args []string, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b"}, cmd.Flags().GetStringSlice("list"))
				require.Equal(t, []string{"pos"}, args)
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
