package api

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"net/http"
	"strings"
)

type httpSummary struct {
	service
}

type httpInfo struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Version     string     `json:"version"`
	Servers     []server   `json:"servers,omitempty"`
	Paths       []pathItem `json:"paths,omitempty"`
}

type pathItem struct {
	Path        string      `json:"path"`
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	Operations  []operation `json:"operations,omitempty"`
}

type operation struct {
	Method      string       `json:"method"`
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	OperationId string       `json:"operationId"`
	RequestBody *requestBody `json:"requestBody,omitempty"`
	Parameters  []parameter  `json:"parameters,omitempty"`
	Responses   []response   `json:"responses,omitempty"`
}

type parameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Style       string      `json:"style"`
	Exploded    bool        `json:"exploded"`
	Schema      *schemaInfo `json:"schema"`
}

type response struct {
	StatusCode  int         `json:"statusCode"`
	Description string      `json:"description"`
	Contents    []mediaType `json:"contents,omitempty"`
}

type requestBody struct {
	Description string      `json:"description"`
	Contents    []mediaType `json:"contents,omitempty"`
	Required    bool        `json:"required"`
}

type mediaType struct {
	Type   string      `json:"type"`
	Schema *schemaInfo `json:"schema"`
}

type schemaInfo struct {
	Ref        string        `json:"ref"`
	Type       string        `json:"type"`
	Properties []property    `json:"properties,omitempty"`
	Enum       []interface{} `json:"enum,omitempty"`
}

type property struct {
	Name  string      `json:"name"`
	Value *schemaInfo `json:"value"`
}

type server struct {
	Url         string `json:"url"`
	Description string `json:"description"`
}

func (h *handler) getHttpServices(w http.ResponseWriter, _ *http.Request) {
	result := getHttpServices(h.app.Http, h.app.Monitor)
	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func getHttpServices(services map[string]*runtime.HttpInfo, m *monitor.Monitor) []interface{} {
	result := make([]interface{}, 0, len(services))
	for _, hs := range services {

		result = append(result, &httpSummary{
			service: service{
				Name:        hs.Info.Name,
				Description: hs.Info.Description,
				Version:     hs.Info.Version,
				Type:        ServiceHttp,
				Metrics:     m.FindAll(metrics.ByNamespace("http"), metrics.ByLabel("service", hs.Info.Name)),
			},
		})
	}
	return result
}

func (h *handler) getHttpService(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	name := segments[4]

	s, ok := h.app.Http[name]
	if !ok {
		w.WriteHeader(404)
		return
	}

	result := httpInfo{
		Name:        s.Info.Name,
		Description: s.Info.Description,
		Version:     s.Info.Version,
	}

	for _, s := range s.Servers {
		result.Servers = append(result.Servers, server{
			Url:         s.Url,
			Description: s.Description,
		})
	}

	for path, p := range s.Paths.Value {
		if p.Value == nil {
			continue
		}
		pi := pathItem{
			Path:        path,
			Summary:     p.Value.Summary,
			Description: p.Value.Description,
		}

		for m, o := range p.Value.Operations() {
			op := operation{
				Method:      m,
				Summary:     o.Summary,
				Description: o.Description,
				OperationId: o.OperationId,
			}
			pi.Operations = append(pi.Operations, op)
			if o.RequestBody != nil && o.RequestBody.Value != nil {
				op.RequestBody = &requestBody{
					Description: o.RequestBody.Value.Description,
					Required:    o.RequestBody.Value.Required,
				}
				for ct, rb := range o.RequestBody.Value.Content {
					op.RequestBody.Contents = append(op.RequestBody.Contents, mediaType{
						Type:   ct,
						Schema: getSchema(rb.Schema),
					})
				}
			}
			for _, p := range o.Parameters {
				if p.Value == nil {
					continue
				}
				op.Parameters = append(op.Parameters, parameter{
					Name:        p.Value.Name,
					Type:        string(p.Value.Type),
					Description: p.Value.Description,
					Required:    p.Value.Required,
					Style:       p.Value.Style,
					Exploded:    p.Value.Explode,
					Schema:      getSchema(p.Value.Schema),
				})
			}

			for it := o.Responses.Iter(); it.Next(); {
				statusCode := it.Key().(int)
				r := it.Value().(openapi.ResponseRef)
				if r.Value == nil {
					continue
				}
				res := response{
					StatusCode:  statusCode,
					Description: r.Value.Description,
				}
				for ct, r := range r.Value.Content {
					res.Contents = append(res.Contents, mediaType{
						Type:   ct,
						Schema: getSchema(r.Schema),
					})
				}

				op.Responses = append(op.Responses, res)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func getSchema(s *schema.Ref) *schemaInfo {
	if s.Value == nil {
		return nil
	}

	result := &schemaInfo{
		Ref:  s.Ref,
		Type: s.Value.Type,
		Enum: s.Value.Enum,
	}

	if s.Value.Properties != nil && s.Value.Properties.Value != nil {
		for it := s.Value.Properties.Value.Iter(); it.Next(); {
			result.Properties = append(result.Properties, property{
				Name:  it.Key().(string),
				Value: nil,
			})
		}
	}

	return result
}
