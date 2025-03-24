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

func TestDirectory_ServeModifyDn(t *testing.T) {
	testcases := []struct {
		name   string
		config *directory.Config
		fn     func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry])
	}{
		{
			name: "replace entry",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {Dn: "cn=foo"},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyDNRequest{
					Dn:          "cn=foo",
					NewRdn:      "cn=bar",
					DeleteOldDn: true,
				}))
				res := rr.Message.(*ldap.ModifyDNResponse)
				require.Equal(t, ldap.Success, res.ResultCode)

				e, ok := entries.Get("cn=bar")
				require.True(t, ok)
				require.Equal(t, e.Dn, "cn=bar")

				_, ok = entries.Get("cn=foo")
				require.False(t, ok)
			},
		},
		{
			name: "copy entry",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {Dn: "cn=foo"},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyDNRequest{
					Dn:          "cn=foo",
					NewRdn:      "cn=bar",
					DeleteOldDn: false,
				}))
				res := rr.Message.(*ldap.ModifyDNResponse)
				require.Equal(t, ldap.Success, res.ResultCode)

				e, ok := entries.Get("cn=foo")
				require.True(t, ok)
				require.Equal(t, e.Dn, "cn=foo")

				e, ok = entries.Get("cn=bar")
				require.True(t, ok)
				require.Equal(t, e.Dn, "cn=bar")
			},
		},
		{
			name: "move entry from root",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {Dn: "cn=foo"},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyDNRequest{
					Dn:            "cn=foo",
					NewSuperiorDn: "ou=users",
					DeleteOldDn:   true,
				}))
				res := rr.Message.(*ldap.ModifyDNResponse)
				require.Equal(t, ldap.Success, res.ResultCode)

				e, ok := entries.Get("cn=foo")
				require.False(t, ok, "cn=foo")

				e, ok = entries.Get("cn=foo,ou=users")
				require.True(t, ok, "cn=foo,ou=users")
				require.Equal(t, e.Dn, "cn=foo,ou=users")
			},
		},
		{
			name: "move entry",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=alice,ou=foo": {Dn: "cn=alice,ou=foo"},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyDNRequest{
					Dn:            "cn=alice,ou=foo",
					NewSuperiorDn: "ou=foo,ou=bar",
					DeleteOldDn:   true,
				}))
				res := rr.Message.(*ldap.ModifyDNResponse)
				require.Equal(t, ldap.Success, res.ResultCode)

				e, ok := entries.Get("cn=alice,ou=foo")
				require.False(t, ok, "cn=alice,ou=foo")

				e, ok = entries.Get("cn=alice,ou=foo,ou=bar")
				require.True(t, ok, "cn=alice,ou=foo,ou=bar")
				require.Equal(t, e.Dn, "cn=alice,ou=foo,ou=bar")
			},
		},
		{
			name: "entry not exists",
			config: &directory.Config{
				Info:    directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyDNRequest{
					Dn:          "cn=foo",
					NewRdn:      "cn=bar",
					DeleteOldDn: false,
				}))
				res := rr.Message.(*ldap.ModifyDNResponse)
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
