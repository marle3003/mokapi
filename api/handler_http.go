package api

import (
	"fmt"
	"maps"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema"
	"mokapi/runtime"
	"mokapi/runtime/metrics"
	"mokapi/runtime/monitor"
	"net/http"
	"slices"
	"strings"
)

type httpInfo struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Version     string           `json:"version,omitempty"`
	Contact     *contact         `json:"contact,omitempty"`
	Servers     []server         `json:"servers,omitempty"`
	Paths       []pathItem       `json:"paths,omitempty"`
	Tags        []tag            `json:"tags,omitempty"`
	Metrics     []metrics.Metric `json:"metrics,omitempty"`
	Configs     []config         `json:"configs,omitempty"`
}

type pathItem struct {
	Path        string          `json:"path"`
	Summary     string          `json:"summary,omitempty"`
	Description string          `json:"description,omitempty"`
	Status      string          `json:"status"`
	Errors      []errorData     `json:"errors,omitempty"`
	Operations  []operationInfo `json:"operations,omitempty"`
}

type operationInfo struct {
	Method      string      `json:"method"`
	Summary     string      `json:"summary,omitempty"`
	Description string      `json:"description,omitempty"`
	OperationId string      `json:"operationId,omitempty"`
	Deprecated  bool        `json:"deprecated"`
	Tags        []string    `json:"tags,omitempty"`
	Status      string      `json:"status"`
	Errors      []errorData `json:"errors,omitempty"`
}

type operation struct {
	Method      string                `json:"method"`
	Path        string                `json:"path"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	OperationId string                `json:"operationId,omitempty"`
	Deprecated  bool                  `json:"deprecated"`
	RequestBody *requestBody          `json:"requestBody,omitempty"`
	Parameters  []param               `json:"parameters,omitempty"`
	Responses   []response            `json:"responses,omitempty"`
	Security    []securityRequirement `json:"security,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Status      string                `json:"status"`
	Errors      []errorData           `json:"errors,omitempty"`
}

type errorData struct {
	Message string `json:"message"`
}

