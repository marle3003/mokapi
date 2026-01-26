package cli_test

import (
	"mokapi/pkg/cli"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDynamic(t *testing.T) {
	newCmd := func(args []string, cfg any) *cli.Command {
		c := &cli.Command{EnvPrefix: "Mokapi_"}
		c.SetArgs(args)
		c.Config = cfg
		c.Run = func(cmd *cli.Command, args []string) error {
			return nil
		}
		return c
	}

	testcases := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "array index 0",
			test: func(t *testing.T) {
				s := &struct{ Foo []string }{}
				c := newCmd([]string{"--foo[0]", "bar"}, &s)
				c.Flags().DynamicString("foo[<index>]", cli.FlagDoc{})
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, []string{"bar"}, s.Foo)
			},
		},
		{
			name: "array index 1",
			test: func(t *testing.T) {
				s := &struct{ Foo []string }{}
				c := newCmd([]string{"--foo[1]", "bar"}, &s)
				c.Flags().DynamicString("foo[<index>]", cli.FlagDoc{})
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, []string{"", "bar"}, s.Foo)
			},
		},
		{
			name: "map",
			test: func(t *testing.T) {
				s := &struct{ Foo map[string]string }{}
				c := newCmd([]string{"--foo-bar", "yuh"}, &s)
				c.Flags().DynamicString("foo-<key>", cli.FlagDoc{})
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, map[string]string{"bar": "yuh"}, s.Foo)
			},
		},
		{
			name: "dynamic with an array",
			test: func(t *testing.T) {
				s := &struct {
					Foo map[string]map[string][]string
				}{}
				c := newCmd([]string{"--foo[0]-bar", "a b c"}, &s)
				c.Flags().DynamicStringSlice("foo[<key>]-bar", false, cli.FlagDoc{})
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, map[string]map[string][]string{
					"[0]": {"bar": []string{"a", "b", "c"}},
				}, s.Foo)
			},
		},
		{
			name: "dynamic with an array and explode",
			test: func(t *testing.T) {
				s := &struct {
					Foo map[string]map[string][]string
				}{}
				c := newCmd([]string{"--foo[0]-bar", "a", "--foo[0]-bar", "b"}, &s)
				c.Flags().DynamicStringSlice("foo[<key>]-bar", true, cli.FlagDoc{})
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, map[string]map[string][]string{
					"[0]": {"bar": []string{"a", "b"}},
				}, s.Foo)
			},
		},
		{
			name: "int",
			test: func(t *testing.T) {
				s := &struct {
					Foo []int
				}{}
				c := newCmd([]string{"--foo[0]", "12"}, &s)
				c.Flags().DynamicInt("foo[<index>]", cli.FlagDoc{})
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, []int{12}, s.Foo)
			},
		},
		{
			name: "old style",
			test: func(t *testing.T) {
				s := &struct {
					Foo []int
				}{}
				c := newCmd([]string{"--foo-0", "12"}, &s)
				c.Flags().DynamicInt("foo[<index>]", cli.FlagDoc{})
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, []int{12}, s.Foo)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
