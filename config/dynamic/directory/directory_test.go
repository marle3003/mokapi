package directory_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic/directory"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/ldap/ldaptest"
	"testing"
)

var testConfig = &directory.Config{
	Info: directory.Info{Name: "foo"},
	Entries: map[string]directory.Entry{
		"": {
			Dn: "dc=foo,dc=com",
			Attributes: map[string][]string{
				"objectclass": {"foo"},
			},
		},
		"user1": {
			Dn: "cn=user1,dc=foo,dc=com",
			Attributes: map[string][]string{
				"objectclass": {"foo"},
				"mail":        {"user1@foo.bar"},
				"number":      {"5"},
			},
		},
		"user2": {
			Dn: "cn=user2,dc=foo,dc=com",
			Attributes: map[string][]string{
				"objectclass": {"foo"},
				"mail":        {"user2@foo.bar"},
				"number":      {"10"},
			},
		},
	}}

func TestDirectory_ServeBind(t *testing.T) {
	testcases := []struct {
		name   string
		config *directory.Config
		fn     func(t *testing.T, h ldap.Handler)
	}{
		{
			name:   "anonymous bind",
			config: &directory.Config{Info: directory.Info{Name: "foo"}},
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.BindRequest{Version: 3, Auth: ldap.Simple}))
				res := rr.Message.(*ldap.BindResponse)
				require.Equal(t, ldap.Success, res.Result)
				require.Equal(t, "", res.MatchedDN)
				require.Equal(t, "", res.Message)
			},
		},
		{
			name:   "unsupported bind request",
			config: &directory.Config{},
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				r := &ldap.BindRequest{
					Version: 3,
					Auth:    3, // sasl
				}
				h.ServeLDAP(rr, &ldap.Request{Message: r})
				res := rr.Message.(*ldap.BindResponse)
				require.Equal(t, ldap.AuthMethodNotSupported, res.Result)
				require.Equal(t, "", res.MatchedDN)
				require.Equal(t, "server supports only simple auth method", res.Message)
			},
		},
		{
			name:   "unsupported ldap version",
			config: &directory.Config{},
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				r := &ldap.BindRequest{
					Version: 2,
					Auth:    ldap.Simple,
				}
				h.ServeLDAP(rr, &ldap.Request{Message: r})
				res := rr.Message.(*ldap.BindResponse)
				require.Equal(t, ldap.ProtocolError, res.Result)
				require.Equal(t, "", res.MatchedDN)
				require.Equal(t, "server supports only ldap version 3", res.Message)
			},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			d := directory.NewHandler(test.config, enginetest.NewEngine())
			test.fn(t, d)
		})
	}
}

