package directory_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/ldap/ldaptest"
	"mokapi/providers/directory"
	"mokapi/sortedmap"
	"testing"
)

func TestDirectory_ServeDelete(t *testing.T) {
	testcases := []struct {
		name   string
		config *directory.Config
		fn     func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry])
	}{
		{
			name: "delete entry",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.DeleteRequest{
					Dn: "cn=foo",
				}))
				res := rr.Message.(*ldap.DeleteResponse)
				require.Equal(t, ldap.Success, res.ResultCode)

				_, ok := entries.Get("cn=foo")
				require.False(t, ok)
			},
		},
		{
			name: "delete but not exists",
			config: &directory.Config{
				Info:    directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.DeleteRequest{
					Dn: "cn=foo",
				}))
				res := rr.Message.(*ldap.DeleteResponse)
				require.Equal(t, ldap.NoSuchObject, res.ResultCode)
			},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			d := directory.NewHandler(test.config, enginetest.NewEngine())
			test.fn(t, d, test.config.Entries)
		})
	}
}