type param struct {
	Name          string         `json:"name"`
	Type          string         `json:"type"`
	Description   string         `json:"description,omitempty"`
	Required      bool           `json:"required"`
	Deprecated    bool           `json:"deprecated"`
	Style         string         `json:"style,omitempty"`
	Explode       bool           `json:"explode"`
	AllowReserved bool           `json:"allowReserved"`
	Schema        *schema.Schema `json:"schema"`
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

type securityRequirement map[string]securityScheme

type securityScheme struct {
	Scopes  []string    `json:"scopes"`
	Configs interface{} `json:"configs"`
}

type tag struct {
	Name        string `yaml:"name" json:"name"`
	Summary     string `yaml:"summary" json:"summary"`
	Description string `yaml:"description" json:"description"`
	Parent      string `yaml:"parent" json:"parent"`
	Kind        string `yaml:"kind" json:"kind"`
}

func getHttpServices(s *runtime.HttpStore, m *monitor.Monitor) []service {
	list := s.List()
	result := make([]service, 0, len(list))
	for _, hs := range list {
		s := service{
			Name:        hs.Info.Name,
			Description: hs.Info.Description,
			Version:     hs.Info.Version,
			Type:        ServiceHttp,
			Status:      hs.GetStatus().String(),
		}

		if hs.Info.Summary != "" {
			s.Description = hs.Info.Summary
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

		result = append(result, s)
	}
	return result
}

func (h *handler) handleHttp(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	segments := strings.Split(path, "/")
	switch {
	case len(segments) == 4:
		name := segments[3]
		if s := h.app.Http.Get(name); s != nil {
			result := h.getHttpService(s)
			w.Header().Set("Content-Type", "application/json")
			writeJsonBody(w, result)
		} else {
			w.WriteHeader(404)
		}
	case len(segments) == 5 && segments[4] == "operations":
		name := segments[3]
		method := r.URL.Query().Get("method")
		p := r.URL.Query().Get("path")
		if s := h.app.Http.Get(name); s != nil {
			result := getOperations(s, p, method)
			w.Header().Set("Content-Type", "application/json")
			writeJsonBody(w, result)

		} else {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(fmt.Sprintf("Not found: %s", name)))
		}
	}
}

func (h *handler) getHttpService(s *runtime.HttpInfo) httpInfo {
	result := httpInfo{
		Name:        s.Info.Name,
		Description: s.Info.Description,
		Version:     s.Info.Version,
	}

	if h.app.Monitor != nil {
		result.Metrics = h.app.Monitor.FindAll(metrics.ByNamespace("http"), metrics.ByLabel("service", s.Info.Name))
	}

	if s.Info.Contact != nil {
		result.Contact = &contact{
			Name:  s.Info.Contact.Name,
			Url:   s.Info.Contact.Url,
			Email: s.Info.Contact.Email,
		}
	}

	for _, item := range s.Servers {
		result.Servers = append(result.Servers, server{
			Url:         item.Url,
			Description: item.Description,
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
			Status:      p.Value.Status.String(),
		}
		if len(p.Summary) > 0 {
			pi.Summary = p.Summary
		}
		if len(p.Description) > 0 {
			pi.Description = p.Description
		}

		for _, err := range p.Value.Errors {
			pi.Errors = append(pi.Errors, errorData{Message: err.Message})
		}

		for method, o := range p.Value.Operations() {
			pi.Operations = append(pi.Operations, operationInfo{
				Method:      strings.ToLower(method),
				Summary:     o.Summary,
				Description: o.Description,
				OperationId: o.OperationId,
				Deprecated:  o.Deprecated,
				Tags:        o.Tags,
				Status:      o.Status.String(),
				Errors:      getErrors(o.Errors),
			})
		}
		result.Paths = append(result.Paths, pi)
	}

	for _, t := range s.Tags {
		result.Tags = append(result.Tags, tag{
			Name:        t.Name,
			Summary:     t.Summary,
			Description: t.Description,
			Parent:      t.Parent,
			Kind:        t.Kind,
		})
	}

	result.Configs = getConfigs(s.Configs())

	return result
}

func getOperations(s *runtime.HttpInfo, path, method string) []operation {
	var paths []string
	if path != "" {
		paths = append(paths, path)
	} else {
		keys := maps.Keys(s.Paths)
		paths = slices.Sorted(keys)
	}

	operations := make([]operation, 0, len(paths))
	for _, ps := range paths {

		p, ok := s.Paths[ps]
		if !ok || p.Value == nil {
			continue
		}

		var methods []string
		if method != "" {
			methods = append(methods, method)
		} else {
			keys := maps.Keys(p.Value.Operations())
			methods = slices.Sorted(keys)
		}

		for _, m := range methods {
			o := p.Value.Operation(m)
			if o == nil {
				continue
			}

			op := operation{
				Method:      strings.ToLower(method),
				Path:        ps,
				Summary:     o.Summary,
				Description: o.Description,
				OperationId: o.OperationId,
				Deprecated:  o.Deprecated,
				Tags:        o.Tags,
				Status:      o.Status.String(),
			}
			if o.RequestBody != nil && o.RequestBody.Value != nil {
				op.RequestBody = &requestBody{
					Description: o.RequestBody.Value.Description,
					Required:    o.RequestBody.Value.Required,
				}
				if len(o.RequestBody.Summary) > 0 {
					op.Summary = o.RequestBody.Summary
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

			if o.Responses != nil {
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
			}

			requirements := append(o.Security, s.Security...)
			for _, sec := range requirements {
				req := securityRequirement{}
				for n, scopes := range sec {
					secConfig := s.Components.SecuritySchemes[n]
					scheme := securityScheme{
						Scopes:  scopes,
						Configs: secConfig,
					}
					req[n] = scheme
				}
				op.Security = append(op.Security, req)
			}
			if o.Errors != nil {
				for _, err := range o.Errors {
					op.Errors = append(op.Errors, errorData{Message: err.Message})
				}
			}

			operations = append(operations, op)
		}
	}
	return operations
}

func getParameters(params openapi.Parameters) (result []param) {
	for _, p := range params {
		if p.Value == nil {
			continue
		}

		pi := param{
			Name:          p.Value.Name,
			Type:          string(p.Value.Type),
			Description:   p.Value.Description,
			Required:      p.Value.Required,
			Deprecated:    p.Value.Deprecated,
			Style:         p.Value.Style,
			Explode:       p.Value.IsExplode(),
			AllowReserved: p.Value.AllowReserved,
			Schema:        p.Value.Schema,
		}
		if len(p.Description) > 0 {
			pi.Description = p.Description
		}

		result = append(result, pi)
	}
	return
}

func getErrors(err []openapi.Error) []errorData {
	var errData []errorData
	for _, e := range err {
		errData = append(errData, errorData{Message: e.Message})
	}
	return errData
}
