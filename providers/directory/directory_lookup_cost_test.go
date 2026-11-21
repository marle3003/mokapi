package directory_test

import (
	"fmt"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/ldap/ldaptest"
	"mokapi/providers/directory"
	"mokapi/runtime/events/eventstest"
	"testing"

	"github.com/stretchr/testify/require"
)

func newDirectoryWithUsers(n int) ldap.Handler {
	entries := directory.NewEntryStore()
	for i := 0; i < n; i++ {
		dn := fmt.Sprintf("cn=user%d,dc=foo,dc=com", i)
		entries.Set(dn, directory.Entry{
			Dn: dn,
			Attributes: map[string][]string{
				"objectclass":  {"person", "top"},
				"userPassword": {"secret"},
				"mail":         {fmt.Sprintf("user%d@foo.bar", i)},
			},
		})
	}
	return directory.NewHandler(&directory.Config{Info: directory.Info{Name: "foo"}, Entries: entries},
		enginetest.NewEngine(), &eventstest.Handler{})
}

// Looking up a single entry by DN must not copy the other entries of the directory.
func TestDirectory_LookupCostIsIndependentOfDirectorySize(t *testing.T) {
	testcases := []struct {
		name    string
		message ldap.Message
	}{
		{
			name:    "bind",
			message: &ldap.BindRequest{Version: 3, Auth: ldap.Simple, Name: "cn=user5,dc=foo,dc=com", Password: "secret"},
		},
		{
			name:    "compare",
			message: &ldap.CompareRequest{Dn: "cn=user5,dc=foo,dc=com", Attribute: "mail", Value: "user5@foo.bar"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			allocs := func(size int) float64 {
				h := newDirectoryWithUsers(size)
				return testing.AllocsPerRun(20, func() {
					h.ServeLDAP(ldaptest.NewRecorder(), ldaptest.NewRequest(0, tc.message))
				})
			}

			small := allocs(10)
			large := allocs(1000)

			require.LessOrEqual(t, large, small+5,
				"allocations per request grow with directory size: %v for 10 entries, %v for 1000 entries", small, large)
		})
	}
}
