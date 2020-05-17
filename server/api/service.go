package api

import (
	"encoding/json"
	"fmt"
	"mokapi/models"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type service struct {
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Version     string     `json:"version,omitempty"`
	BaseUrls    []baseUrl  `json:"baseUrls,omitempty"`
	Endpoints   []endpoint `json:"endpoints,omitempty"`
}

type baseUrl struct {
	Url         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

type endpoint struct {
	Path       string      `json:"path,omitempty"`
	Operations []operation `json:"operations,omitempty"`
}

type operation struct {
	Method      string      `json:"method,omitempty"`
	Summary     string      `json:"summary,omitempty"`
	Description string      `json:"description,omitempty"`
	Parameters  []parameter `json:"parameters,omitempty"`
	Responses   []response  `json:"responses,omitempty"`
}

type parameter struct {
	Name        string `json:"name,omitempty"`
	In          string `json:"in,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

type response struct {
	Status       int               `json:"status,omitempty"`
	Description  string            `json:"description,omitempty"`
	ContentTypes []responseContent `json:"contentTypes,omitempty"`
}

type responseContent struct {
	Type   string  `json:"type,omitempty"`
	Schema *schema `json:"schema,omitempty"`
}

type schema struct {
	Name       string    `json:"name,omitempty"`
	Type       string    `json:"type,omitempty"`
	Properties []*schema `json:"properties,omitempty"`
}

func newService(s *models.ServiceInfo) service {
	service := service{Name: s.Service.Name, Description: s.Service.Description, Version: s.Service.Version, BaseUrls: make([]baseUrl, 0)}

	for _, server := range s.Service.Servers {
		service.BaseUrls = append(service.BaseUrls, newBaseUrl(server))
	}

	for _, e := range s.Service.Endpoint {
		service.Endpoints = append(service.Endpoints, newEndpoint(e))
	}

	return service
}

func newBaseUrl(s models.Server) baseUrl {
	return baseUrl{Url: fmt.Sprintf("http://%v:%v%v", s.Host, s.Port, s.Path), Description: s.Description}
}

func newEndpoint(e *models.Endpoint) endpoint {
	v := endpoint{Path: e.Path, Operations: make([]operation, 0)}
	if e.Get != nil {
		v.Operations = append(v.Operations, newOperation("get", e.Get))
	} else if e.Post != nil {
		v.Operations = append(v.Operations, newOperation("post", e.Post))
	} else if e.Put != nil {
		v.Operations = append(v.Operations, newOperation("put", e.Put))
	} else if e.Patch != nil {
		v.Operations = append(v.Operations, newOperation("patch", e.Patch))
	} else if e.Delete != nil {
		v.Operations = append(v.Operations, newOperation("delete", e.Delete))
	} else if e.Head != nil {
		v.Operations = append(v.Operations, newOperation("head", e.Head))
	} else if e.Options != nil {
		v.Operations = append(v.Operations, newOperation("options", e.Options))
	} else if e.Trace != nil {
		v.Operations = append(v.Operations, newOperation("trace", e.Trace))
	}
	return v
}

func newOperation(method string, o *models.Operation) operation {
	v := operation{Method: method, Summary: o.Summary, Description: o.Description, Parameters: make([]parameter, 0), Responses: make([]response, 0)}

	for _, p := range o.Parameters {
		v.Parameters = append(v.Parameters, newParameter(p))
	}

	for s, r := range o.Responses {
		v.Responses = append(v.Responses, newResponse(s, r))
	}

	return v
}

func newParameter(p *models.Parameter) parameter {
	return parameter{In: p.Type.String(), Name: p.Name, Type: p.Schema.Type, Description: p.Description}
}

func newResponse(status models.HttpStatus, r *models.Response) response {
	response := response{Status: int(status), Description: r.Description, ContentTypes: make([]responseContent, 0)}

	for t, c := range r.ContentTypes {
		response.ContentTypes = append(response.ContentTypes, responseContent{Type: t.String(), Schema: newSchema("", c.Schema)})
	}

	return response
}

func newSchema(name string, s *models.Schema) *schema {
	if s == nil {
		return nil
	}

	v := &schema{Name: name, Type: s.Type, Properties: make([]*schema, 0)}

	for s, p := range s.Properties {
		v.Properties = append(v.Properties, newSchema(s, p))
	}

	return v
}

func (h Handler) getService(rw http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	rw.Header().Set("Access-Control-Allow-Origin", "*")

	if s, ok := h.application.Services[vars["name"]]; ok {
		service := newService(s)

		rw.Header().Set("Content-Type", "application/json")

		error := json.NewEncoder(rw).Encode(service)
		if error != nil {
			log.Errorf("Error in writing service response: %v", error.Error())
		}
	} else {
		rw.WriteHeader(404)
	}

}
