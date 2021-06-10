package openapi

import (
	"fmt"
	"mokapi/config/dynamic/openapi"
	"sort"
	"strings"
)

type Service struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Version     string     `json:"version"`
	BaseUrls    []baseUrl  `json:"baseUrls"`
	Endpoints   []Endpoint `json:"endpoints"`
	Models      []*Schema  `json:"models"`
	Mokapi      string     `json:"mokapifile"`
	Type        string     `json:"type"`
}

type baseUrl struct {
	Url         string `json:"url"`
	Description string `json:"description"`
}

type Endpoint struct {
	Path        string      `json:"path"`
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	Operations  []Operation `json:"operations"`
}

type Operation struct {
	Method        string        `json:"method"`
	Summary       string        `json:"summary"`
	Description   string        `json:"description"`
	Parameters    []Parameter   `json:"parameters"`
	Responses     []Response    `json:"responses"`
	Pipeline      string        `json:"pipeline"`
	RequestBodies []RequestBody `json:"requestBodies"`
}

type RequestBody struct {
	Description  string  `json:"description"`
	ContentTypes string  `json:"contentType"`
	Schema       *Schema `json:"schema"`
	Required     bool    `json:"required"`
}

type Parameter struct {
	Name        string  `json:"name"`
	Location    string  `json:"location"`
	Schema      *Schema `json:"schema"`
	Description string  `json:"description"`
	Required    bool    `json:"required"`
	Style       string  `json:"style"`
	Explode     bool    `json:"explode"`
}

type Response struct {
	Status       int               `json:"status"`
	Description  string            `json:"description"`
	ContentTypes []ResponseContent `json:"contentTypes"`
}

type ResponseContent struct {
	Type   string  `json:"type"`
	Schema *Schema `json:"schema"`
}

type Schema struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Properties  []*Schema `json:"properties"`
	Items       *Schema   `json:"items"`
	Ref         string    `json:"ref"`
	Description string    `json:"description"`
	Required    []string  `json:"required"`
	Format      string    `json:"format"`
	Faker       string    `json:"faker"`
	Nullable    bool      `json:"nullable"`
}

func NewService(s *openapi.Config) Service {
	service := Service{
		Name:        s.Info.Name,
		Description: s.Info.Description,
		Version:     s.Info.Version,
		Type:        "OpenAPI",
	}

	for _, server := range s.Servers {
		service.BaseUrls = append(service.BaseUrls, newBaseUrl(server))
	}

	for path, e := range s.EndPoints {
		if e.Value != nil {
			service.Endpoints = append(service.Endpoints, newEndpoint(path, e.Value))
		}
	}

	if s.Components.Schemas != nil {
		for name, model := range s.Components.Schemas.Value {
			service.Models = append(service.Models, newModel(name, model))
		}
	}

	return service
}

func newBaseUrl(s *openapi.Server) baseUrl {
	return baseUrl{Url: s.Url, Description: s.Description}
}

func newEndpoint(path string, e *openapi.Endpoint) Endpoint {
	v := Endpoint{Path: path, Summary: e.Summary, Description: e.Description, Operations: make([]Operation, 0)}
	if e.Get != nil {
		v.Operations = append(v.Operations, newOperation("get", e.Get, e.Pipeline))
	}
	if e.Post != nil {
		v.Operations = append(v.Operations, newOperation("post", e.Post, e.Pipeline))
	}
	if e.Put != nil {
		v.Operations = append(v.Operations, newOperation("put", e.Put, e.Pipeline))
	}
	if e.Patch != nil {
		v.Operations = append(v.Operations, newOperation("patch", e.Patch, e.Pipeline))
	}
	if e.Delete != nil {
		v.Operations = append(v.Operations, newOperation("delete", e.Delete, e.Pipeline))
	}
	if e.Head != nil {
		v.Operations = append(v.Operations, newOperation("head", e.Head, e.Pipeline))
	}
	if e.Options != nil {
		v.Operations = append(v.Operations, newOperation("options", e.Options, e.Pipeline))
	}
	if e.Trace != nil {
		v.Operations = append(v.Operations, newOperation("trace", e.Trace, e.Pipeline))
	}
	return v
}

