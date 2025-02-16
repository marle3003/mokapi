package directory

import (
	"mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/runtime/events"
)

type SearchLog struct {
	Request  *SearchRequest   `json:"request"`
	Response *SearchResponse  `json:"response"`
	Duration int64            `json:"duration"`
	Actions  []*common.Action `json:"actions"`
}

type SearchRequest struct {
	Operation         string        `json:"operation"`
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

func NewSearchLogEvent(r *ldap.SearchRequest, traits events.Traits) *SearchLog {
	event := &SearchLog{
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
		Operation:         "Search",
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

type AddLog struct {
	Request  *AddRequest      `json:"request"`
	Response *Response        `json:"response"`
	Duration int64            `json:"duration"`
	Actions  []*common.Action `json:"actions"`
}

type AddRequest struct {
	Operation  string      `json:"operation"`
	Dn         string      `json:"dn"`
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Type   string   `json:"type"`
	Values []string `json:"values"`
}

type Response struct {
	Status    string `json:"status"`
	MatchedDn string `json:"matchedDn"`
	Message   string `json:"message"`
}

func NewAddLogEvent(req *ldap.AddRequest, res *ldap.AddResponse, traits events.Traits) *AddLog {
	var attr []Attribute
	for _, v := range req.Attributes {
		attr = append(attr, Attribute{
			Type:   v.Type,
			Values: v.Values,
		})
	}

	l := &AddLog{
		Request: &AddRequest{
			Operation:  "Add",
			Dn:         req.Dn,
			Attributes: attr,
		},
		Response: &Response{
			Status:    ldap.StatusText[res.ResultCode],
			MatchedDn: res.MatchedDn,
			Message:   res.Message,
		},
		Duration: 0,
		Actions:  nil,
	}
	_ = events.Push(l, traits.WithNamespace("ldap"))

	return l
}

type ModifyLog struct {
	Request  *ModifyRequest   `json:"request"`
	Response *Response        `json:"response"`
	Duration int64            `json:"duration"`
	Actions  []*common.Action `json:"actions"`
}

type ModifyRequest struct {
	Operation string   `json:"operation"`
	Dn        string   `json:"dn"`
	Items     []Modify `json:"items"`
}

type Modify struct {
	Modification string     `json:"modification"`
	Attribute    *Attribute `json:"attribute"`
}

func NewModifyLogEvent(req *ldap.ModifyRequest, res *ldap.ModifyResponse, traits events.Traits) *ModifyLog {
	var items []Modify
	for _, v := range req.Items {
		items = append(items, Modify{
			Modification: ldap.ModifyOperationText[v.Operation],
			Attribute: &Attribute{
				Type:   v.Modification.Type,
				Values: v.Modification.Values,
			},
		})
	}

	l := &ModifyLog{
		Request: &ModifyRequest{
			Operation: "Modify",
			Dn:        req.Dn,
			Items:     items,
		},
		Response: &Response{
			Status:    ldap.StatusText[res.ResultCode],
			MatchedDn: res.MatchedDn,
			Message:   res.Message,
		},
		Duration: 0,
		Actions:  nil,
	}
	_ = events.Push(l, traits.WithNamespace("ldap"))

	return l
}

type DeleteLog struct {
	Request  *DeleteRequest   `json:"request"`
	Response *Response        `json:"response"`
	Duration int64            `json:"duration"`
	Actions  []*common.Action `json:"actions"`
}

type DeleteRequest struct {
	Operation string `json:"operation"`
	Dn        string `json:"dn"`
}

func NewDeleteLogEvent(req *ldap.DeleteRequest, res *ldap.DeleteResponse, traits events.Traits) *DeleteLog {
	l := &DeleteLog{
		Request: &DeleteRequest{
			Operation: "Delete",
			Dn:        req.Dn,
		},
		Response: &Response{
			Status:    ldap.StatusText[res.ResultCode],
			MatchedDn: res.MatchedDn,
			Message:   res.Message,
		},
		Duration: 0,
		Actions:  nil,
	}
	_ = events.Push(l, traits.WithNamespace("ldap"))

	return l
}

type ModifyDNLog struct {
	Request  *ModifyDNRequest `json:"request"`
	Response *Response        `json:"response"`
	Duration int64            `json:"duration"`
	Actions  []*common.Action `json:"actions"`
}

type ModifyDNRequest struct {
	Operation     string `json:"operation"`
	Dn            string `json:"dn"`
	NewRdn        string `json:"newRdn"`
	DeleteOldDn   bool   `json:"deleteOldDn"`
	NewSuperiorDn string `json:"newSuperiorDn"`
}

func NewModifyDNLogEvent(req *ldap.ModifyDNRequest, res *ldap.ModifyDNResponse, traits events.Traits) *ModifyDNLog {
	l := &ModifyDNLog{
		Request: &ModifyDNRequest{
			Operation:     "ModifyDN",
			Dn:            req.Dn,
			NewRdn:        req.NewRdn,
			DeleteOldDn:   req.DeleteOldDn,
			NewSuperiorDn: req.NewSuperiorDn,
		},
		Response: &Response{
			Status:    ldap.StatusText[res.ResultCode],
			MatchedDn: res.MatchedDn,
			Message:   res.Message,
		},
		Duration: 0,
		Actions:  nil,
	}
	_ = events.Push(l, traits.WithNamespace("ldap"))

	return l
}

type CompareLog struct {
	Request  *CompareRequest  `json:"request"`
	Response *Response        `json:"response"`
	Duration int64            `json:"duration"`
	Actions  []*common.Action `json:"actions"`
}

type CompareRequest struct {
	Operation string `json:"operation"`
	Dn        string `json:"dn"`
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
}

func NewCompareLogEvent(req *ldap.CompareRequest, res *ldap.CompareResponse, traits events.Traits) *CompareLog {
	l := &CompareLog{
		Request: &CompareRequest{
			Operation: "Compare",
			Dn:        req.Dn,
			Attribute: req.Attribute,
			Value:     req.Value,
		},
		Response: &Response{
			Status:  ldap.StatusText[res.ResultCode],
			Message: res.Message,
		},
		Duration: 0,
		Actions:  nil,
	}
	_ = events.Push(l, traits.WithNamespace("ldap"))

	return l
}