func TestDirectory_ServeSearch(t *testing.T) {
	testcases := []struct {
		name   string
		config *directory.Config
		fn     func(t *testing.T, h ldap.Handler)
	}{
		{
			name:   "empty base object",
			config: &directory.Config{Info: directory.Info{Name: "foo"}},
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "",
					Filter: "(objectClass=*)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 0)
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name: "base object",
			config: &directory.Config{Info: directory.Info{Name: "foo"},
				Entries: map[string]directory.Entry{
					"": {Dn: "", Attributes: map[string][]string{"foo": {"bar"}}}},
			},
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "",
					Filter: "(objectClass=*)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.Equal(t, "", res.Results[0].Dn)
				require.Equal(t, []string{"bar"}, res.Results[0].Attributes["foo"])
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name: "objectclass=* scope=ScopeBaseObject",
			config: &directory.Config{Info: directory.Info{Name: "foo"},
				Entries: map[string]directory.Entry{
					"": {
						Dn: "dc=foo,dc=com",
						Attributes: map[string][]string{
							"objectclass": {"foo"},
						},
					},
					"not": {
						Dn: "not",
					},
				}},
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Filter: "(objectClass=*)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.Equal(t, "dc=foo,dc=com", res.Results[0].Dn)
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "objectclass=* scope=ScopeSingleLevel",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(objectClass=*)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 2)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.True(t, hasResult(res.Results, "cn=user2,dc=foo,dc=com"), "search result should contain user2")
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "(mail=user2@foo.bar)",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(mail=user2@foo.bar)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user2,dc=foo,dc=com"), "search result should contain user2")
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "(mail=user2*)",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(mail=user2*)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user2,dc=foo,dc=com"), "search result should contain user2")
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "(mail=*@foo.bar)",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(mail=*@foo.bar)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 2)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.True(t, hasResult(res.Results, "cn=user2,dc=foo,dc=com"), "search result should contain user2")
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "(mail=us*1*@*f*b*)",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(mail=us*1*@*f*b*)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "(&(mail=user1*)(objectclass=foo))",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(&(mail=user1*)(objectclass=foo))",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "(&(mail=user1*)(objectclass=bar))",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(&(mail=user1*)(objectclass=bar))",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 0)
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "(|(mail=user1*)(objectclass=bar))",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(|(mail=user1*)(objectclass=bar))",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "(!(mail=user2*))",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(!(mail=user2*))",
				}))
				res := rr.Message.(*ldap.SearchResponse)
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "(number>=6)",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(number>=6)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user2,dc=foo,dc=com"), "search result should contain user2")
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "(number>=5)",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(number>=5)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 2)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.True(t, hasResult(res.Results, "cn=user2,dc=foo,dc=com"), "search result should contain user2")
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "(mail>=5)",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(mail>=5)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 0)
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "(number<=5)",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(number<=5)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "(number<=4)",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(number<=4)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 0)
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "attributes",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN:     "cn=user1,dc=foo,dc=com",
					Scope:      ldap.ScopeBaseObject,
					Filter:     "(objectClass=*)",
					Attributes: []string{"mail"},
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.Len(t, res.Results[0].Attributes, 2, "mail and objectClass")
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name:   "sizeLimit=1",
			config: testConfig,
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN:    "dc=foo,dc=com",
					Scope:     ldap.ScopeWholeSubtree,
					Filter:    "(objectclass=foo)",
					SizeLimit: 1,
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name: "exceed size limit",
			config: &directory.Config{
				Info:      testConfig.Info,
				Address:   testConfig.Address,
				SizeLimit: 1,
				Entries:   testConfig.Entries,
			},
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN:    "dc=foo,dc=com",
					Scope:     ldap.ScopeWholeSubtree,
					Filter:    "(objectclass=foo)",
					SizeLimit: 1000,
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.SizeLimitExceeded, res.Status)
				require.Len(t, res.Results, 1)
				require.Equal(t, "SizeLimitExceeded", res.Message)
			},
		},
		{
			name: "options (cn;lang-en=John Doe)",
			config: &directory.Config{
				Entries: map[string]directory.Entry{
					"user2": {
						Dn: "cn=john,dc=foo,dc=com",
						Attributes: map[string][]string{
							"cn;lang-en": {"John Doe"},
						},
					},
				}},
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(cn;lang-en=John Doe)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=john,dc=foo,dc=com"), "search result should contain user2")
				require.Equal(t, "Success", res.Message)
			},
		},
		{
			name: "paging",
			config: &directory.Config{
				Entries: map[string]directory.Entry{
					"user1": {
						Dn: "cn=john,dc=foo,dc=com",
						Attributes: map[string][]string{
							"cn": {"John Doe"},
						},
					},
					"user2": {
						Dn: "cn=carol,dc=foo,dc=com",
						Attributes: map[string][]string{
							"cn": {"Carol Doe"},
						},
					},
				}},
			fn: func(t *testing.T, h ldap.Handler) {
				ctx := ldap.NewPagingFromContext(context.Background())

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequestWithContext(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(objectClass=*)",
					Controls: []ldap.Control{
						&ldap.PagedResultsControl{
							PageSize: 1,
						},
					},
				}, ctx))
				res := rr.Message.(*ldap.SearchResponse)

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.Equal(t, "Success", res.Message)
				require.NotNil(t, res.Controls[0].(*ldap.PagedResultsControl).Cookie)
				result := res.Results[0].Dn

				rr = ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequestWithContext(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: "(objectClass=*)",
					Controls: []ldap.Control{
						&ldap.PagedResultsControl{
							PageSize: 1,
							Cookie:   res.Controls[0].(*ldap.PagedResultsControl).Cookie,
						},
					},
				}, ctx))
				res = rr.Message.(*ldap.SearchResponse)
				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.Equal(t, "Success", res.Message)
				require.NotEqual(t, result, res.Results[0].Dn, "should return next page")
			},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			h := directory.NewHandler(test.config, enginetest.NewEngine())
			test.fn(t, h)
		})
	}
}

func hasResult(results []ldap.SearchResult, dn string) bool {
	for _, r := range results {
		if r.Dn == dn {
			return true
		}
	}
	return false
}
