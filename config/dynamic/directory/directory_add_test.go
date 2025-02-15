package directory_test

import (
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/directory"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/ldap/ldaptest"
	"mokapi/sortedmap"
	"testing"
)

func TestDirectory_ServeAdd(t *testing.T) {
	testcases := []struct {
		name   string
		config *directory.Config
		fn     func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry])
	}{
		{
			name: "add entry",
			config: &directory.Config{
				Info:    directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.AddRequest{
					Dn: "cn=foo",
				}))
				res := rr.Message.(*ldap.AddResponse)
				require.Equal(t, ldap.Success, res.ResultCode)

				_, ok := entries.Get("cn=foo")
				require.True(t, ok)
			},
		},
		{
			name: "add entry with attributes",
			config: &directory.Config{
				Info:    directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.AddRequest{
					Dn: "cn=foo",
					Attributes: []ldap.Attribute{
						{
							Type:   "foo",
							Values: []string{"bar"},
						},
					},
				}))
				res := rr.Message.(*ldap.AddResponse)
				require.Equal(t, ldap.Success, res.ResultCode)

				e, ok := entries.Get("cn=foo")
				require.True(t, ok)
				require.Contains(t, e.Attributes, "foo")
				require.Equal(t, []string{"bar"}, e.Attributes["foo"])
			},
		},
		{
			name: "add entry but already exists",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.AddRequest{
					Dn: "cn=foo",
				}))
				res := rr.Message.(*ldap.AddResponse)
				require.Equal(t, ldap.EntryAlreadyExists, res.ResultCode)
			},
		},
		{
			name: "invalid attribute value",
			config: &directory.Config{
				Info:    directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{}),
				Schema: &directory.Schema{AttributeTypes: map[string]*directory.AttributeType{
					"foo": {
						Syntax: "1.3.6.1.4.1.1466.115.121.1.7",
					},
				}},
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.AddRequest{
					Dn: "cn=foo",
					Attributes: []ldap.Attribute{
						{
							Type:   "foo",
							Values: []string{"bar"},
						},
					},
				}))
				res := rr.Message.(*ldap.AddResponse)
				require.Equal(t, ldap.ConstraintViolation, res.ResultCode)
				require.Equal(t, "cn=foo", res.MatchedDn)
				require.Equal(t, "invalid value for attribute foo=bar: SYNTAX: 1.3.6.1.4.1.1466.115.121.1.7", res.Message)
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
