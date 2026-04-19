package mcp

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"mokapi/js/compiler"
	"mokapi/js/faker"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema"
	"mokapi/runtime"
	"mokapi/schema/json/generator"
	"net/http"
	"net/textproto"
	"reflect"
	"slices"
	"strings"

	"github.com/dop251/goja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	log "github.com/sirupsen/logrus"
)

type RunInput struct {
	Code string `json:"code"`
}

type RunOutput struct {
	Result any `json:"result"`
}

func (s *Service) registerRunTool(server *mcp.Server) {
	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"code": map[string]any{
				"type":        "string",
				"description": "JavaScript code to execute in the Mokapi runtime. The last expression is returned as the result.",
			},
		},
		"required": []string{"code"},
	}

	outputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"result": map[string]any{
				"description": "The result of the executed code.",
				"nullable":    true,
			},
		},
		"required": []string{"result"},
	}

	registerTool(server, &mcp.Tool{
		Name: "mokapi_execute_code",
		Description: `Executes JavaScript code in a sandboxed Mokapi runtime.
The last expression in the code is returned as the result.

MANDATORY WORKFLOW:
1. FIRST: Call 'mokapi_get_automation_definitions' to get the latest API types.
2. SECOND: Use this tool to query live API data (endpoints, schemas, events).
NEVER guess the API structure; always use the definitions as a reference.

Important for Object Returns:
JavaScript interprets {} at the start of a line as a block, not an object. To return an object literal, wrap it in parentheses ({ ... }) or assign it to a variable and put the variable name in the last line.
Example: const result = { a: 1 }; result


Use this tool to:
- Explore mocked APIs (OpenAPI, AsyncAPI, LDAP, Mail)
- Inspect operations and schemas
- Invoke API operations directly`,
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
	}, s.GetRunResponse)

}

func (s *Service) GetRunResponse(_ context.Context, in RunInput) (RunOutput, error) {
	m := newMokapi(s.app)
	r, err := m.run(in.Code)
	if err != nil {
		return RunOutput{}, err
	}

	return RunOutput{Result: r}, nil
}

type mokapi struct {
	app      *runtime.App
	vm       *goja.Runtime
	compiler *compiler.Compiler
}

func newMokapi(app *runtime.App) *mokapi {
	vm := goja.New()
	vm.SetFieldNameMapper(&customFieldNameMapper{})
	c, _ := compiler.New()
	return &mokapi{app: app, vm: vm, compiler: c}
}

func (m *mokapi) run(code string) (any, error) {
	obj := m.vm.NewObject()
	m.init(obj)
	_ = m.vm.Set("mokapi", obj)
	p, err := m.compiler.Compile("mokapi_execute_code.js", code)
	if err != nil {
		return nil, err
	}
	v, err := m.vm.RunProgram(p)
	if err != nil {
		var ex *goja.Exception
		if errors.As(err, &ex) {
			return nil, ex.Unwrap()
		}
		return nil, err
	}
	return v.Export(), nil
}

type ApiSummary struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (m *mokapi) init(obj *goja.Object) {
	_ = obj.Set("getApis", m.getApis)
	_ = obj.Set("getApi", m.getApi)
	_ = obj.Set("fake", m.fake)
	_ = obj.Set("getEvents", m.getEvents)
}

func (m *mokapi) getApis() []ApiSummary {
	var result []ApiSummary
	for _, api := range m.app.ListHttp() {
		if api.Info.Name == "" {
			log.Warnf("mcp tool mokapi_get_api_spec: skip empty HTTTP API name")
			continue
		}
		result = append(result, ApiSummary{
			Name: api.Info.Name,
			Type: "http",
		})
	}
	slices.SortStableFunc(result, func(a, b ApiSummary) int {
		return strings.Compare(a.Name, b.Name)
	})
	return result
}

func (m *mokapi) getApi(name string) any {
	for _, api := range m.app.ListHttp() {
		if api.Info.Name == name {
			return &OpenAPI{
				Name:    name,
				Type:    "http",
				info:    api,
				handler: api.Handler(m.app.Monitor.Http, m.app.Engine, m.app.Events),
			}
		}
	}
	return nil
}

func (m *mokapi) fake(v goja.Value) (any, error) {
	js, err := faker.ToJsonSchema(v, m.vm)
	if err != nil {
		return nil, err
	}
	return generator.New(&generator.Request{Schema: js})
}

type OpenAPI struct {
	Name string `json:"name"`
	Type string `json:"type"`

	info    *runtime.HttpInfo
	handler openapi.Handler
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
	Content     []Content `json:"content"`
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
			return r, nil
		}
	}
	return nil, fmt.Errorf("operation with ID '%s' not found. Hint: Use getOperations() to see the full list of valid IDs", id)
}

func (op *Operation) GetResponseSchema(statusCode int) *Response {
	r := op.spec.Responses.GetResponse(statusCode)
	if r == nil {
		return nil
	}
	result := &Response{
		StatusCode:  statusCode,
		Description: r.Description,
	}
	for ct, content := range r.Content {
		result.Content = append(result.Content, Content{
			ContentType: ct,
			Schema:      content.Schema,
		})
	}
	return result
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

type customFieldNameMapper struct {
}

func (cfm customFieldNameMapper) FieldName(_ reflect.Type, f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if len(tag) == 0 {
		return uncapitalize(f.Name)
	}
	if idx := strings.IndexByte(tag, ','); idx != -1 {
		tag = tag[:idx]
	}

	return tag
}

func (cfm customFieldNameMapper) MethodName(_ reflect.Type, m reflect.Method) string {
	return uncapitalize(m.Name)
}

func uncapitalize(s string) string {
	return strings.ToLower(s[0:1]) + s[1:]
}

func getOperationId(method string, op *openapi.Operation) string {
	if op == nil {
		return ""
	}
	if op.OperationId != "" {
		return op.OperationId
	}
	return strings.ToLower(fmt.Sprintf("%s-%s", method, op.Path.Path))
}
