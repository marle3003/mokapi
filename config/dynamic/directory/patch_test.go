package directory

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig_Patch(t *testing.T) {
	testcases := []struct {
		name    string
		configs []*Config
		test    func(t *testing.T, result *Config)
	}{
		{
			name: "description and version",
			configs: []*Config{
				{},
				{Info: Info{
					Description: "foo",
					Version:     "2.0",
				}},
			},
			test: func(t *testing.T, result *Config) {
				require.Equal(t, "foo", result.Info.Description)
				require.Equal(t, "2.0", result.Info.Version)
			},
		},
		{
			name: "address",
			configs: []*Config{
				{},
				{Address: "foo.bar"},
			},
			test: func(t *testing.T, result *Config) {
				require.Equal(t, "foo.bar", result.Address)
			},
		},
		{
			name: "address not overwrite",
			configs: []*Config{
				{Address: "foo.bar"},
				{Address: "bar.foo"},
			},
			test: func(t *testing.T, result *Config) {
				require.Equal(t, "foo.bar", result.Address)
			},
		},
		{
			name: "root dn",
			configs: []*Config{
				{},
				{Root: Entry{Dn: "foo"}},
			},
			test: func(t *testing.T, result *Config) {
				require.Equal(t, "foo", result.Root.Dn)
			},
		},
		{
			name: "add attribute",
			configs: []*Config{
				{Root: Entry{Attributes: map[string][]string{"foo": {"bar"}}}},
				{Root: Entry{Attributes: map[string][]string{"bar": {"foo"}}}},
			},
			test: func(t *testing.T, result *Config) {
				require.Len(t, result.Root.Attributes, 2)
				require.Equal(t, []string{"bar"}, result.Root.Attributes["foo"])
				require.Equal(t, []string{"foo"}, result.Root.Attributes["bar"])
			},
		},
		{
			name: "size limit",
			configs: []*Config{
				{},
				{SizeLimit: 10},
			},
			test: func(t *testing.T, result *Config) {
				require.Equal(t, int64(10), result.SizeLimit)
			},
		},
		{
			name: "entries",
			configs: []*Config{
				{},
				{Entries: map[string]Entry{
					"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"bar"}}},
				}},
			},
			test: func(t *testing.T, result *Config) {
				require.Len(t, result.Entries, 1)
				require.Equal(t, "foo", result.Entries["foo"].Dn)
				require.Equal(t, []string{"bar"}, result.Entries["foo"].Attributes["foo"])
			},
		},
		{
			name: "merge entries",
			configs: []*Config{
				{Entries: map[string]Entry{
					"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"bar"}}},
				}},
				{Entries: map[string]Entry{
					"bar": {Dn: "bar", Attributes: map[string][]string{"bar": {"foo"}}},
				}},
			},
			test: func(t *testing.T, result *Config) {
				require.Len(t, result.Entries, 2)
				require.Equal(t, "foo", result.Entries["foo"].Dn)
				require.Equal(t, []string{"bar"}, result.Entries["foo"].Attributes["foo"])

				require.Equal(t, "bar", result.Entries["bar"].Dn)
				require.Equal(t, []string{"foo"}, result.Entries["bar"].Attributes["bar"])
			},
		},
		{
			name: "add attribute to existing entry",
			configs: []*Config{
				{Entries: map[string]Entry{
					"foo": {Dn: "foo"},
				}},
				{Entries: map[string]Entry{
					"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"bar"}}},
				}},
			},
			test: func(t *testing.T, result *Config) {
				require.Len(t, result.Entries, 1)
				require.Equal(t, "foo", result.Entries["foo"].Dn)
				require.Equal(t, []string{"bar"}, result.Entries["foo"].Attributes["foo"])
			},
		},
		{
			name: "add attribute to existing entry and attributes",
			configs: []*Config{
				{Entries: map[string]Entry{
					"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"bar"}}},
				}},
				{Entries: map[string]Entry{
					"foo": {Dn: "foo", Attributes: map[string][]string{"bar": {"foo"}}},
				}},
			},
			test: func(t *testing.T, result *Config) {
				require.Len(t, result.Entries, 1)
				require.Equal(t, "foo", result.Entries["foo"].Dn)
				require.Equal(t, []string{"bar"}, result.Entries["foo"].Attributes["foo"])
				require.Equal(t, []string{"foo"}, result.Entries["foo"].Attributes["bar"])
			},
		},
		{
			name: "add attribute value to existing entry",
			configs: []*Config{
				{Entries: map[string]Entry{
					"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"bar"}}},
				}},
				{Entries: map[string]Entry{
					"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"foo"}}},
				}},
			},
			test: func(t *testing.T, result *Config) {
				require.Len(t, result.Entries, 1)
				require.Equal(t, "foo", result.Entries["foo"].Dn)
				require.Equal(t, []string{"bar", "foo"}, result.Entries["foo"].Attributes["foo"])
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := tc.configs[0]
			for _, p := range tc.configs[1:] {
				c.Patch(p)
			}
			tc.test(t, c)
		})
	}
}
