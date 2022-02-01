package ldap_test

import (
	"github.com/stretchr/testify/require"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
	config "mokapi/config/dynamic/ldap"
	"mokapi/server/ldap"
	"mokapi/server/ldap/ldaptest"
	"testing"
)

func TestDirectory_ServeBind(t *testing.T) {
	testcases := []struct {
		name   string
		config *config.Config
		fn     func(t *testing.T, d *ldap.Directory)
	}{
		{
			name:   "anonymous bind",
			config: &config.Config{},
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSimpleBindRequest(0, 3, "", ""))
				require.Equal(t, int64(0), rr.Responses[0].MessageId)
				require.Equal(t, ldap.ResultSuccess, rr.Responses[0].Body.Children[0].Value)
				require.Equal(t, "", rr.Responses[0].Body.Children[1].Value)
				require.Equal(t, "", rr.Responses[0].Body.Children[2].Value)
			},
		},
		{
			name:   "unsupported bind request",
			config: &config.Config{},
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				r := ldaptest.NewSimpleBindRequest(0, 3, "", "")
				// switch to sasl
				r.Body.Children[2].Tag = 3
				d.Serve(rr, r)
				require.Equal(t, int64(0), rr.Responses[0].MessageId)
				require.Equal(t, ldap.AuthMethodNotSupported, rr.Responses[0].Body.Children[0].Value)
				require.Equal(t, "", rr.Responses[0].Body.Children[1].Value)
				require.Equal(t, "server supports only simple auth method", rr.Responses[0].Body.Children[2].Value)
			},
		},
		{
			name:   "unsupported ldap version",
			config: &config.Config{},
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				r := ldaptest.NewSimpleBindRequest(0, 3, "", "")
				// switch version
				r.Body.Children[0].Value = int64(2)
				d.Serve(rr, r)
				require.Equal(t, int64(0), rr.Responses[0].MessageId)
				require.Equal(t, ldap.ProtocolError, rr.Responses[0].Body.Children[0].Value)
				require.Equal(t, "", rr.Responses[0].Body.Children[1].Value)
				require.Equal(t, "server supports only ldap version 3", rr.Responses[0].Body.Children[2].Value)
			},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			d := ldap.NewDirectory(test.config)
			test.fn(t, d)
		})
	}
}