func newOperation(method string, o *openapi.Operation, pipeline string) Operation {
	v := Operation{
		Method:      method,
		Summary:     o.Summary,
		Description: o.Description,
		Parameters:  make([]Parameter, 0),
		Responses:   make([]Response, 0),
		Pipeline:    o.Pipeline,
	}

	if len(v.Pipeline) == 0 {
		v.Pipeline = pipeline
	}

	for _, p := range o.Parameters {
		if p.Value != nil {
			v.Parameters = append(v.Parameters, newParameter(p.Value))
		}
	}

	if o.RequestBody != nil {
		for c, r := range o.RequestBody.Value.Content {
			v.RequestBodies = append(v.RequestBodies,
				RequestBody{
					Description:  o.RequestBody.Value.Description,
					Required:     o.RequestBody.Value.Required,
					ContentTypes: c,
					Schema:       newSchema("", r.Schema, 0)})
		}
	}

	for s, r := range o.Responses {
		v.Responses = append(v.Responses, newResponse(s, r))
	}

	return v
}

func newParameter(p *openapi.Parameter) Parameter {
	return Parameter{
		Location:    string(p.Type),
		Name:        p.Name,
		Schema:      newSchema("", p.Schema, 0),
		Description: p.Description,
		Required:    p.Required,
		Style:       p.Style,
		Explode:     p.Explode,
	}
}

func newResponse(status openapi.HttpStatus, r *openapi.ResponseRef) Response {
	response := Response{Status: int(status), Description: r.Value.Description, ContentTypes: make([]ResponseContent, 0)}

	for t, c := range r.Value.Content {
		response.ContentTypes = append(response.ContentTypes, ResponseContent{Type: t, Schema: newSchema("", c.Schema, 0)})
	}

	return response
}

func newSchema(name string, s *openapi.SchemaRef, level int) *Schema {
	if s == nil {
		return nil
	}

	v := &Schema{
		Name:        name,
		Type:        s.Value.Type,
		Properties:  make([]*Schema, 0),
		Ref:         s.Ref,
		Description: s.Value.Description,
		Required:    s.Value.Required,
		Format:      s.Value.Format,
		Faker:       s.Value.Faker,
		Nullable:    s.Value.Nullable,
	}

	if s.Value.Items != nil {
		v.Items = newSchema("", s.Value.Items, level+1)
	}

	if level > 10 {
		return v
	}

	if s.Value.Properties != nil {
		for s, p := range s.Value.Properties.Value {
			v.Properties = append(v.Properties, newSchema(s, p, level+1))
		}
	}

	sort.Slice(v.Properties, func(i int, j int) bool {
		return strings.Compare(v.Properties[i].Name, v.Properties[j].Name) < 0
	})

	return v
}

func newModel(name string, s *openapi.SchemaRef) *Schema {
	if s == nil {
		return nil
	}

	v := &Schema{Name: name, Type: s.Value.Type, Properties: make([]*Schema, 0), Ref: s.Ref}

	for s, p := range s.Value.Properties.Value {
		if p.Value.Type == "array" && p.Value.Items != nil {
			tName := p.Value.Items.Value.Type
			if len(p.Value.Items.Ref) > 0 {
				seg := strings.Split(p.Value.Items.Ref, "/")
				tName = seg[len(seg)-1]
			}
			v.Properties = append(v.Properties, &Schema{Name: s, Type: fmt.Sprintf("array[%v]", tName)})
		} else if len(p.Ref) > 0 {
			seg := strings.Split(p.Ref, "/")
			tName := seg[len(seg)-1]
			v.Properties = append(v.Properties, &Schema{Name: s, Type: tName})
		} else {
			v.Properties = append(v.Properties, newSchema(s, p, 0))
		}
	}

	if s.Value.Type == "array" && s.Value.Items != nil {
		tName := s.Value.Items.Value.Type
		if len(s.Value.Items.Ref) > 0 {
			seg := strings.Split(s.Value.Items.Ref, "/")
			tName = seg[len(seg)-1]
		}
		v.Type = fmt.Sprintf("array[%v]", tName)
	}

	return v
}
