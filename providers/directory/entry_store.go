package directory

import (
	"mokapi/sortedmap"
	"sync"

	log "github.com/sirupsen/logrus"
)

// EntryStore holds the entries of a directory. LDAP clients write to it while other clients
// search, so the entries are only reachable under the store's lock: reads return deep copies and
// Apply is the only way for clients to write. The key of an entry is its normalized DN.
// Readers can call Get, Len and Entries on a nil store.
type EntryStore struct {
	m       sync.RWMutex
	entries *sortedmap.LinkedHashMap[string, Entry]
}

func NewEntryStore() *EntryStore {
	return &EntryStore{entries: newEntryMap()}
}

func newEntryMap() *sortedmap.LinkedHashMap[string, Entry] {
	return &sortedmap.LinkedHashMap[string, Entry]{KeyNormalizer: normalizeDN}
}

func (s *EntryStore) Get(dn string) (Entry, bool) {
	if s == nil {
		return Entry{}, false
	}
	s.m.RLock()
	defer s.m.RUnlock()

	e, ok := s.entries.Get(dn)
	if !ok {
		return Entry{}, false
	}
	return e.copy(), true
}

// Set stores a copy of e.
func (s *EntryStore) Set(dn string, e Entry) {
	s.m.Lock()
	defer s.m.Unlock()

	s.entries.Set(dn, e.copy())
}

func (s *EntryStore) Len() int {
	if s == nil {
		return 0
	}
	s.m.RLock()
	defer s.m.RUnlock()

	return s.entries.Len()
}

// Entries returns copies of all entries in insertion order.
func (s *EntryStore) Entries() []Entry {
	if s == nil {
		return nil
	}
	s.m.RLock()
	defer s.m.RUnlock()

	r := make([]Entry, 0, s.entries.Len())
	for it := s.entries.Iter(); it.Next(); {
		e := it.Value()
		r = append(r, e.copy())
	}
	return r
}

// Apply applies a write operation. The store keeps memberOf in sync with the member attributes,
// so a client sees the new memberOf values right after its write.
func (s *EntryStore) Apply(rec LdifRecord, schema *Schema) error {
	s.m.Lock()
	defer s.m.Unlock()

	if err := rec.Apply(s.entries, schema); err != nil {
		return err
	}
	rebuildMemberOf(s.entries)
	return nil
}

// Clone returns an independent store with deep copies of the entries. A nil store gives an empty one.
func (s *EntryStore) Clone() *EntryStore {
	if s == nil {
		return NewEntryStore()
	}
	s.m.RLock()
	defer s.m.RUnlock()

	entries := s.entries.CloneWith(func(e Entry) Entry { return e.copy() })
	return &EntryStore{entries: &entries}
}

// Merge adds the entries of patch: attribute values of an entry that already exists are appended.
// The store does not share any value with patch.
func (s *EntryStore) Merge(patch *EntryStore) {
	if patch == nil {
		return
	}
	// copy first, so no lock of patch is held while the lock of s is taken
	copied := patch.Clone().entries

	s.m.Lock()
	defer s.m.Unlock()

	for it := copied.Iter(); it.Next(); {
		if e, ok := s.entries.Get(it.Key()); ok {
			e.patch(it.Value())
			s.entries.Set(it.Key(), e)
		} else {
			s.entries.Set(it.Key(), it.Value())
		}
	}
}

// rebuildMemberOf sets the memberOf attribute of every entry from the member attributes of the
// groups. It replaces the attributes of changed entries instead of changing them in place.
func rebuildMemberOf(entries *sortedmap.LinkedHashMap[string, Entry]) {
	type membership struct{ group, member string }
	var memberships []membership
	var withMemberOf []string
	for it := entries.Iter(); it.Next(); {
		e := it.Value()
		if _, ok := e.Attributes["memberOf"]; ok {
			withMemberOf = append(withMemberOf, it.Key())
		}
		for _, member := range e.Attributes["member"] {
			memberships = append(memberships, membership{group: e.Dn, member: member})
		}
	}

	for _, key := range withMemberOf {
		e, _ := entries.Get(key)
		e = e.copy()
		delete(e.Attributes, "memberOf")
		entries.Set(key, e)
	}

	for _, m := range memberships {
		e, ok := entries.Get(m.member)
		if !ok {
			log.Warnf("LDAP: member %s for group %s not found", m.member, m.group)
			continue
		}
		e = e.copy()
		e.Attributes["memberOf"] = append(e.Attributes["memberOf"], m.group)
		entries.Set(m.member, e)
	}
}