func TestDirectory_ServeSearch(t *testing.T) {
	testcases := []struct {
		name   string
		config *config.Config
		fn     func(t *testing.T, d *ldap.Directory)
	}{
		{
			name:   "empty base object",
			config: &config.Config{},
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "",
					Filter: anyObjectClass(),
				}))
				require.Equal(t, int64(0), rr.Responses[0].MessageId)
				require.Equal(t, "", rr.Responses[0].Body.Children[0].Value) // object name
				require.Len(t, rr.Responses[0].Body.Children[1].Children, 0) // attributes

				require.Equal(t, ldap.ResultSuccess, rr.Responses[1].Body.Children[0].Value)
				require.Equal(t, "", rr.Responses[1].Body.Children[1].Value)
				require.Equal(t, "", rr.Responses[1].Body.Children[2].Value)
			},
		},
		{
			name:   "base object",
			config: &config.Config{Root: config.Entry{Attributes: map[string][]string{"foo": {"bar"}}}},
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "",
					Filter: anyObjectClass(),
				}))
				require.Equal(t, int64(0), rr.Responses[0].MessageId)
				require.Equal(t, "", rr.Responses[0].Body.Children[0].Value) // object name
				require.Len(t, rr.Responses[0].Body.Children[1].Children, 1) // attributes
				attr := rr.Responses[0].Body.Children[1].Children[0]
				require.Equal(t, "foo", attr.Children[0].Value)             // name
				require.Equal(t, "bar", attr.Children[1].Children[0].Value) // first value

				require.Equal(t, ldap.ResultSuccess, rr.Responses[1].Body.Children[0].Value)
				require.Equal(t, "", rr.Responses[1].Body.Children[1].Value)
				require.Equal(t, "", rr.Responses[1].Body.Children[2].Value)
			},
		},
		{
			name: "objectclass=* scope=ScopeBaseObject",
			config: &config.Config{Entries: map[string]config.Entry{
				"": {
					Dn: "dc=foo,dc=com",
					Attributes: map[string][]string{
						"objectclass": {"foo"},
					},
				},
			}},
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Filter: anyObjectClass(),
				}))

				require.Len(t, rr.Responses, 2)

				require.Equal(t, int64(0), rr.Responses[0].MessageId)
				require.Equal(t, "dc=foo,dc=com", rr.Responses[0].Body.Children[0].Value) // object name
				require.Len(t, rr.Responses[0].Body.Children[1].Children, 1)              // attributes
				attr := rr.Responses[0].Body.Children[1].Children[0]
				require.Equal(t, "objectclass", attr.Children[0].Value)     // name
				require.Equal(t, "foo", attr.Children[1].Children[0].Value) // first value

				require.Equal(t, ldap.ResultSuccess, rr.Responses[1].Body.Children[0].Value)
				require.Equal(t, "", rr.Responses[1].Body.Children[1].Value)
				require.Equal(t, "", rr.Responses[1].Body.Children[2].Value)
			},
		},
		{
			name: "objectclass=* scope=ScopeBaseObject only return base",
			config: &config.Config{Entries: map[string]config.Entry{
				"": {
					Dn: "dc=foo,dc=com",
					Attributes: map[string][]string{
						"objectclass": {"foo"},
					},
				},
				"user": {
					Dn: "cn=user,dc=foo,dc=com",
					Attributes: map[string][]string{
						"objectclass": {"foo"},
					},
				},
			}},
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Filter: anyObjectClass(),
				}))

				require.Len(t, rr.Responses, 2)

				require.Equal(t, int64(0), rr.Responses[0].MessageId)
				require.Equal(t, "dc=foo,dc=com", rr.Responses[0].Body.Children[0].Value) // object name
				require.Len(t, rr.Responses[0].Body.Children[1].Children, 1)              // attributes
				attr := rr.Responses[0].Body.Children[1].Children[0]
				require.Equal(t, "objectclass", attr.Children[0].Value)     // name
				require.Equal(t, "foo", attr.Children[1].Children[0].Value) // first value

				require.Equal(t, ldap.ResultSuccess, rr.Responses[1].Body.Children[0].Value)
				require.Equal(t, "", rr.Responses[1].Body.Children[1].Value)
				require.Equal(t, "", rr.Responses[1].Body.Children[2].Value)
			},
		},
		{
			name: "objectclass=* scope=ScopeSingleLevel",
			config: &config.Config{Entries: map[string]config.Entry{
				"": {
					Dn: "dc=foo,dc=com",
					Attributes: map[string][]string{
						"objectclass": {"foo"},
					},
				},
				"user": {
					Dn: "cn=user,dc=foo,dc=com",
					Attributes: map[string][]string{
						"objectclass": {"foo"},
					},
				},
			}},
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: anyObjectClass(),
				}))

				require.Len(t, rr.Responses, 2)

				require.Equal(t, int64(0), rr.Responses[0].MessageId)
				require.Equal(t, "cn=user,dc=foo,dc=com", rr.Responses[0].Body.Children[0].Value) // object name
				require.Len(t, rr.Responses[0].Body.Children[1].Children, 1)                      // attributes
				attr := rr.Responses[0].Body.Children[1].Children[0]
				require.Equal(t, "objectclass", attr.Children[0].Value)     // name
				require.Equal(t, "foo", attr.Children[1].Children[0].Value) // first value

				require.Equal(t, ldap.ResultSuccess, rr.Responses[1].Body.Children[0].Value)
				require.Equal(t, "", rr.Responses[1].Body.Children[1].Value)
				require.Equal(t, "", rr.Responses[1].Body.Children[2].Value)
			},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			d := ldap.NewDirectory(test.config)
			test.fn(t, d)
		})
	}
}

func anyObjectClass() *ber.Packet {
	body := ber.NewString(
		ber.ClassContext,
		ber.TypePrimitive,
		ldap.FilterPresent,
		"objectclass",
		"Present")

	return body
}

func equals() *ber.Packet {
	body := ber.Encode(
		ber.ClassContext,
		ber.TypeConstructed,
		ldap.FilterEqualityMatch,
		nil,
		"Equality Match")

	body.AppendChild(ber.NewString(
		ber.ClassContext,
		ber.TypeConstructed,
		ber.TagOctetString,
		"foo",
		"Attribute"))

	body.AppendChild(ber.NewString(
		ber.ClassContext,
		ber.TypeConstructed,
		ber.TagOctetString,
		"bar",
		"Condition"))

	return body
}
