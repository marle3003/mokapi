package directory

import (
	"encoding/json"
	"fmt"
	"mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/runtime/events"
	"slices"
	"sync"
)

type LogInfo struct {
	Operation string `json:"operation"`
	Dn        string `json:"dn,omitempty"`
	Criteria  string `json:"criteria,omitempty"`
	Status    string `json:"status,omitempty"`
	Duration  int64  `json:"duration"`
}

// SearchLog is pushed to the event store before the search has finished. The response is
// filled while searching, so it is only changed through AddResult and SetStatus.
type SearchLog struct {
	common.EventLog
	Request  *SearchRequest
	Response *SearchResponse
	// m guards Response
	m sync.Mutex
}

func (l *SearchLog) AddResult(result SearchResult) {
	l.m.Lock()
	defer l.m.Unlock()
	l.Response.Results = append(l.Response.Results, result)
}

func (l *SearchLog) SetStatus(status string) {
	l.m.Lock()
	defer l.m.Unlock()
	l.Response.Status = status
}

type SearchRequest struct {
	Operation         string   `json:"operation"`
	BaseDN            string   `json:"baseDN"`
	Scope             string   `json:"scope"`
	DereferencePolicy int64    `json:"dereferencePolicy"`
	SizeLimit         int64    `json:"sizeLimit"`
	TimeLimit         int64    `json:"timeLimit"`
	TypesOnly         bool     `json:"typesOnly"`
	Filter            string   `json:"filter"`
	Attributes        []string `json:"attributes"`
	Controls          []any    `json:"controls"`
}

type SearchResponse struct {
	Status  string         `json:"status"`
	Results []SearchResult `json:"results"`
}

type SearchResult struct {
	Dn         string              `json:"dn"`
	Attributes map[string][]string `json:"attributes"`
}

