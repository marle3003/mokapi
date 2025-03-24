package api

import (
	"mokapi/providers/openapi/parameter"
	"mokapi/providers/openapi/schema"
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
	Description string           `json:"description,omitempty"`
	Version     string           `json:"version,omitempty"`
	Contact     *contact         `json:"contact,omitempty"`
	Servers     []server         `json:"servers,omitempty"`
	Paths       []pathItem       `json:"paths,omitempty"`
	Metrics     []metrics.Metric `json:"metrics,omitempty"`
	Configs     []config         `json:"configs,omitempty"`
}

type pathItem struct {
	Path        string      `json:"path"`
	Summary     string      `json:"summary,omitempty"`
	Description string      `json:"description,omitempty"`
	Operations  []operation `json:"operations,omitempty"`
}

type operation struct {
	Method      string       `json:"method"`
	Summary     string       `json:"summary,omitempty"`
	Description string       `json:"description,omitempty"`
	OperationId string       `json:"operationId,omitempty"`
	Deprecated  bool         `json:"deprecated"`
	RequestBody *requestBody `json:"requestBody,omitempty"`
	Parameters  []param      `json:"parameters,omitempty"`
	Responses   []response   `json:"responses,omitempty"`
}

type param struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required"`
	Deprecated  bool           `json:"deprecated"`
	Style       string         `json:"style,omitempty"`
	Exploded    bool           `json:"exploded"`
	Schema      *schema.Schema `json:"schema"`
}

type response struct {
	StatusCode  string      `json:"statusCode"`
	Description string      `json:"description"`
	Contents    []mediaType `json:"contents,omitempty"`
	Headers     []header    `json:"headers,omitempty"`
}

type header struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Schema      *schema.Schema `json:"schema"`
}

type requestBody struct {
	Description string      `json:"description"`
	Contents    []mediaType `json:"contents,omitempty"`
	Required    bool        `json:"required"`
}

type mediaType struct {
	Type   string         `json:"type"`
	Schema *schema.Schema `json:"schema"`
}

type server struct {
	Url         string `json:"url"`
	Description string `json:"description"`
}

func getHttpServices(store *runtime.HttpStore, m *monitor.Monitor) []interface{} {
	list := store.List()
	result := make([]interface{}, 0, len(list))
	for _, hs := range list {
		s := service{
			Name:        hs.Info.Name,
			Description: hs.Info.Description,
			Version:     hs.Info.Version,
			Type:        ServiceHttp,
		}

		if m != nil {
			s.Metrics = m.FindAll(metrics.ByNamespace("http"), metrics.ByLabel("service", hs.Info.Name))
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

	s := h.app.Http.Get(name)
	if s == nil {
		w.WriteHeader(404)
		return
	}

	result := httpInfo{
		Name:        s.Info.Name,
		Description: s.Info.Description,
		Version:     s.Info.Version,
	}

	if m != nil {
		result.Metrics = m.FindAll(metrics.ByNamespace("http"), metrics.ByLabel("service", s.Info.Name))
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

	for path, p := range s.Paths {
		if p.Value == nil {
			continue
		}
		pi := pathItem{
			Path:        path,
			Summary:     p.Value.Summary,
			Description: p.Value.Description,
		}
		if len(p.Summary) > 0 {
			pi.Summary = p.Summary
		}
		if len(p.Description) > 0 {
			pi.Description = p.Description
		}

		for m, o := range p.Value.Operations() {
			op := operation{
				Method:      strings.ToLower(m),
				Summary:     o.Summary,
				Description: o.Description,
				OperationId: o.OperationId,
				Deprecated:  o.Deprecated,
			}
			if o.RequestBody != nil && o.RequestBody.Value != nil {
				op.RequestBody = &requestBody{
					Description: o.RequestBody.Value.Description,
					Required:    o.RequestBody.Value.Required,
				}
				if len(o.RequestBody.Summary) > 0 {
					op.Summary = o.RequestBody.Summary
				}
				if len(o.RequestBody.Description) > 0 {
					pi.Description = o.RequestBody.Description
				}

				for ct, rb := range o.RequestBody.Value.Content {
					op.RequestBody.Contents = append(op.RequestBody.Contents, mediaType{
						Type:   ct,
						Schema: rb.Schema,
					})
				}
			}
			op.Parameters = getParameters(p.Value.Parameters)
			op.Parameters = append(op.Parameters, getParameters(o.Parameters)...)

			for it := o.Responses.Iter(); it.Next(); {
				statusCode := it.Key()
				r := it.Value()
				if r.Value == nil {
					continue
				}
				res := response{
					StatusCode:  statusCode,
					Description: r.Value.Description,
				}
				if len(r.Description) > 0 {
					res.Description = r.Description
				}

				for ct, r := range r.Value.Content {
					res.Contents = append(res.Contents, mediaType{
						Type:   ct,
						Schema: r.Schema,
					})
				}
				for name, h := range r.Value.Headers {
					if h.Value == nil {
						continue
					}

					hi := header{
						Name:        name,
						Description: h.Value.Description,
						Schema:      h.Value.Schema,
					}
					if len(h.Description) > 0 {
						hi.Description = h.Description
					}

					res.Headers = append(res.Headers, hi)
				}

				op.Responses = append(op.Responses, res)
			}
			pi.Operations = append(pi.Operations, op)
		}
		result.Paths = append(result.Paths, pi)
	}

	result.Configs = getConfigs(s.Configs())

	w.Header().Set("Content-Type", "application/json")
	writeJsonBody(w, result)
}

func getParameters(params parameter.Parameters) (result []param) {
	for _, p := range params {
		if p.Value == nil {
			continue
		}

		pi := param{
			Name:        p.Value.Name,
			Type:        string(p.Value.Type),
			Description: p.Value.Description,
			Required:    p.Value.Required,
			Deprecated:  p.Value.Deprecated,
			Style:       p.Value.Style,
			Exploded:    p.Value.IsExplode(),
			Schema:      p.Value.Schema,
		}
		if len(p.Description) > 0 {
			pi.Description = p.Description
		}

		result = append(result, pi)
	}
	return
}
