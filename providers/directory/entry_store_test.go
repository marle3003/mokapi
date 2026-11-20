package directory_test

import (
	"mokapi/providers/directory"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func newStore(entries ...directory.Entry) *directory.EntryStore {
	s := directory.NewEntryStore()
	for _, e := range entries {
		s.Set(e.Dn, e)
	}
	return s
}

func lookup(s *directory.EntryStore, dn string) directory.Entry {
	e, _ := s.Get(dn)
	return e
}

func person(dn string, attributes map[string][]string) directory.Entry {
	return directory.Entry{Dn: dn, Attributes: attributes}
}

func dns(entries []directory.Entry) []string {
	var r []string
	for _, e := range entries {
		r = append(r, e.Dn)
	}
	return r
}

func TestEntryStore_Read(t *testing.T) {
	t.Run("nil store is empty", func(t *testing.T) {
		var s *directory.EntryStore

		_, ok := s.Get("cn=foo")

		require.False(t, ok)
		require.Equal(t, 0, s.Len())
		require.Empty(t, s.Entries())
	})

	t.Run("empty store", func(t *testing.T) {
		s := directory.NewEntryStore()

		_, ok := s.Get("cn=foo")

		require.False(t, ok)
		require.Equal(t, 0, s.Len())
		require.Empty(t, s.Entries())
	})

	t.Run("DN is case and space insensitive", func(t *testing.T) {
		s := newStore(person("CN=Foo, DC=bar", nil))

		e, ok := s.Get("cn=foo,dc=bar")

		require.True(t, ok)
		require.Equal(t, "CN=Foo, DC=bar", e.Dn)
		require.Equal(t, 1, s.Len())
	})

	t.Run("Entries keeps the insertion order", func(t *testing.T) {
		s := newStore(person("cn=b", nil), person("cn=a", nil), person("cn=c", nil))

		require.Equal(t, []string{"cn=b", "cn=a", "cn=c"}, dns(s.Entries()))
	})

	t.Run("Get returns a deep copy", func(t *testing.T) {
		s := newStore(person("cn=foo", map[string][]string{"mail": {"a@foo.bar"}}))

		e, _ := s.Get("cn=foo")
		e.Attributes["mail"][0] = "changed"
		e.Attributes["other"] = []string{"x"}
		e.Dn = "changed"

		again, _ := s.Get("cn=foo")
		require.Equal(t, "cn=foo", again.Dn)
		require.Equal(t, []string{"a@foo.bar"}, again.Attributes["mail"])
		require.NotContains(t, again.Attributes, "other")
	})

	t.Run("Entries returns deep copies", func(t *testing.T) {
		s := newStore(person("cn=foo", map[string][]string{"mail": {"a@foo.bar"}}))

		entries := s.Entries()
		entries[0].Attributes["mail"][0] = "changed"
		entries[0].Attributes["other"] = []string{"x"}

		again, _ := s.Get("cn=foo")
		require.Equal(t, []string{"a@foo.bar"}, again.Attributes["mail"])
		require.NotContains(t, again.Attributes, "other")
	})

	t.Run("Set stores its own copy of the attributes", func(t *testing.T) {
		attributes := map[string][]string{"mail": {"a@foo.bar"}}
		s := directory.NewEntryStore()
		s.Set("cn=foo", directory.Entry{Dn: "cn=foo", Attributes: attributes})

		attributes["mail"][0] = "changed"

		e, _ := s.Get("cn=foo")
		require.Equal(t, []string{"a@foo.bar"}, e.Attributes["mail"])
	})
}

func TestEntryStore_Apply(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		s := directory.NewEntryStore()

		err := s.Apply(&directory.AddRecord{Dn: "cn=foo", Attributes: map[string][]string{"mail": {"a@foo.bar"}}}, nil)

		require.NoError(t, err)
		e, ok := s.Get("cn=foo")
		require.True(t, ok)
		require.Equal(t, []string{"a@foo.bar"}, e.Attributes["mail"])
	})

	t.Run("add stores its own copy of the attributes", func(t *testing.T) {
		attributes := map[string][]string{"mail": {"a@foo.bar"}}
		s := directory.NewEntryStore()
		require.NoError(t, s.Apply(&directory.AddRecord{Dn: "cn=foo", Attributes: attributes}, nil))

		attributes["mail"][0] = "changed"

		e, _ := s.Get("cn=foo")
		require.Equal(t, []string{"a@foo.bar"}, e.Attributes["mail"])
	})

	t.Run("modify", func(t *testing.T) {
		s := newStore(person("cn=foo", map[string][]string{"mail": {"a@foo.bar"}}))

		err := s.Apply(&directory.ModifyRecord{Dn: "cn=foo", Actions: []*directory.ModifyAction{
			{Type: "replace", Name: "mail", Attributes: map[string][]string{"mail": {"b@foo.bar"}}},
		}}, nil)

		require.NoError(t, err)
		e, _ := s.Get("cn=foo")
		require.Equal(t, []string{"b@foo.bar"}, e.Attributes["mail"])
	})

	t.Run("delete", func(t *testing.T) {
		s := newStore(person("cn=foo", nil))

		err := s.Apply(&directory.DeleteRecord{Dn: "cn=foo"}, nil)

		require.NoError(t, err)
		require.Equal(t, 0, s.Len())
	})

	t.Run("modifyDN with delete of the old entry", func(t *testing.T) {
		s := newStore(person("cn=foo,dc=bar", map[string][]string{"mail": {"a@foo.bar"}}))

		err := s.Apply(&directory.ModifyDnRecord{Dn: "cn=foo,dc=bar", NewRdn: "cn=baz", DeleteOldDn: true}, nil)

		require.NoError(t, err)
		require.Equal(t, []string{"cn=baz,dc=bar"}, dns(s.Entries()))
	})

	t.Run("modifyDN without delete of the old entry keeps two independent entries", func(t *testing.T) {
		s := newStore(person("cn=foo,dc=bar", map[string][]string{"mail": {"a@foo.bar"}}))
		require.NoError(t, s.Apply(&directory.ModifyDnRecord{Dn: "cn=foo,dc=bar", NewRdn: "cn=baz"}, nil))

		require.NoError(t, s.Apply(&directory.ModifyRecord{Dn: "cn=baz,dc=bar", Actions: []*directory.ModifyAction{
			{Type: "replace", Name: "mail", Attributes: map[string][]string{"mail": {"b@foo.bar"}}},
		}}, nil))

		old, _ := s.Get("cn=foo,dc=bar")
		require.Equal(t, []string{"a@foo.bar"}, old.Attributes["mail"])
	})

	t.Run("failing apply leaves the store unchanged", func(t *testing.T) {
		s := newStore(person("cn=foo", map[string][]string{"mail": {"a@foo.bar"}}))

		require.Error(t, s.Apply(&directory.AddRecord{Dn: "cn=foo"}, nil), "already exists")
		require.Error(t, s.Apply(&directory.DeleteRecord{Dn: "cn=unknown"}, nil))
		require.Error(t, s.Apply(&directory.ModifyRecord{Dn: "cn=unknown"}, nil))

		require.Equal(t, []string{"cn=foo"}, dns(s.Entries()))
		e, _ := s.Get("cn=foo")
		require.Equal(t, []string{"a@foo.bar"}, e.Attributes["mail"])
	})
}