func NewSearchLogEvent(r *ldap.SearchRequest, eh events.Handler, traits events.Traits) *SearchLog {
	event := &SearchLog{
		Request:  fromRequest(r),
		Response: &SearchResponse{},
	}
	_ = eh.Push(event, traits.WithNamespace("ldap").With("operation", "search"))
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

type BindLog struct {
	common.EventLog
	Request  *BindRequest
	Response *Response
}

type UnbindLog struct {
	common.EventLog
	Request *UnbindRequest
}

type UnbindRequest struct {
	Operation string `json:"operation"`
}

type BindRequest struct {
	Operation string `json:"operation"`
	Version   int64  `json:"version"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	Auth      string `json:"auth"`
}

type AddLog struct {
	common.EventLog
	Request  *AddRequest
	Response *Response
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
	MatchedDn string `json:"matchedDn,omitempty"`
	Message   string `json:"message"`
}

func NewBindLogEvent(req *ldap.BindRequest, res *ldap.BindResponse, eh events.Handler, traits events.Traits) *BindLog {
	l := &BindLog{
		Request: &BindRequest{
			Operation: "Bind",
			Version:   req.Version,
			Name:      req.Name,
			Password:  req.Password,
			Auth:      "Simple",
		},
		Response: &Response{
			Status:  ldap.StatusText[res.Result],
			Message: res.Message,
		},
	}
	_ = eh.Push(l, traits.WithNamespace("ldap").With("operation", "bind"))
	return l
}

func NewUnbindEvent(eh events.Handler, traits events.Traits) *UnbindLog {
	l := &UnbindLog{Request: &UnbindRequest{Operation: "Unbind"}}
	_ = eh.Push(l, traits.WithNamespace("ldap").With("operation", "unbind"))
	return l
}

func NewAddLogEvent(req *ldap.AddRequest, res *ldap.AddResponse, eh events.Handler, traits events.Traits) *AddLog {
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
	}
	_ = eh.Push(l, traits.WithNamespace("ldap").With("operation", "add"))

	return l
}

type ModifyLog struct {
	common.EventLog
	Request  *ModifyRequest
	Response *Response
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

func NewModifyLogEvent(req *ldap.ModifyRequest, res *ldap.ModifyResponse, eh events.Handler, traits events.Traits) *ModifyLog {
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
	}
	_ = eh.Push(l, traits.WithNamespace("ldap").With("operation", "modify"))

	return l
}

type DeleteLog struct {
	common.EventLog
	Request  *DeleteRequest
	Response *Response
}

type DeleteRequest struct {
	Operation string `json:"operation"`
	Dn        string `json:"dn"`
}

func NewDeleteLogEvent(req *ldap.DeleteRequest, res *ldap.DeleteResponse, eh events.Handler, traits events.Traits) *DeleteLog {
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
	}
	_ = eh.Push(l, traits.WithNamespace("ldap").With("operation", "delete"))

	return l
}

type ModifyDNLog struct {
	common.EventLog
	Request  *ModifyDNRequest
	Response *Response
}

type ModifyDNRequest struct {
	Operation     string `json:"operation"`
	Dn            string `json:"dn"`
	NewRdn        string `json:"newRdn"`
	DeleteOldDn   bool   `json:"deleteOldDn"`
	NewSuperiorDn string `json:"newSuperiorDn"`
}

func NewModifyDNLogEvent(req *ldap.ModifyDNRequest, res *ldap.ModifyDNResponse, eh events.Handler, traits events.Traits) *ModifyDNLog {
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
	}
	_ = eh.Push(l, traits.WithNamespace("ldap").With("operation", "modifydn"))

	return l
}

type CompareLog struct {
	common.EventLog
	Request  *CompareRequest
	Response *Response
}

type CompareRequest struct {
	Operation string `json:"operation"`
	Dn        string `json:"dn"`
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
}

func NewCompareLogEvent(req *ldap.CompareRequest, res *ldap.CompareResponse, eh events.Handler, traits events.Traits) *CompareLog {
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
	}
	_ = eh.Push(l, traits.WithNamespace("ldap").With("operation", "compare"))

	return l

}

func (l *BindLog) Title() string {
	return fmt.Sprintf("%s %s %s", l.Request.Auth, l.Request.Name, l.Request.Password)
}

func (l *UnbindLog) Title() string {
	return ""
}

func (l *SearchLog) Title() string {
	return l.Request.Filter
}

// IndexFields returns a copy of the response, because the index reads it after the
// lock is released while results may still be added.
func (l *SearchLog) IndexFields() map[string]any {
	l.m.Lock()
	defer l.m.Unlock()
	response := &SearchResponse{Status: l.Response.Status, Results: slices.Clone(l.Response.Results)}
	m := map[string]any{
		"operation":       l.Request.Operation,
		"request":         l.Request,
		"response":        response,
		"metadata.baseDN": l.Request.BaseDN,
	}
	return m
}

func (l *CompareLog) Title() string {
	return fmt.Sprintf("%s %s", l.Request.Operation, l.Request.Value)
}

func (l *ModifyLog) Title() string {
	return fmt.Sprintf("%s %s", l.Request.Operation, l.Request.Dn)
}

type ModifyInfo struct {
	Operation string   `json:"operation"`
	Dn        string   `json:"dn,omitempty"`
	Criteria  []string `json:"criteria,omitempty"`
	Status    string   `json:"status,omitempty"`
	Duration  int64    `json:"duration"`
}

func (l *ModifyDNLog) Title() string {
	return fmt.Sprintf("%s %s", l.Request.Operation, l.Request.Dn)
}

func (l *DeleteLog) Title() string {
	return fmt.Sprintf("%s %s", l.Request.Operation, l.Request.Dn)
}

func (l *AddLog) Title() string {
	return fmt.Sprintf("%s %s", l.Request.Operation, l.Request.Dn)
}

func (l *SearchLog) MarshalJSON() ([]byte, error) {
	l.m.Lock()
	defer l.m.Unlock()
	return marshalLog(&l.EventLog, l.Request, l.Response)
}

func (l *BindLog) MarshalJSON() ([]byte, error) {
	return marshalLog(&l.EventLog, l.Request, l.Response)
}

func (l *UnbindLog) MarshalJSON() ([]byte, error) {
	return marshalLog(&l.EventLog, l.Request, nil)
}

func (l *AddLog) MarshalJSON() ([]byte, error) {
	return marshalLog(&l.EventLog, l.Request, l.Response)
}

func (l *ModifyLog) MarshalJSON() ([]byte, error) {
	return marshalLog(&l.EventLog, l.Request, l.Response)
}

func (l *DeleteLog) MarshalJSON() ([]byte, error) {
	return marshalLog(&l.EventLog, l.Request, l.Response)
}

func (l *ModifyDNLog) MarshalJSON() ([]byte, error) {
	return marshalLog(&l.EventLog, l.Request, l.Response)
}

func (l *CompareLog) MarshalJSON() ([]byte, error) {
	return marshalLog(&l.EventLog, l.Request, l.Response)
}

// marshalLog is the JSON shape of all LDAP logs. The response is left out when it is nil
// (unbind has none).
func marshalLog(e *common.EventLog, request, response any) ([]byte, error) {
	return json.Marshal(struct {
		Request  any              `json:"request"`
		Response any              `json:"response,omitempty"`
		Duration int64            `json:"duration"`
		Actions  []*common.Action `json:"actions"`
	}{request, response, e.GetDuration(), e.GetActions()})
}
