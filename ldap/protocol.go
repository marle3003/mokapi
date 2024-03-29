package ldap

// Filters: https://ldapwiki.com/wiki/LDAP%20Filter%20Choices
// Search: https://ldapwiki.com/wiki/SearchRequest
// https://datatracker.ietf.org/doc/html/rfc4511

const (
	bindRequest    = 0
	bindResponse   = 1
	unbindRequest  = 2
	searchRequest  = 3
	searchResult   = 4
	searchDone     = 5
	abandonRequest = 16

	FilterAnd            = 0
	FilterOr             = 1
	FilterNot            = 2
	FilterEqualityMatch  = 3
	FilterSubstrings     = 4
	FilterGreaterOrEqual = 5
	FilterLessOrEqual    = 6
	FilterPresent        = 7
	FilterApproxMatch    = 8

	FilterSubstringsStartWith uint8 = 0
	FilterSubstringsAny       uint8 = 1
	FilterSubstringsEndWith   uint8 = 2

	// ScopeBaseObject examines only the level specified by the base DN (and none of its child entries)
	ScopeBaseObject = 0
	// ScopeSingleLevel examines only the level immediately below the base DN
	ScopeSingleLevel = 1
	// ScopeWholeSubtree examines the subtree below the base DN and includes the base DN level
	ScopeWholeSubtree = 2

	Success                uint8 = 0
	OperationsError        uint8 = 1
	ProtocolError          uint8 = 2
	SizeLimitExceeded      uint8 = 4
	AuthMethodNotSupported uint8 = 7
	CannotCancel           uint8 = 121
)

var OperatorText = map[int]string{
	FilterAnd:           "&",
	FilterOr:            "|",
	FilterNot:           "!",
	FilterEqualityMatch: "=",
	FilterPresent:       "=*",
}

var SubstringText = map[uint8]string{
	FilterSubstringsStartWith: "StartWith",
	FilterSubstringsAny:       "Any",
	FilterSubstringsEndWith:   "EndWith",
}

var StatusText = map[uint8]string{
	Success:                "Success",
	OperationsError:        "OperationsError",
	ProtocolError:          "ProtocolError",
	SizeLimitExceeded:      "SizeLimitExceeded",
	AuthMethodNotSupported: "AuthMethodNotSupported",
	CannotCancel:           "CannotCancel",
}
