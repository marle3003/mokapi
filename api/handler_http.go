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
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"version"`
	Contact     *contact         `json:"contact"`
	Servers     []server         `json:"servers,omitempty"`
	Paths       []pathItem       `json:"paths,omitempty"`
	Metrics     []metrics.Metric `json:"metrics"`
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
	Headers     []parameter `json:"parameters,omitempty"`
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
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Ref         string        `json:"ref,omitempty"`
	Type        string        `json:"type"`
	Properties  []*schemaInfo `json:"properties,omitempty"`
	Enum        []interface{} `json:"enum,omitempty"`
	Items       *schemaInfo   `json:"items,omitempty"`
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
		s := service{
			Name:        hs.Info.Name,
			Description: hs.Info.Description,
			Version:     hs.Info.Version,
			Type:        ServiceHttp,
			Metrics:     m.FindAll(metrics.ByNamespace("http"), metrics.ByLabel("service", hs.Info.Name)),
		}
		if hs.Info.Contact != nil {
			c := hs.Info.Contact
			s.Contact = &contact{
				Name:  c.Name,
				Url:   c.Url,
				Email: c.Email,
			}
		}

		result = append(result, &httpSummary{service: s})
	}
	return result
}

func (h *handler) getHttpService(w http.ResponseWriter, r *http.Request, m *monitor.Monitor) {
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
		Metrics:     m.FindAll(metrics.ByNamespace("http"), metrics.ByLabel("service", s.Info.Name)),
	}
	if s.Info.Contact != nil {
		result.Contact = &contact{
			Name:  s.Info.Contact.Name,
			Url:   s.Info.Contact.Url,
			Email: s.Info.Contact.Email,
		}
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
				Method:      strings.ToLower(m),
				Summary:     o.Summary,
				Description: o.Description,
				OperationId: o.OperationId,
			}
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
				r := it.Value().(*openapi.ResponseRef)
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
				for name, header := range r.Value.Headers {
					if header.Value != nil {
						continue
					}
					res.Headers = append(res.Headers, parameter{
						Name:        name,
						Type:        "header",
						Description: header.Value.Description,
						Schema:      getSchema(header.Value.Schema),
					})
				}

				op.Responses = append(op.Responses, res)
			}
			pi.Operations = append(pi.Operations, op)
		}
		result.Paths = append(result.Paths, pi)
	}

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func getSchema(s *schema.Ref) *schemaInfo {
	if s == nil || s.Value == nil {
		return nil
	}

	result := &schemaInfo{
		Description: s.Value.Description,
		Ref:         s.Ref,
		Type:        s.Value.Type,
		Enum:        s.Value.Enum,
		Items:       getSchema(s.Value.Items),
	}

	if s.Value.Properties != nil && s.Value.Properties.Value != nil {
		for it := s.Value.Properties.Value.Iter(); it.Next(); {
			prop := getSchema(it.Value().(*schema.Ref))
			prop.Name = it.Key().(string)
			result.Properties = append(result.Properties, prop)
		}
	}

	return result
}
