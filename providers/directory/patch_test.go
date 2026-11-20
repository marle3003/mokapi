package directory_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/providers/directory"
	"testing"
)

func TestConfig_Patch(t *testing.T) {

	testcases := []struct {
		name    string
		configs []*directory.Config
		test    func(t *testing.T, result *directory.Config)
	}{
		{
			name: "description and version",
			configs: []*directory.Config{
				{},
				{Info: directory.Info{
					Description: "foo",
					Version:     "2.0",
				}},
			},
			test: func(t *testing.T, result *directory.Config) {
				require.Equal(t, "foo", result.Info.Description)
				require.Equal(t, "2.0", result.Info.Version)
			},
		},
		{
			name: "address",
			configs: []*directory.Config{
				{},
				{Address: "foo.bar"},
			},
			test: func(t *testing.T, result *directory.Config) {
				require.Equal(t, "foo.bar", result.Address)
			},
		},
		{
			name: "address is overwrite",
			configs: []*directory.Config{
				{Address: "foo.bar"},
				{Address: "bar.foo"},
			},
			test: func(t *testing.T, result *directory.Config) {
				require.Equal(t, "bar.foo", result.Address)
			},
		},
		{
			name: "size limit",
			configs: []*directory.Config{
				{},
				{SizeLimit: 10},
			},
			test: func(t *testing.T, result *directory.Config) {
				require.Equal(t, int64(10), result.SizeLimit)
			},
		},
		{
			name: "entries",
			configs: []*directory.Config{
				{},
				{
					Entries: convert(map[string]directory.Entry{
						"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"bar"}}},
					}),
				},
			},
			test: func(t *testing.T, result *directory.Config) {
				require.Equal(t, 1, result.Entries.Len())
				require.Equal(t, "foo", lookup(result.Entries, "foo").Dn)
				require.Equal(t, []string{"bar"}, lookup(result.Entries, "foo").Attributes["foo"])
			},
		},
		{
			name: "merge entries",
			configs: []*directory.Config{
				{
					Entries: convert(map[string]directory.Entry{
						"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"bar"}}},
					}),
				},
				{
					Entries: convert(map[string]directory.Entry{
						"bar": {Dn: "bar", Attributes: map[string][]string{"bar": {"foo"}}},
					}),
				},
			},
			test: func(t *testing.T, result *directory.Config) {
				require.Equal(t, 2, result.Entries.Len())
				require.Equal(t, "foo", lookup(result.Entries, "foo").Dn)
				require.Equal(t, []string{"bar"}, lookup(result.Entries, "foo").Attributes["foo"])

				require.Equal(t, "bar", lookup(result.Entries, "bar").Dn)
				require.Equal(t, []string{"foo"}, lookup(result.Entries, "bar").Attributes["bar"])
			},
		},
		{
			name: "add attribute to existing entry",
			configs: []*directory.Config{
				{
					Entries: convert(map[string]directory.Entry{
						"foo": {Dn: "foo"},
					}),
				},
				{
					Entries: convert(map[string]directory.Entry{
						"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"bar"}}},
					}),
				},
			},
			test: func(t *testing.T, result *directory.Config) {
				require.Equal(t, 1, result.Entries.Len())
				require.Equal(t, "foo", lookup(result.Entries, "foo").Dn)
				require.Equal(t, []string{"bar"}, lookup(result.Entries, "foo").Attributes["foo"])
			},
		},
		{
			name: "add attribute to existing entry and attributes",
			configs: []*directory.Config{
				{
					Entries: convert(map[string]directory.Entry{
						"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"bar"}}},
					}),
				},
				{
					Entries: convert(map[string]directory.Entry{
						"foo": {Dn: "foo", Attributes: map[string][]string{"bar": {"foo"}}},
					}),
				},
			},
			test: func(t *testing.T, result *directory.Config) {
				require.Equal(t, 1, result.Entries.Len())
				require.Equal(t, "foo", lookup(result.Entries, "foo").Dn)
				require.Equal(t, []string{"bar"}, lookup(result.Entries, "foo").Attributes["foo"])
				require.Equal(t, []string{"foo"}, lookup(result.Entries, "foo").Attributes["bar"])
			},
		},
		{
			name: "add attribute value to existing entry",
			configs: []*directory.Config{
				{
					Entries: convert(map[string]directory.Entry{
						"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"bar"}}},
					}),
				},
				{
					Entries: convert(map[string]directory.Entry{
						"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"foo"}}},
					}),
				},
			},
			test: func(t *testing.T, result *directory.Config) {
				require.Equal(t, 1, result.Entries.Len())
				require.Equal(t, "foo", lookup(result.Entries, "foo").Dn)
				require.Equal(t, []string{"bar", "foo"}, lookup(result.Entries, "foo").Attributes["foo"])
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

func convert(m map[string]directory.Entry) *directory.EntryStore {
	r := directory.NewEntryStore()
	for k, v := range m {
		r.Set(k, v)
	}
	return r
}

func TestConfig_Clone(t *testing.T) {
	t.Run("entries are independent of the original", func(t *testing.T) {
		original := &directory.Config{
			Address: "foo.bar",
			Entries: convert(map[string]directory.Entry{
				"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"bar"}}},
			}),
		}

		clone := original.Clone()
		require.NoError(t, clone.Entries.Apply(&directory.AddRecord{Dn: "bar"}, nil))
		require.NoError(t, clone.Entries.Apply(&directory.ModifyRecord{Dn: "foo", Actions: []*directory.ModifyAction{
			{Type: "replace", Name: "foo", Attributes: map[string][]string{"foo": {"changed"}}},
		}}, nil))

		require.Equal(t, "foo.bar", clone.Address)
		require.Equal(t, 1, original.Entries.Len())
		require.Equal(t, []string{"bar"}, lookup(original.Entries, "foo").Attributes["foo"])
	})

	t.Run("config without entries gets an empty store", func(t *testing.T) {
		clone := (&directory.Config{}).Clone()

		require.NotNil(t, clone.Entries)
		require.NoError(t, clone.Entries.Apply(&directory.AddRecord{Dn: "foo"}, nil))
	})

	t.Run("patch does not change the original", func(t *testing.T) {
		original := &directory.Config{
			Info: directory.Info{Description: "a"},
			Entries: convert(map[string]directory.Entry{
				"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"bar"}}},
			}),
		}

		clone := original.Clone()
		clone.Patch(&directory.Config{
			Info: directory.Info{Description: "b"},
			Entries: convert(map[string]directory.Entry{
				"foo": {Dn: "foo", Attributes: map[string][]string{"foo": {"baz"}}},
			}),
		})

		require.Equal(t, "a", original.Info.Description)
		require.Equal(t, []string{"bar"}, lookup(original.Entries, "foo").Attributes["foo"])
		require.Equal(t, []string{"bar", "baz"}, lookup(clone.Entries, "foo").Attributes["foo"])
	})
}
