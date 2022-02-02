package ldap_test

import (
	"github.com/stretchr/testify/require"
	ber "gopkg.in/go-asn1-ber/asn1-ber.v1"
	config "mokapi/config/dynamic/ldap"
	"mokapi/server/ldap"
	"mokapi/server/ldap/ldaptest"
	"strings"
	"testing"
)

var testConfig = &config.Config{Entries: map[string]config.Entry{
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
		},
	},
	"user2": {
		Dn: "cn=user2,dc=foo,dc=com",
		Attributes: map[string][]string{
			"objectclass": {"foo"},
			"mail":        {"user2@foo.bar"},
		},
	},
}}

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

func TestDirectory_ServeUnbind(t *testing.T) {
	testcases := []struct {
		name   string
		config *config.Config
		fn     func(t *testing.T, d *ldap.Directory)
	}{
		{
			name:   "unbind",
			config: &config.Config{},
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewUnbindRequest(0))
				require.Len(t, rr.Responses, 0)
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

func TestDirectory_ServeAbandon(t *testing.T) {
	testcases := []struct {
		name   string
		config *config.Config
		fn     func(t *testing.T, d *ldap.Directory)
	}{
		{
			name:   "unbind",
			config: &config.Config{},
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewAbandonRequest(0))
				require.Len(t, rr.Responses, 0)
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
			name:   "objectclass=* scope=ScopeBaseObject only return base",
			config: testConfig,
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
			name:   "objectclass=* scope=ScopeSingleLevel",
			config: testConfig,
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: anyObjectClass(),
				}))

				require.Len(t, rr.Responses, 3)

				// order of results is not guaranteed
				results := []string{
					rr.Responses[0].Body.Children[0].Value.(string),
					rr.Responses[1].Body.Children[0].Value.(string),
				}
				require.Contains(t, results, "cn=user1,dc=foo,dc=com")
				require.Contains(t, results, "cn=user2,dc=foo,dc=com")

				// check message ids
				require.Equal(t, int64(0), rr.Responses[0].MessageId)
				require.Equal(t, int64(0), rr.Responses[1].MessageId)
				require.Equal(t, int64(0), rr.Responses[2].MessageId)

				require.Len(t, rr.Responses[0].Body.Children[1].Children, 2) // attributes
				attr := rr.Responses[0].Body.Children[1].Children[0]
				require.Equal(t, "objectclass", attr.Children[0].Value)     // name
				require.Equal(t, "foo", attr.Children[1].Children[0].Value) // first value

				done := rr.Responses[2]
				require.Equal(t, ldap.ResultSuccess, done.Body.Children[0].Value)
				require.Equal(t, "", done.Body.Children[1].Value)
				require.Equal(t, "", done.Body.Children[2].Value)
			},
		},
		{
			name:   "equals",
			config: testConfig,
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: equals("mail", "user2@foo.bar"),
				}))

				require.Len(t, rr.Responses, 2)
				require.Equal(t, "cn=user2,dc=foo,dc=com", rr.Responses[0].Body.Children[0].Value)
			},
		},
		{
			name:   "not equals",
			config: testConfig,
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: equals("mail", "user3@foo.bar"),
				}))

				require.Len(t, rr.Responses, 1)
			},
		},
		{
			name:   "starts with",
			config: testConfig,
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: substring("mail", "user2*"),
				}))

				require.Len(t, rr.Responses, 2)
			},
		},
		{
			name:   "ends with",
			config: testConfig,
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: substring("mail", "*foo.bar"),
				}))

				require.Len(t, rr.Responses, 3)
			},
		},
		{
			name:   "any",
			config: testConfig,
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: substring("mail", "us*1*@*f*b*"),
				}))

				require.Len(t, rr.Responses, 2)
			},
		},
		{
			name:   "and",
			config: testConfig,
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: and(substring("mail", "user1*"), substring("mail", "*foo.bar")),
				}))

				require.Len(t, rr.Responses, 2)
			},
		},
		{
			name:   "or",
			config: testConfig,
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: or(substring("mail", "user1*"), substring("mail", "user2*")),
				}))

				require.Len(t, rr.Responses, 3)
			},
		},
		{
			name:   "not",
			config: testConfig,
			fn: func(t *testing.T, d *ldap.Directory) {
				rr := ldaptest.NewRecorder()
				d.Serve(rr, ldaptest.NewSearchRequest(0, &ldap.SearchRequest{
					BaseDN: "dc=foo,dc=com",
					Scope:  ldap.ScopeSingleLevel,
					Filter: not(substring("mail", "user1*")),
				}))

				require.Len(t, rr.Responses, 2)
				require.Equal(t, "cn=user2,dc=foo,dc=com", rr.Responses[0].Body.Children[0].Value)
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

func TestStart(t *testing.T) {
	d := ldap.NewDirectory(&config.Config{Address: ":12345"})
	d.Start()
	defer d.Close()

	r := ldaptest.NewSimpleBindRequest(0, 3, "", "")
	client := ldaptest.NewClient("127.0.0.1:12345")
	res, err := client.Send(r)
	require.NoError(t, err)
	require.Equal(t, ldap.ResultSuccess, res.Body.Children[0].Value)
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

func equals(attr, value string) *ber.Packet {
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
		attr,
		"Attribute"))

	body.AppendChild(ber.NewString(
		ber.ClassContext,
		ber.TypeConstructed,
		ber.TagOctetString,
		value,
		"Condition"))

	return body
}

func substring(attr, value string) *ber.Packet {
	body := ber.Encode(
		ber.ClassContext,
		ber.TypeConstructed,
		ldap.FilterSubstrings,
		nil,
		"Substring")

	body.AppendChild(ber.NewString(
		ber.ClassContext,
		ber.TypeConstructed,
		ber.TagOctetString,
		attr,
		"Attribute"))

	seq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Substrings")
	body.AppendChild(seq)

	values := strings.Split(value, "*")
	for i, v := range values {
		if len(v) == 0 {
			continue
		}
		var tag ber.Tag
		switch i {
		case 0:
			tag = ldap.FilterSubstringsStartWith
		case len(values) - 1:
			tag = ldap.FilterSubstringsEndWith
		default:
			tag = ldap.FilterSubstringsAny
		}

		seq.AppendChild(ber.NewString(
			ber.ClassContext,
			ber.TypeConstructed,
			tag,
			v,
			"Condition"))
	}

	return body
}

func and(p1, p2 *ber.Packet) *ber.Packet {
	body := ber.Encode(
		ber.ClassContext,
		ber.TypeConstructed,
		ldap.FilterAnd,
		nil,
		"And")

	body.AppendChild(p1)
	body.AppendChild(p2)
	return body
}

func or(p1, p2 *ber.Packet) *ber.Packet {
	body := ber.Encode(
		ber.ClassContext,
		ber.TypeConstructed,
		ldap.FilterOr,
		nil,
		"Or")

	body.AppendChild(p1)
	body.AppendChild(p2)
	return body
}

func not(p *ber.Packet) *ber.Packet {
	body := ber.Encode(
		ber.ClassContext,
		ber.TypeConstructed,
		ldap.FilterNot,
		nil,
		"Not")

	body.AppendChild(p)
	return body
}
