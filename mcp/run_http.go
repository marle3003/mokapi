package mcp

import (
	"fmt"
	"io"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema"
	"mokapi/runtime"
	"mokapi/schema/json/generator"
	"net/http"
	"net/textproto"
	"slices"
	"strconv"
	"strings"
)

type OpenAPI struct {
	Name    string          `json:"name"`
	Type    string          `json:"type"`
	Servers []OpenAPIServer `json:"servers"`

	info    *runtime.HttpInfo
	handler openapi.Handler
}

type OpenAPIServer struct {
	Url         string `json:"url"`
	Description string `json:"description"`
}

type OperationSummary struct {
	Id         string   `json:"id"`
	Method     string   `json:"method"`
	Path       string   `json:"path"`
	Summary    string   `json:"summary"`
	Parameters []string `json:"parameters"`
}

type Operation struct {
	OperationId string              `json:"operationId"`
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	Summary     string              `json:"summary"`
	Description string              `json:"description,omitempty"`
	Parameters  []RequestParameters `json:"parameters,omitempty"`
	RequestBody RequestBody         `json:"requestBody,omitempty"`
	Responses   []Response          `json:"responses,omitempty"`

	spec    *openapi.Operation
	handler openapi.Handler
}

type RequestParameters struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Required    bool   `json:"required"`
	Schema      *schema.Schema
	Description string `json:"description,omitempty"`
}

type RequestBody struct {
	Description string    `json:"description,omitempty"`
	Required    bool      `json:"required"`
	Contents    []Content `json:"contents"`
}

type Content struct {
	ContentType string         `json:"contentType"`
	Schema      *schema.Schema `json:"schema"`
}

type Response struct {
	StatusCode  int       `json:"statusCode"`
	Description string    `json:"description,omitempty"`
	Contents    []Content `json:"contents"`
}

func (m *mokapi) getHttpApi(name string) any {
	for _, api := range m.app.Http.List() {
		if api.Info.Name == name {
			result := &OpenAPI{
				Name:    name,
				Type:    "http",
				info:    api,
				handler: api.Handler(m.app.Monitor.Http, m.app.Engine, m.app.Events),
			}
			for _, server := range api.Servers {
				result.Servers = append(result.Servers, OpenAPIServer{
					Url:         server.Url,
					Description: server.Description,
				})
			}

			return result
		}
	}
	return nil
}
func (o *OpenAPI) GetOperations() []OperationSummary {
	var result []OperationSummary
	for _, p := range o.info.Paths {
		if p.Value == nil {
			continue
		}
		for method, op := range p.Value.Operations() {
			os := OperationSummary{
				Id:      getOperationId(method, op),
				Method:  method,
				Path:    p.Value.Path,
				Summary: op.Summary,
			}

			if os.Summary == "" {
				os.Summary = p.Value.Summary
			}

			params := append(op.Path.Parameters, op.Parameters...)
			for _, param := range params {
				if param.Value == nil {
					continue
				}
				os.Parameters = append(os.Parameters, param.Value.Name)
			}
			slices.SortStableFunc(os.Parameters, func(a, b string) int {
				return strings.Compare(a, b)
			})

			result = append(result, os)
		}
	}

	slices.SortStableFunc(result, func(a, b OperationSummary) int {
		c := strings.Compare(a.Path, b.Path)
		if c != 0 {
			return c
		}
		return strings.Compare(a.Method, b.Method)
	})

	return result
}

