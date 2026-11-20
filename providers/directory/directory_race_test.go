package directory_test

import (
	"io"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/ldap/ldaptest"
	"mokapi/providers/directory"
	"mokapi/runtime/events/eventstest"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

// LDAP clients write entries while other clients search. Run with -race.
func TestDirectory_WritesDoNotRaceWithSearches(t *testing.T) {
	logrus.SetOutput(io.Discard)

	cfg := &directory.Config{
		Info: directory.Info{Name: "foo"},
		Entries: convert(map[string]directory.Entry{
			"": {Attributes: map[string][]string{"subschemaSubentry": {"cn=subschema"}}},
			"cn=group": {
				Dn: "cn=group",
				Attributes: map[string][]string{
					"objectClass": {"group"},
					"member":      {"cn=user1", "cn=tmp"},
				},
			},
			"cn=user1": {
				Dn:         "cn=user1",
				Attributes: map[string][]string{"objectClass": {"person"}},
			},
		}),
	}
	h := directory.NewHandler(cfg, enginetest.NewEngine(), &eventstest.Handler{})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		write := func(msg ldap.Message) {
			h.ServeLDAP(ldaptest.NewRecorder(), ldaptest.NewRequest(0, msg))
		}
		for i := 0; i < 100; i++ {
			write(&ldap.AddRequest{
				Dn:         "cn=tmp",
				Attributes: []ldap.Attribute{{Type: "objectClass", Values: []string{"person"}}},
			})
			write(&ldap.ModifyRequest{
				Dn: "cn=tmp",
				Items: []ldap.ModificationItem{{
					Operation:    ldap.ReplaceOperation,
					Modification: ldap.Modification{Type: "mail", Values: []string{"tmp@foo.bar"}},
				}},
			})
			write(&ldap.ModifyDNRequest{Dn: "cn=tmp", NewRdn: "cn=tmp2", DeleteOldDn: false})
			write(&ldap.DeleteRequest{Dn: "cn=tmp2"})
			write(&ldap.DeleteRequest{Dn: "cn=tmp"})
		}
	}()

	for r := 0; r < 3; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(objectClass=*)",
				}))
				if _, ok := rr.Message.(*ldap.SearchResponse); !ok {
					t.Error("expected a search response")
					return
				}
				h.ServeLDAP(ldaptest.NewRecorder(), ldaptest.NewRequest(0, &ldap.CompareRequest{Dn: "cn=user1", Attribute: "memberOf", Value: "cn=group"}))
			}
		}()
	}
	wg.Wait()
}
