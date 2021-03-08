package api

import (
	"fmt"
	"mokapi/config/dynamic/openapi"
	"sort"
	"strings"
)

type service struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Version     string     `json:"version"`
	BaseUrls    []baseUrl  `json:"baseUrls"`
	Endpoints   []endpoint `json:"endpoints"`
	Models      []*schema  `json:"models"`
	MokapiFile  string     `json:"mokapifile"`
}

type baseUrl struct {
	Url         string `json:"url"`
	Description string `json:"description"`
}

type endpoint struct {
	Path        string      `json:"path"`
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	Operations  []operation `json:"operations"`
}

type operation struct {
	Method        string        `json:"method"`
	Summary       string        `json:"summary"`
	Description   string        `json:"description"`
	Parameters    []parameter   `json:"parameters"`
	Responses     []response    `json:"responses"`
	Pipeline      string        `json:"pipeline"`
	RequestBodies []requestBody `json:"requestBodies"`
}

type requestBody struct {
	Description  string  `json:"description"`
	ContentTypes string  `json:"contentType"`
	Schema       *schema `json:"schema"`
	Required     bool    `json:"required"`
}

type parameter struct {
	Name        string  `json:"name"`
	Location    string  `json:"location"`
	Schema      *schema `json:"schema"`
	Description string  `json:"description"`
	Required    bool    `json:"required"`
	Style       string  `json:"style"`
	Explode     bool    `json:"explode"`
}

type response struct {
	Status       int               `json:"status"`
	Description  string            `json:"description"`
	ContentTypes []responseContent `json:"contentTypes"`
}

type responseContent struct {
	Type   string  `json:"type"`
	Schema *schema `json:"schema"`
}

type schema struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Properties  []*schema `json:"properties"`
	Items       *schema   `json:"items"`
	Ref         string    `json:"ref"`
	Description string    `json:"description"`
	Required    []string  `json:"required"`
	Format      string    `json:"format"`
	Faker       string    `json:"faker"`
	Nullable    bool      `json:"nullable"`
}

func newService(s *openapi.Config) service {
	service := service{
		Name:        s.Info.Name,
		Description: s.Info.Description,
		Version:     s.Info.Version,
		MokapiFile:  s.Info.Mokapi.ConfigPath,
	}

	for _, server := range s.Servers {
		service.BaseUrls = append(service.BaseUrls, newBaseUrl(server))
	}

	for path, e := range s.EndPoints {
		service.Endpoints = append(service.Endpoints, newEndpoint(path, e))
	}

	//for name, model := range s.Components.Schemas {
	//	service.Models = append(service.Models, newModel(name, model))
	//}

	return service
}

func newBaseUrl(s *openapi.Server) baseUrl {
	return baseUrl{Url: s.Url, Description: s.Description}
}

func newEndpoint(path string, e *openapi.Endpoint) endpoint {
	v := endpoint{Path: path, Summary: e.Summary, Description: e.Description, Operations: make([]operation, 0)}
	if e.Get != nil {
		v.Operations = append(v.Operations, newOperation("get", e.Get, e.Pipeline))
	} else if e.Post != nil {
		v.Operations = append(v.Operations, newOperation("post", e.Post, e.Pipeline))
	} else if e.Put != nil {
		v.Operations = append(v.Operations, newOperation("put", e.Put, e.Pipeline))
	} else if e.Patch != nil {
		v.Operations = append(v.Operations, newOperation("patch", e.Patch, e.Pipeline))
	} else if e.Delete != nil {
		v.Operations = append(v.Operations, newOperation("delete", e.Delete, e.Pipeline))
	} else if e.Head != nil {
		v.Operations = append(v.Operations, newOperation("head", e.Head, e.Pipeline))
	} else if e.Options != nil {
		v.Operations = append(v.Operations, newOperation("options", e.Options, e.Pipeline))
	} else if e.Trace != nil {
		v.Operations = append(v.Operations, newOperation("trace", e.Trace, e.Pipeline))
	}
	return v
}

