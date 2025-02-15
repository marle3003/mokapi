package directory_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/directory"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/engine/enginetest"
	"mokapi/ldap"
	"mokapi/ldap/ldaptest"
	"mokapi/try"
	"testing"
)

func TestSearch_Schema(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		reader dynamic.Reader
		test   func(t *testing.T, h ldap.Handler, err error)
	}{
		{
			name:  "caseIgnoreMatch",
			input: `{ "files": [ "./schema.ldif", "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/schema.ldif": {Raw: []byte("dn: \nsubschemaSubentry: cn=schema\n\ndn: cn=schema\nattributeTypes: ( 2.5.4.3 NAME 'cn' DESC 'Common Name' EQUALITY caseIgnoreMatch SYNTAX 1.3.6.1.4.1.1466.115.121.1.15 SINGLE-VALUE )")},
				"file:/users.ldif":  {Raw: []byte("dn: cn=user\ncn: UsEr")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(cn=user)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 1)
			},
		},
		{
			name:  "caseIgnoreMatch",
			input: `{ "files": [ "./schema.ldif", "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/schema.ldif": {Raw: []byte("dn: \nsubschemaSubentry: cn=schema\n\ndn: cn=schema\nattributeTypes: ( 1.3.6.1.1.1.1.0 NAME 'uidNumber' DESC 'User ID' \n  EQUALITY integerMatch \n  SYNTAX 1.3.6.1.4.1.1466.115.121.1.27 SINGLE-VALUE )")},
				"file:/users.ldif":  {Raw: []byte("dn: cn=user\nuidNumber: 1001")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(uidNumber=0001001)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 1)
			},
		},
		{
			name:  "octetStringMatch",
			input: `{ "files": [ "./schema.ldif", "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/schema.ldif": {Raw: []byte("dn: \nsubschemaSubentry: cn=schema\n\ndn: cn=schema\nattributeTypes: ( 1.3.6.1.4.1.99999.1.1 NAME 'customBinaryAttribute'\n  DESC 'Example attribute storing raw binary data'\n  EQUALITY octetStringMatch\n  SYNTAX 1.3.6.1.4.1.1466.115.121.1.40 )")},
				"file:/users.ldif":  {Raw: []byte("dn: cn=user\ncustomBinaryAttribute:: bXlTZWNyZXREYXRh")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(customBinaryAttribute=bXlTZWNyZXREYXRh)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 1)
			},
		},
		{
			name:  "octetStringMatch not matching",
			input: `{ "files": [ "./schema.ldif", "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/schema.ldif": {Raw: []byte("dn: \nsubschemaSubentry: cn=schema\n\ndn: cn=schema\nattributeTypes: ( 1.3.6.1.4.1.99999.1.1 NAME 'customBinaryAttribute'\n  DESC 'Example attribute storing raw binary data'\n  EQUALITY octetStringMatch\n  SYNTAX 1.3.6.1.4.1.1466.115.121.1.40 )")},
				"file:/users.ldif":  {Raw: []byte("dn: cn=user\ncustomBinaryAttribute:: bXlTZWNyZXREYXRh")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(customBinaryAttribute=bXlTZGZyZWREYXRh)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 0)
			},
		},
		{
			name:  "booleanMatch",
			input: `{ "files": [ "./schema.ldif", "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/schema.ldif": {Raw: []byte("dn: \nsubschemaSubentry: cn=schema\n\ndn: cn=schema\nattributeTypes: ( 1.3.6.1.4.1.99999.2.1 NAME 'isActive'\n  DESC 'Indicates whether a user is active or not'\n  EQUALITY booleanMatch\n  SYNTAX 1.3.6.1.4.1.1466.115.121.1.7 )")},
				"file:/users.ldif":  {Raw: []byte("dn: cn=user\nisActive: TRUE")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(isActive=TRUE)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 1)
			},
		},
		{
			name:  "booleanMatch case sensitive",
			input: `{ "files": [ "./schema.ldif", "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/schema.ldif": {Raw: []byte("dn: \nsubschemaSubentry: cn=schema\n\ndn: cn=schema\nattributeTypes: ( 1.3.6.1.4.1.99999.2.1 NAME 'isActive'\n  DESC 'Indicates whether a user is active or not'\n  EQUALITY booleanMatch\n  SYNTAX 1.3.6.1.4.1.1466.115.121.1.7 )")},
				"file:/users.ldif":  {Raw: []byte("dn: cn=user\nisActive: TRUE")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(isActive=true)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 0)
			},
		},
		{
			name:  "numericStringMatch",
			input: `{ "files": [ "./schema.ldif", "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/schema.ldif": {Raw: []byte("dn: \nsubschemaSubentry: cn=schema\n\ndn: cn=schema\nattributeTypes: ( 1.3.6.1.4.1.99999.1.1\n  NAME 'phoneNumber'\n  DESC 'A phone number as a numeric string'\n  EQUALITY numericStringMatch\n  SYNTAX 1.3.6.1.4.1.1466.115.121.1.36\n  SINGLE-VALUE )")},
				"file:/users.ldif":  {Raw: []byte("dn: cn=user\nphoneNumber: 00123456789")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(phoneNumber=00123456789)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 1)
			},
		},
		{
			name:  "numericStringMatch not matching",
			input: `{ "files": [ "./schema.ldif", "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/schema.ldif": {Raw: []byte("dn: \nsubschemaSubentry: cn=schema\n\ndn: cn=schema\nattributeTypes: ( 1.3.6.1.4.1.99999.1.1\n  NAME 'phoneNumber'\n  DESC 'A phone number as a numeric string'\n  EQUALITY numericStringMatch\n  SYNTAX 1.3.6.1.4.1.1466.115.121.1.36\n  SINGLE-VALUE )")},
				"file:/users.ldif":  {Raw: []byte("dn: cn=user\nphoneNumber: 00123456789")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(phoneNumber=123456789)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 0)
			},
		},
		{
			name:  "distinguishedNameMatch",
			input: `{ "files": [ "./schema.ldif", "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/schema.ldif": {Raw: []byte("dn: \nsubschemaSubentry: cn=schema\n\ndn: cn=schema\nattributeTypes: ( 1.3.6.1.4.1.99999.1.3\n  NAME 'managerDN'\n  DESC 'A manager distinguished name (DN)'\n  EQUALITY distinguishedNameMatch\n  SYNTAX 1.3.6.1.4.1.1466.115.121.1.12\n  SINGLE-VALUE )")},
				"file:/users.ldif":  {Raw: []byte("dn: cn=user\nmanagerDN: cn=manager1,ou=employees,dc=example,dc=com")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(managerDN=cn=manager1,ou=employees,dc=example,dc=com)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 1)
			},
		},
		{
			name:  "telephoneNumberMatch",
			input: `{ "files": [ "./schema.ldif", "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/schema.ldif": {Raw: []byte("dn: \nsubschemaSubentry: cn=schema\n\ndn: cn=schema\nattributeTypes: ( 2.5.4.20\n  NAME 'telephoneNumber'\n  DESC 'Telephone number'\n  EQUALITY telephoneNumberMatch \n  SYNTAX 1.3.6.1.4.1.1466.115.121.1.50\n  SINGLE-VALUE )")},
				"file:/users.ldif":  {Raw: []byte("dn: cn=user\ntelephoneNumber: +1 555 123 4567")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(telephoneNumber=+15551234567)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 1)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var c *directory.Config
			err := json.Unmarshal([]byte(tc.input), &c)
			if err != nil {
				tc.test(t, nil, err)
			} else {
				err = c.Parse(&dynamic.Config{Data: c, Info: dynamic.ConfigInfo{Url: try.MustUrl("file:/foo.yml")}}, tc.reader)
				tc.test(t, directory.NewHandler(c, enginetest.NewEngine()), err)
			}
		})
	}
}

func TestSearch(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		reader dynamic.Reader
		test   func(t *testing.T, h ldap.Handler, err error)
	}{
		{
			name:  "presence",
			input: `{ "files": [ "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/users.ldif": {Raw: []byte("dn: cn=user\nmail: user@foo.com")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(mail=*)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 1)
			},
		},
		{
			name:  "greater than",
			input: `{ "files": [ "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/users.ldif": {Raw: []byte("dn: cn=user\nage: 30")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(age>10)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 1)
			},
		},
		{
			name:  "less than",
			input: `{ "files": [ "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/users.ldif": {Raw: []byte("dn: cn=user\nage: 30")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(age<31)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 1)
			},
		},
		{
			name:  "approximate match",
			input: `{ "files": [ "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/users.ldif": {Raw: []byte("dn: cn=user\ncn: Smith")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(cn~=Smit)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 1)
			},
		},
		{
			name:  "approximate match 2",
			input: `{ "files": [ "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/users.ldif": {Raw: []byte("dn: cn=user\ndescription: Software Developers")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(description~=Developers)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 1)
			},
		},
		{
			name:  "approximate match 3",
			input: `{ "files": [ "./users.ldif" ] }`,
			reader: &dynamictest.Reader{Data: map[string]*dynamic.Config{
				"file:/users.ldif": {Raw: []byte("dn: cn=user\ndescription: Software Developers")},
			}},
			test: func(t *testing.T, h ldap.Handler, err error) {
				require.NoError(t, err)

				rr := ldaptest.NewRecorder()
				h.ServeLDAP(rr, ldaptest.NewRequest(0, &ldap.SearchRequest{
					Scope:  ldap.ScopeWholeSubtree,
					Filter: "(description~=software developer)",
				}))
				res := rr.Message.(*ldap.SearchResponse)

				require.Len(t, res.Results, 1)
			},
		},
	}

	t.Parallel()
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var c *directory.Config
			err := json.Unmarshal([]byte(tc.input), &c)
			if err != nil {
				tc.test(t, nil, err)
			} else {
				err = c.Parse(&dynamic.Config{Data: c, Info: dynamic.ConfigInfo{Url: try.MustUrl("file:/foo.yml")}}, tc.reader)
				tc.test(t, directory.NewHandler(c, enginetest.NewEngine()), err)
			}
		})
	}
}
