package directory

import (
	"mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/runtime/events"
)

type LdapSearchLog struct {
	Request  *SearchRequest   `json:"request"`
	Response *SearchResponse  `json:"response"`
	Duration int64            `json:"duration"`
	Actions  []*common.Action `json:"actions"`
}

type SearchRequest struct {
	BaseDN            string        `json:"baseDN"`
	Scope             string        `json:"scope"`
	DereferencePolicy int64         `json:"dereferencePolicy"`
	SizeLimit         int64         `json:"sizeLimit"`
	TimeLimit         int64         `json:"timeLimit"`
	TypesOnly         bool          `json:"typesOnly"`
	Filter            string        `json:"filter"`
	Attributes        []string      `json:"attributes"`
	Controls          []interface{} `json:"controls"`
}

type SearchResponse struct {
	Status  string         `json:"status"`
	Results []SearchResult `json:"results"`
}

type SearchResult struct {
	Dn         string              `json:"dn"`
	Attributes map[string][]string `json:"attributes"`
}

func NewLogEvent(r *ldap.SearchRequest, traits events.Traits) *LdapSearchLog {
	event := &LdapSearchLog{
		Request:  fromRequest(r),
		Response: &SearchResponse{},
		Duration: 0,
		Actions:  nil,
	}
	_ = events.Push(event, traits.WithNamespace("ldap"))
	return event
}

func fromRequest(r *ldap.SearchRequest) *SearchRequest {
	return &SearchRequest{
		BaseDN:            r.BaseDN,
		Scope:             ldap.ScopeText[uint8(r.Scope)],
		DereferencePolicy: r.DereferencePolicy,
		SizeLimit:         r.SizeLimit,
		TimeLimit:         r.TimeLimit,
		TypesOnly:         r.TypesOnly,
		Filter:            r.Filter,
		Attributes:        r.Attributes,
	}
}
