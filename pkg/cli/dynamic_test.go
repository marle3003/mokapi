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
				c.Flags().DynamicString("foo[<index>]", "")
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
				c.Flags().DynamicString("foo[<index>]", "")
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
				c.Flags().DynamicString("foo-<key>", "")
				err := c.Execute()
				require.NoError(t, err)
				require.Equal(t, map[string]string{"bar": "yuh"}, s.Foo)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
