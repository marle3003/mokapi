package cli_test

import (
	"mokapi/pkg/cli"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBind_Slice(t *testing.T) {
	type config struct {
		Foo []string
	}

	testcases := []struct {
		name   string
		config *config
		args   []string
		test   func(t *testing.T, cfg *config, err error)
	}{
		{
			name:   "replaces a longer existing list",
			config: &config{Foo: []string{"x", "y", "z"}},
			args:   []string{"--foo", "a"},
			test: func(t *testing.T, cfg *config, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a"}, cfg.Foo)
			},
		},
		{
			name:   "replaces a shorter existing list",
			config: &config{Foo: []string{"x"}},
			args:   []string{"--foo", "a,b,c"},
			test: func(t *testing.T, cfg *config, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"a", "b", "c"}, cfg.Foo)
			},
		},
		{
			name:   "index path sets only the item",
			config: &config{Foo: []string{"x", "y", "z"}},
			args:   []string{"--foo[1]", "a"},
			test: func(t *testing.T, cfg *config, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"x", "a", "z"}, cfg.Foo)
			},
		},
		{
			name:   "index path extends the list",
			config: &config{},
			args:   []string{"--foo[2]", "a"},
			test: func(t *testing.T, cfg *config, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"", "", "a"}, cfg.Foo)
			},
		},
		{
			name:   "empty value keeps existing list",
			config: &config{Foo: []string{"x", "y"}},
			args:   []string{"--foo", ""},
			test: func(t *testing.T, cfg *config, err error) {
				require.NoError(t, err)
				require.Equal(t, []string{"x", "y"}, cfg.Foo)
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cli.Command{Name: "foo", Config: tc.config}
			cmd.Flags().StringSlice("foo", nil, false, cli.FlagDoc{})
			cmd.Flags().String("foo[1]", "", cli.FlagDoc{})
			cmd.Flags().String("foo[2]", "", cli.FlagDoc{})
			cmd.SetArgs(tc.args)
			cmd.Run = func(c *cli.Command, a []string) error { return nil }

			err := cmd.Execute()

			tc.test(t, tc.config, err)
		})
	}
}

func TestBind_SliceOfStructs(t *testing.T) {
	type item struct {
		Name string
		Note string
	}
	type config struct {
		Items []item
	}

	testcases := []struct {
		name   string
		config *config
		args   []string
		expect []item
	}{
		{
			name:   "list replaces items without inheriting their fields",
			config: &config{Items: []item{{Name: "x", Note: "old"}}},
			args:   []string{"--items", "name=a"},
			expect: []item{{Name: "a"}},
		},
		{
			name:   "index flag refines item of new list, set before the list flag",
			config: &config{},
			args:   []string{"--items-0-note", "n", "--items", "name=a"},
			expect: []item{{Name: "a", Note: "n"}},
		},
		{
			name:   "index flag refines item of new list, set after the list flag",
			config: &config{},
			args:   []string{"--items", "name=a", "--items-0-note", "n"},
			expect: []item{{Name: "a", Note: "n"}},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cli.Command{Name: "foo", Config: tc.config}
			cmd.Flags().StringSlice("items", nil, false, cli.FlagDoc{})
			cmd.Flags().String("items-0-note", "", cli.FlagDoc{})
			cmd.SetArgs(tc.args)
			cmd.Run = func(c *cli.Command, a []string) error { return nil }

			require.NoError(t, cmd.Execute())

			require.Equal(t, tc.expect, tc.config.Items)
		})
	}
}
