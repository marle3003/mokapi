package cli_test

import (
	"mokapi/pkg/cli"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlagSet_Group(t *testing.T) {
	t.Run("flags get the group set before their registration", func(t *testing.T) {
		fs := (&cli.Command{}).Flags()
		fs.String("before", "", cli.FlagDoc{})
		fs.Group("Log")
		fs.String("level", "", cli.FlagDoc{})
		fs.Int("size", 0, cli.FlagDoc{})
		fs.Group("API")
		fs.Bool("dashboard", true, cli.FlagDoc{})

		require.Equal(t, "", fs.Lookup("before").Group)
		require.Equal(t, "Log", fs.Lookup("level").Group)
		require.Equal(t, "Log", fs.Lookup("size").Group)
		require.Equal(t, "API", fs.Lookup("dashboard").Group)
	})

	t.Run("order follows registration", func(t *testing.T) {
		fs := (&cli.Command{}).Flags()
		fs.String("b", "", cli.FlagDoc{})
		fs.String("a", "", cli.FlagDoc{})
		fs.String("c", "", cli.FlagDoc{})

		require.Less(t, fs.Lookup("b").Order, fs.Lookup("a").Order)
		require.Less(t, fs.Lookup("a").Order, fs.Lookup("c").Order)
	})

	t.Run("order does not change when a flag is set", func(t *testing.T) {
		c := &cli.Command{Run: func(cmd *cli.Command, args []string) error { return nil }}
		c.Flags().String("b", "", cli.FlagDoc{})
		c.Flags().String("a", "", cli.FlagDoc{})
		order := c.Flags().Lookup("b").Order
		c.SetArgs([]string{"--a", "x", "--b", "y"})

		require.NoError(t, c.Execute())

		require.Equal(t, order, c.Flags().Lookup("b").Order)
	})

	t.Run("dynamic flag keeps group and order", func(t *testing.T) {
		c := &cli.Command{Run: func(cmd *cli.Command, args []string) error { return nil }}
		c.Flags().Group("Events")
		c.Flags().DynamicInt("store-<name>-size", cli.FlagDoc{})
		c.Flags().Group("Log")
		c.Flags().String("level", "", cli.FlagDoc{})
		c.SetArgs([]string{"--store-foo-size", "5"})

		require.NoError(t, c.Execute())

		f := c.Flags().Lookup("store-foo-size")
		require.NotNil(t, f)
		require.Equal(t, "Events", f.Group)
		require.Less(t, f.Order, c.Flags().Lookup("level").Order)
	})

	t.Run("dynamic flag in docs listing has group and order", func(t *testing.T) {
		fs := (&cli.Command{}).Flags()
		fs.Group("Events")
		fs.DynamicString("store-<name>", cli.FlagDoc{})

		var found *cli.Flag
		_ = fs.VisitAll(func(f *cli.Flag) error {
			if f.Name == "store-<name>" {
				found = f
			}
			return nil
		})

		require.NotNil(t, found)
		require.Equal(t, "Events", found.Group)
		require.NotZero(t, found.Order)
	})
}
