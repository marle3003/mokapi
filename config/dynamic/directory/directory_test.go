package directory

import (
	"github.com/stretchr/testify/require"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/ldap/ldaptest"
	"testing"
)

var testConfig = &Config{
	Info: Info{Name: "foo"},
	Entries: map[string]Entry{
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
		config *Config
		fn     func(t *testing.T, h ldap.Handler)
	}{
		{
			name:   "anonymous bind",
			config: &Config{Info: Info{Name: "foo"}},
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
			config: &Config{},
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
			config: &Config{},
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
			d := NewHandler(test.config, enginetest.NewEngine())
			test.fn(t, d)
		})
	}
}

func TestDirectory_ServeSearch(t *testing.T) {
	testcases := []struct {
		name   string
		config *Config
		fn     func(t *testing.T, h ldap.Handler)
	}{
		{
			name:   "empty base object",
			config: &Config{Info: Info{Name: "foo"}},
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "",
					Filter: "(objectClass=*)",
				}))
				res := rr.Message.(*ldap.SearchResponse)
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.Equal(t, "", res.Message)
			},
		},
		{
			name: "base object",
			config: &Config{Info: Info{Name: "foo"},
				Root: Entry{Dn: "root", Attributes: map[string][]string{"foo": {"bar"}}}},
			fn: func(t *testing.T, h ldap.Handler) {
				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					BaseDN: "",
					Filter: "(objectClass=*)",
				}))
				res := rr.Message.(*ldap.SearchResponse)
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.Equal(t, "root", res.Results[0].Dn)
				require.Equal(t, []string{"bar"}, res.Results[0].Attributes["foo"])
				require.Equal(t, "", res.Message)
			},
		},
		{
			name: "objectclass=* scope=ScopeBaseObject",
			config: &Config{Info: Info{Name: "foo"},
				Entries: map[string]Entry{
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.Equal(t, "dc=foo,dc=com", res.Results[0].Dn)
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 2)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.True(t, hasResult(res.Results, "cn=user2,dc=foo,dc=com"), "search result should contain user2")
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user2,dc=foo,dc=com"), "search result should contain user2")
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user2,dc=foo,dc=com"), "search result should contain user2")
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 2)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.True(t, hasResult(res.Results, "cn=user2,dc=foo,dc=com"), "search result should contain user2")
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 0)
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.Equal(t, "", res.Message)
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
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user2,dc=foo,dc=com"), "search result should contain user2")
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 2)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.True(t, hasResult(res.Results, "cn=user2,dc=foo,dc=com"), "search result should contain user2")
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 0)
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.True(t, hasResult(res.Results, "cn=user1,dc=foo,dc=com"), "search result should contain user1")
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 0)
				require.Equal(t, "", res.Message)
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
				_ = res

				require.Equal(t, ldap.Success, res.Status)
				require.Len(t, res.Results, 1)
				require.Len(t, res.Results[0].Attributes, 2, "mail and objectClass")
				require.Equal(t, "", res.Message)
			},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			h := NewHandler(test.config, enginetest.NewEngine())
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
