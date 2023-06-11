package directory

import (
	"mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/runtime/events"
)

type LdapSearchLog struct {
	Request  *ldap.SearchRequest `json:"request"`
	Response *LdapSearchResponse `json:"response"`
	Duration int64               `json:"duration"`
	Actions  []*common.Action    `json:"actions"`
}

type LdapSearchResponse struct {
	Status  string             `json:"status"`
	Results []LdapSearchResult `json:"results"`
}

type LdapSearchResult struct {
	Dn         string              `json:"dn"`
	Attributes map[string][]string `json:"attributes"`
}

func NewLogEvent(r *ldap.SearchRequest, traits events.Traits) *LdapSearchLog {
	event := &LdapSearchLog{
		Request:  r,
		Response: &LdapSearchResponse{},
		Duration: 0,
		Actions:  nil,
	}
	_ = events.Push(event, traits.WithNamespace("ldap"))
	return event
}
