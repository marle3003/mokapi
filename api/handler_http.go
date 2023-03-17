package api

import (
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/parameter"
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
	Metrics     []metrics.Metric `json:"metrics"`
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
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required"`
	Deprecated  bool        `json:"deprecated"`
	Style       string      `json:"style,omitempty"`
	Exploded    bool        `json:"exploded"`
	Schema      *schemaInfo `json:"schema"`
}

type response struct {
	StatusCode  int         `json:"statusCode"`
	Description string      `json:"description"`
	Contents    []mediaType `json:"contents,omitempty"`
	Headers     []param     `json:"parameters,omitempty"`
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
				Deprecated:  o.Deprecated,
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
			op.Parameters = getParameters(p.Value.Parameters)
			op.Parameters = append(op.Parameters, getParameters(o.Parameters)...)

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
					res.Headers = append(res.Headers, param{
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

func getParameters(params parameter.Parameters) (result []param) {
	for _, p := range params {
		if p.Value == nil {
			continue
		}
		result = append(result, param{
			Name:        p.Value.Name,
			Type:        string(p.Value.Type),
			Description: p.Value.Description,
			Required:    p.Value.Required,
			Deprecated:  p.Value.Deprecated,
			Style:       p.Value.Style,
			Exploded:    p.Value.Explode,
			Schema:      getSchema(p.Value.Schema),
		})
	}
	return
}
