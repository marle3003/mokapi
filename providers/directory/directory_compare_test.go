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

func TestDirectory_ServeCompare(t *testing.T) {
	testcases := []struct {
		name   string
		config *directory.Config
		fn     func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry])
	}{
		{
			name: "attribute and value exists",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {Dn: "cn=foo", Attributes: map[string][]string{"foo": {"bar"}}},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.CompareRequest{
					Dn:        "cn=foo",
					Attribute: "foo",
					Value:     "bar",
				}))
				res := rr.Message.(*ldap.CompareResponse)
				require.Equal(t, ldap.CompareTrue, res.ResultCode)
			},
		},
		{
			name: "attribute exists but value does not match",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {Dn: "cn=foo", Attributes: map[string][]string{"foo": {"bar"}}},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.CompareRequest{
					Dn:        "cn=foo",
					Attribute: "foo",
					Value:     "yuh",
				}))
				res := rr.Message.(*ldap.CompareResponse)
				require.Equal(t, ldap.CompareFalse, res.ResultCode)
			},
		},
		{
			name: "attribute does not exist",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {Dn: "cn=foo", Attributes: map[string][]string{"foo": {"bar"}}},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.CompareRequest{
					Dn:        "cn=foo",
					Attribute: "bar",
					Value:     "yuh",
				}))
				res := rr.Message.(*ldap.CompareResponse)
				require.Equal(t, ldap.CompareFalse, res.ResultCode)
			},
		},
		{
			name: "entry does not exist",
			config: &directory.Config{
				Info:    directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.CompareRequest{
					Dn:        "cn=foo",
					Attribute: "bar",
					Value:     "yuh",
				}))
				res := rr.Message.(*ldap.CompareResponse)
				require.Equal(t, ldap.CompareFalse, res.ResultCode)
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
