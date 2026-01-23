package cli_test

import (
	"mokapi/pkg/cli"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlagString(t *testing.T) {
	type config struct {
		Foo string
	}

	testcases := []struct {
		name string
		args []string
		test func(t *testing.T)
	}{
		{
			name: "string value",
			test: func(t *testing.T) {
				c := cli.Command{Config: &config{}, Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().String("foo", "", cli.FlagDoc{})
				c.SetArgs([]string{"--foo bar"})

				err := c.Execute()

				require.NoError(t, err)
				require.Equal(t, "bar", c.Flags().GetString("foo"))
				require.Equal(t, &config{Foo: "bar"}, c.Config)
			},
		},
		{
			name: "file",
			test: func(t *testing.T) {
				path := createTempFile(t, "test.yml", "foobar")

				c := cli.Command{Config: &config{}, Run: func(cmd *cli.Command, args []string) error {
					return nil
				}}
				c.Flags().String("foo", "", cli.FlagDoc{})
				c.SetArgs([]string{"--foo file:" + path})

				err := c.Execute()

				require.NoError(t, err)
				require.Equal(t, "file:"+path, c.Flags().GetString("foo"))
				require.Equal(t, &config{Foo: "foobar"}, c.Config)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