func TestEntryStore_MemberOf(t *testing.T) {
	group := person("cn=group", map[string][]string{"member": {"cn=alice"}})

	t.Run("is set right after the write that adds the group", func(t *testing.T) {
		s := newStore(person("cn=alice", map[string][]string{"mail": {"a@foo.bar"}}))

		require.NoError(t, s.Apply(&directory.AddRecord{Dn: group.Dn, Attributes: group.Attributes}, nil))

		alice, _ := s.Get("cn=alice")
		require.Equal(t, []string{"cn=group"}, alice.Attributes["memberOf"])
		require.Equal(t, []string{"a@foo.bar"}, alice.Attributes["mail"])
	})

	t.Run("is removed with the group", func(t *testing.T) {
		s := newStore(person("cn=alice", nil), group)
		require.NoError(t, s.Apply(&directory.ModifyRecord{Dn: "cn=group", Actions: []*directory.ModifyAction{
			{Type: "add", Name: "description", Attributes: map[string][]string{"description": {"x"}}},
		}}, nil))
		alice, _ := s.Get("cn=alice")
		require.Equal(t, []string{"cn=group"}, alice.Attributes["memberOf"])

		require.NoError(t, s.Apply(&directory.DeleteRecord{Dn: "cn=group"}, nil))

		alice, _ = s.Get("cn=alice")
		require.NotContains(t, alice.Attributes, "memberOf")
	})

	t.Run("a copy taken before the write is not changed", func(t *testing.T) {
		s := newStore(person("cn=alice", nil))
		before, _ := s.Get("cn=alice")

		require.NoError(t, s.Apply(&directory.AddRecord{Dn: group.Dn, Attributes: group.Attributes}, nil))

		require.NotContains(t, before.Attributes, "memberOf")
	})

	t.Run("failing apply does not touch memberOf", func(t *testing.T) {
		s := newStore(person("cn=alice", nil))

		require.Error(t, s.Apply(&directory.DeleteRecord{Dn: "cn=unknown"}, nil))

		alice, _ := s.Get("cn=alice")
		require.NotContains(t, alice.Attributes, "memberOf")
	})
}