func newOperation(method string, o *openapi.Operation, pipeline string) operation {
	v := operation{
		Method:      method,
		Summary:     o.Summary,
		Description: o.Description,
		Parameters:  make([]parameter, 0),
		Responses:   make([]response, 0),
		Pipeline:    o.Pipeline,
	}

	if len(v.Pipeline) == 0 {
		v.Pipeline = pipeline
	}

	for _, p := range o.Parameters {
		v.Parameters = append(v.Parameters, newParameter(p))
	}

	if o.RequestBody != nil {
		for c, r := range o.RequestBody.Content {
			v.RequestBodies = append(v.RequestBodies,
				requestBody{
					Description:  o.RequestBody.Description,
					Required:     o.RequestBody.Required,
					ContentTypes: c,
					Schema:       newSchema("", r.Schema, 0)})
		}
	}

	for s, r := range o.Responses {
		v.Responses = append(v.Responses, newResponse(s, r))
	}

	return v
}

func newParameter(p *openapi.Parameter) parameter {
	return parameter{
		Location:    string(p.Type),
		Name:        p.Name,
		Schema:      newSchema("", p.Schema, 0),
		Description: p.Description,
		Required:    p.Required,
		Style:       p.Style,
		Explode:     p.Explode,
	}
}

func newResponse(status openapi.HttpStatus, r *openapi.Response) response {
	response := response{Status: int(status), Description: r.Description, ContentTypes: make([]responseContent, 0)}

	for t, c := range r.Content {
		response.ContentTypes = append(response.ContentTypes, responseContent{Type: t, Schema: newSchema("", c.Schema, 0)})
	}

	return response
}

func newSchema(name string, s *openapi.Schema, level int) *schema {
	if s == nil {
		return nil
	}

	v := &schema{
		Name:        name,
		Type:        s.Type,
		Properties:  make([]*schema, 0),
		Ref:         s.Reference,
		Description: s.Description,
		Required:    s.Required,
		Format:      s.Format,
		Faker:       s.Faker,
		Nullable:    s.Nullable,
	}

	if s.Items != nil {
		v.Items = newSchema("", s.Items, level+1)
	}

	if level > 10 {
		return v
	}

	for s, p := range s.Properties {
		v.Properties = append(v.Properties, newSchema(s, p, level+1))
	}

	sort.Slice(v.Properties, func(i int, j int) bool {
		return strings.Compare(v.Properties[i].Name, v.Properties[j].Name) < 0
	})

	return v
}

func newModel(name string, s *openapi.Schema) *schema {
	if s == nil {
		return nil
	}

	v := &schema{Name: name, Type: s.Type, Properties: make([]*schema, 0), Ref: s.Reference}

	for s, p := range s.Properties {
		if p.Type == "array" && p.Items != nil {
			tName := p.Items.Type
			if len(p.Items.Reference) > 0 {
				seg := strings.Split(p.Items.Reference, "/")
				tName = seg[len(seg)-1]
			}
			v.Properties = append(v.Properties, &schema{Name: s, Type: fmt.Sprintf("array[%v]", tName)})
		} else if len(p.Reference) > 0 {
			seg := strings.Split(p.Reference, "/")
			tName := seg[len(seg)-1]
			v.Properties = append(v.Properties, &schema{Name: s, Type: tName})
		} else {
			v.Properties = append(v.Properties, newSchema(s, p, 0))
		}
	}

	if s.Type == "array" && s.Items != nil {
		tName := s.Items.Type
		if len(s.Items.Reference) > 0 {
			seg := strings.Split(s.Items.Reference, "/")
			tName = seg[len(seg)-1]
		}
		v.Type = fmt.Sprintf("array[%v]", tName)
	}

	return v
}

func newServiceSummary(s *openapi.Config) serviceSummary {
	return serviceSummary{Name: s.Info.Name, Description: s.Info.Description, Version: s.Info.Version}
}
