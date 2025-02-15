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

func TestDirectory_ServeModify(t *testing.T) {
	testcases := []struct {
		name   string
		config *directory.Config
		fn     func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry])
	}{
		{
			name: "add attribute",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {
						Dn: "cn=foo",
					},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyRequest{
					Dn: "cn=foo",
					Items: []ldap.ModificationItem{
						{
							Operation:    ldap.AddOperation,
							Modification: ldap.Modification{Type: "foo", Values: []string{"bar"}},
						},
					},
				}))
				res := rr.Message.(*ldap.ModifyResponse)
				require.Equal(t, ldap.Success, res.ResultCode)

				e, ok := entries.Get("cn=foo")
				require.True(t, ok)
				require.Contains(t, e.Attributes, "foo")
				require.Equal(t, "bar", e.Attributes["foo"][0])
			},
		},
		{
			name: "delete attribute",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {
						Dn:         "cn=foo",
						Attributes: map[string][]string{"foo": {"bar"}},
					},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyRequest{
					Dn: "cn=foo",
					Items: []ldap.ModificationItem{
						{
							Operation:    ldap.DeleteOperation,
							Modification: ldap.Modification{Type: "foo"},
						},
					},
				}))
				res := rr.Message.(*ldap.ModifyResponse)
				require.Equal(t, ldap.Success, res.ResultCode)

				e, ok := entries.Get("cn=foo")
				require.True(t, ok)
				require.NotContains(t, e.Attributes, "foo")
			},
		},
		{
			name: "delete specific value attribute",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {
						Dn:         "cn=foo",
						Attributes: map[string][]string{"foo": {"bar1", "bar2"}},
					},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyRequest{
					Dn: "cn=foo",
					Items: []ldap.ModificationItem{
						{
							Operation:    ldap.DeleteOperation,
							Modification: ldap.Modification{Type: "foo", Values: []string{"bar1"}},
						},
					},
				}))
				res := rr.Message.(*ldap.ModifyResponse)
				require.Equal(t, ldap.Success, res.ResultCode)

				e, ok := entries.Get("cn=foo")
				require.True(t, ok)
				require.Contains(t, e.Attributes, "foo")
				require.Equal(t, "bar2", e.Attributes["foo"][0])
			},
		},
		{
			name: "delete all values should also remove attribute",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {
						Dn:         "cn=foo",
						Attributes: map[string][]string{"foo": {"bar1", "bar2"}},
					},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyRequest{
					Dn: "cn=foo",
					Items: []ldap.ModificationItem{
						{
							Operation:    ldap.DeleteOperation,
							Modification: ldap.Modification{Type: "foo", Values: []string{"bar1", "bar2"}},
						},
					},
				}))
				res := rr.Message.(*ldap.ModifyResponse)
				require.Equal(t, ldap.Success, res.ResultCode)

				e, ok := entries.Get("cn=foo")
				require.True(t, ok)
				require.NotContains(t, e.Attributes, "foo")
			},
		},
		{
			name: "replace attribute with one value",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {
						Dn:         "cn=foo",
						Attributes: map[string][]string{"foo": {"bar1", "bar2"}},
					},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyRequest{
					Dn: "cn=foo",
					Items: []ldap.ModificationItem{
						{
							Operation:    ldap.ReplaceOperation,
							Modification: ldap.Modification{Type: "foo", Values: []string{"yuh"}},
						},
					},
				}))
				res := rr.Message.(*ldap.ModifyResponse)
				require.Equal(t, ldap.Success, res.ResultCode)

				e, ok := entries.Get("cn=foo")
				require.True(t, ok)
				require.Contains(t, e.Attributes, "foo")
				require.Equal(t, "yuh", e.Attributes["foo"][0])
			},
		},
		{
			name: "replace attribute with two value",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {
						Dn:         "cn=foo",
						Attributes: map[string][]string{"foo": {"bar1", "bar2"}},
					},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyRequest{
					Dn: "cn=foo",
					Items: []ldap.ModificationItem{
						{
							Operation:    ldap.ReplaceOperation,
							Modification: ldap.Modification{Type: "foo", Values: []string{"yuh1", "yuh2"}},
						},
					},
				}))
				res := rr.Message.(*ldap.ModifyResponse)
				require.Equal(t, ldap.Success, res.ResultCode)

				e, ok := entries.Get("cn=foo")
				require.True(t, ok)
				require.Contains(t, e.Attributes, "foo")
				require.Equal(t, []string{"yuh1", "yuh2"}, e.Attributes["foo"])
			},
		},
		{
			name: "not found",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {
						Dn: "cn=foo",
					},
				}),
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyRequest{
					Dn: "cn=bar",
					Items: []ldap.ModificationItem{
						{
							Operation:    ldap.AddOperation,
							Modification: ldap.Modification{Type: "foo", Values: []string{"bar"}},
						},
					},
				}))
				res := rr.Message.(*ldap.ModifyResponse)
				require.Equal(t, ldap.NoSuchObject, res.ResultCode)
				require.Equal(t, "", res.MatchedDn)
				require.Equal(t, "", res.Message)
			},
		},
		{
			name: "invalid value",
			config: &directory.Config{
				Info: directory.Info{Name: "foo"},
				Entries: convert(map[string]directory.Entry{
					"cn=foo": {
						Dn: "cn=foo",
					},
				}),
				Schema: &directory.Schema{AttributeTypes: map[string]*directory.AttributeType{
					"foo": {
						Syntax: "1.3.6.1.4.1.1466.115.121.1.7",
					},
				}},
			},
			fn: func(t *testing.T, h ldap.Handler, entries *sortedmap.LinkedHashMap[string, directory.Entry]) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.ModifyRequest{
					Dn: "cn=foo",
					Items: []ldap.ModificationItem{
						{
							Operation:    ldap.AddOperation,
							Modification: ldap.Modification{Type: "foo", Values: []string{"bar"}},
						},
					},
				}))
				res := rr.Message.(*ldap.ModifyResponse)
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