func (o *OpenAPI) GetOperation(id string) (*Operation, error) {
	for _, p := range o.info.Paths {
		if p.Value == nil {
			continue
		}
		for method, op := range p.Value.Operations() {

			operationId := getOperationId(method, op)
			if id != operationId {
				continue
			}

			r := &Operation{
				OperationId: operationId,
				Method:      method,
				Path:        p.Value.Path,
				Summary:     op.Summary,
				Description: op.Description,
				spec:        op,
				handler:     o.handler,
			}
			for _, param := range op.Parameters {
				if param.Value == nil {
					continue
				}
				r.Parameters = append(r.Parameters, RequestParameters{
					Name:        param.Value.Name,
					In:          param.Value.Type.String(),
					Required:    param.Value.Required,
					Schema:      param.Value.Schema,
					Description: param.Value.Description,
				})
			}
			slices.SortStableFunc(r.Parameters, func(a, b RequestParameters) int {
				return strings.Compare(a.Name, b.Name)
			})

			if op.RequestBody != nil && op.RequestBody.Value != nil {
				r.RequestBody = RequestBody{
					Description: op.RequestBody.Value.Description,
					Required:    op.RequestBody.Value.Required,
				}
				for ct, content := range op.RequestBody.Value.Content {
					r.RequestBody.Contents = append(r.RequestBody.Contents, Content{
						ContentType: ct,
						Schema:      content.Schema,
					})
				}
			}
			for it := op.Responses.Iter(); it.Next(); {
				status, err := strconv.Atoi(it.Key())
				if err != nil {
					continue
				}

				res := it.Value()
				if res.Value == nil {
					continue
				}
				var contents []Content
				for ct, content := range res.Value.Content {
					contents = append(contents, Content{
						ContentType: ct,
						Schema:      content.Schema,
					})
				}
				r.Responses = append(r.Responses, Response{
					StatusCode:  status,
					Description: res.Value.Description,
					Contents:    contents,
				})
			}

			return r, nil
		}
	}
	return nil, fmt.Errorf("operation with ID '%s' not found. Hint: Use getOperations() to see the full list of valid IDs", id)
}

type InvokeRequest struct {
	Path   map[string]string   `json:"path"`
	Query  map[string]string   `json:"query"`
	Header map[string][]string `json:"header"`
	Body   string              `json:"body"`
}

type InvokeResponse struct {
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
}

func (op *Operation) Invoke(req InvokeRequest) (InvokeResponse, error) {
	result := InvokeResponse{Headers: make(map[string][]string)}

	var body io.Reader
	if req.Body != "" {
		body = strings.NewReader(req.Body)
	}

	path := op.Path
	query := ""
	params := append(op.spec.Path.Parameters, op.spec.Parameters...)
	for _, p := range params {
		if p.Value == nil {
			continue
		}
		switch p.Value.Type {
		case openapi.ParameterPath:
			if req.Path == nil {
				return result, fmt.Errorf("invoke request %s %s failed: missing path parameter '%s'", op.Method, op.Path, p.Value.Name)
			}
			val, ok := req.Path[p.Value.Name]
			if !ok {
				return result, fmt.Errorf("invoke request %s %s failed: missing path parameter '%s'", op.Method, op.Path, p.Value.Name)
			}
			path = strings.ReplaceAll(path, fmt.Sprintf("{%s}", p.Value.Name), val)
		case openapi.ParameterQuery:
			if req.Query == nil && p.Value.Required {
				return result, fmt.Errorf("invoke request %s %s failed: missing query parameter '%s'", op.Method, op.Path, p.Value.Name)
			}
			val, ok := req.Query[p.Value.Name]
			if !ok {
				if !p.Value.Required {
					continue
				}
				return result, fmt.Errorf("invoke request %s %s failed: missing query parameter '%s'", op.Method, op.Path, p.Value.Name)
			}
			if query != "" {
				query += "&"
			}
			query += fmt.Sprintf("%s=%s", p.Value.Name, val)
		}
	}

	if query != "" {
		path += "?" + query
	}

	r, err := http.NewRequest(op.Method, path, body)
	if err != nil {
		return result, fmt.Errorf("error creating request: %w", err)
	}
	for _, p := range params {
		if p.Value == nil || p.Value.Type != openapi.ParameterHeader {
			continue
		}
		if req.Header == nil && p.Value.Required {
			return result, fmt.Errorf("invoke request %s %s failed: missing header parameter '%s'", op.Method, op.Path, p.Value.Name)
		}
		val, ok := req.Header[p.Value.Name]
		if !ok {
			if !p.Value.Required {
				continue
			}
			return result, fmt.Errorf("invoke request %s %s failed: missing header parameter '%s'", op.Method, op.Path, p.Value.Name)
		}
		r.Header[textproto.CanonicalMIMEHeaderKey(p.Value.Name)] = val
	}

	he := op.handler.ServeHTTP(&result, r)
	if he != nil {
		result.StatusCode = he.StatusCode
		result.Body = he.Message
	}
	return result, nil
}

func (r *InvokeResponse) Header() http.Header {
	return r.Headers
}

func (r *InvokeResponse) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
}

func (r *InvokeResponse) Write(body []byte) (int, error) {
	r.Body = string(body)
	return len(body), nil
}

func (c *Content) GenerateExample() (any, error) {
	js := schema.ConvertToJsonSchema(c.Schema)
	return generator.New(&generator.Request{Schema: js})
}
