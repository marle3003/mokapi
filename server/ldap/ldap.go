package ldap

const (
	ApplicationBindRequest       = 0
	ApplicationBindResponse      = 1
	ApplicationUnbindRequest     = 2
	ApplicationSearchRequest     = 3
	ApplicationSearchResultEntry = 4
	ApplicationSearchResultDone  = 5
	ApplicationAbandonRequest    = 16

	// https://ldapwiki.com/wiki/LDAP%20Filter%20Choices
	FilterAnd            = 0
	FilterOr             = 1
	FilterNot            = 2
	FilterEqualityMatch  = 3
	FilterSubstrings     = 4
	FilterGreaterOrEqual = 5
	FilterLessOrEqual    = 6
	FilterPresent        = 7
	FilterApproxMatch    = 8

	// https://ldapwiki.com/wiki/SearchRequest
	ScopeBaseObject   = 0
	ScopeSingleLevel  = 1
	ScopeWholeSubtree = 2

	LDAPResultSuccess = 0
)

var RootDSE = map[string][]string{
	"supportedLDAPVersion": []string{"3"},
	"subSchemaSubentry":    []string{"cn=schema"},
}