func TestEntryStore_Clone(t *testing.T) {
	t.Run("is independent of the original", func(t *testing.T) {
		s := newStore(person("cn=foo", map[string][]string{"mail": {"a@foo.bar"}}))

		c := s.Clone()
		require.NoError(t, c.Apply(&directory.ModifyRecord{Dn: "cn=foo", Actions: []*directory.ModifyAction{
			{Type: "add", Name: "mail", Attributes: map[string][]string{"mail": {"b@foo.bar"}}},
		}}, nil))
		require.NoError(t, c.Apply(&directory.AddRecord{Dn: "cn=bar"}, nil))

		require.Equal(t, 1, s.Len())
		e, _ := s.Get("cn=foo")
		require.Equal(t, []string{"a@foo.bar"}, e.Attributes["mail"])
		require.Equal(t, 2, c.Len())
	})

	t.Run("keeps order and DN normalization", func(t *testing.T) {
		s := newStore(person("CN=b", nil), person("CN=a", nil))

		c := s.Clone()

		require.Equal(t, []string{"CN=b", "CN=a"}, dns(c.Entries()))
		_, ok := c.Get("cn=a")
		require.True(t, ok)
	})

	t.Run("nil store gives an empty usable store", func(t *testing.T) {
		var s *directory.EntryStore

		c := s.Clone()

		require.NotNil(t, c)
		require.NoError(t, c.Apply(&directory.AddRecord{Dn: "cn=foo"}, nil))
	})
}

func TestEntryStore_Merge(t *testing.T) {
	t.Run("adds entries and attribute values of the patch", func(t *testing.T) {
		s := newStore(person("cn=foo", map[string][]string{"mail": {"a@foo.bar"}}))
		patch := newStore(
			person("cn=foo", map[string][]string{"mail": {"b@foo.bar"}, "phone": {"1"}}),
			person("cn=bar", map[string][]string{"mail": {"c@foo.bar"}}),
		)

		s.Merge(patch)

		require.Equal(t, []string{"cn=foo", "cn=bar"}, dns(s.Entries()))
		foo, _ := s.Get("cn=foo")
		require.Equal(t, []string{"a@foo.bar", "b@foo.bar"}, foo.Attributes["mail"])
		require.Equal(t, []string{"1"}, foo.Attributes["phone"])
	})

	t.Run("does not share values with the patch", func(t *testing.T) {
		s := newStore(person("cn=foo", map[string][]string{"mail": {"a@foo.bar"}}))
		patch := newStore(
			person("cn=foo", map[string][]string{"phone": {"1"}}),
			person("cn=bar", map[string][]string{"mail": {"c@foo.bar"}}),
		)

		s.Merge(patch)
		require.NoError(t, s.Apply(&directory.ModifyRecord{Dn: "cn=bar", Actions: []*directory.ModifyAction{
			{Type: "add", Name: "mail", Attributes: map[string][]string{"mail": {"d@foo.bar"}}},
		}}, nil))
		require.NoError(t, s.Apply(&directory.ModifyRecord{Dn: "cn=foo", Actions: []*directory.ModifyAction{
			{Type: "add", Name: "phone", Attributes: map[string][]string{"phone": {"2"}}},
		}}, nil))

		bar, _ := patch.Get("cn=bar")
		require.Equal(t, []string{"c@foo.bar"}, bar.Attributes["mail"])
		foo, _ := patch.Get("cn=foo")
		require.Equal(t, []string{"1"}, foo.Attributes["phone"])
	})

	t.Run("nil patch changes nothing", func(t *testing.T) {
		s := newStore(person("cn=foo", nil))

		s.Merge(nil)

		require.Equal(t, 1, s.Len())
	})
}

// Writes of LDAP clients run while other clients search. Run with -race.
func TestEntryStore_ConcurrentApplyAndRead(t *testing.T) {
	s := newStore(person("cn=alice", map[string][]string{"mail": {"a@foo.bar"}}))
	group := &directory.AddRecord{Dn: "cn=group", Attributes: map[string][]string{"member": {"cn=alice"}}}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = s.Apply(group, nil)
			_ = s.Apply(&directory.DeleteRecord{Dn: "cn=group"}, nil)
		}
	}()
	for r := 0; r < 3; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				for _, e := range s.Entries() {
					_ = e.Attributes["memberOf"]
				}
				_, _ = s.Get("cn=alice")
				_ = s.Len()
			}
		}()
	}
	wg.Wait()
}
