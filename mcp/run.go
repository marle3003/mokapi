package mcp

import (
	"context"
	_ "embed"
	"mokapi/runtime"
	"reflect"
	"slices"
	"strings"

	"github.com/dop251/goja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	log "github.com/sirupsen/logrus"
)

//go:embed run.ts
var types string

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

Important:
Before writing any code, be sure to read the API definitions at api://execute-types to understand
the available global objects, functions, and types.

Use this tool to:
- Explore mocked APIs (OpenAPI, AsyncAPI, LDAP, Mail)
- Inspect operations and schemas
- Invoke API operations directly

Prefer this tool over retrieving full API specifications, as it returns only the computed result.`,
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
	}, s.GenerateHttpMockResponse)

	server.AddResource(&mcp.Resource{
		URI:  "api://execute-types",
		Name: "api-docs",
	}, func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      "api://types",
					MIMEType: "application/typescript",
					Text:     types,
				},
			},
		}, nil
	})
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
	app *runtime.App
	vm  *goja.Runtime
}

func newMokapi(app *runtime.App) *mokapi {
	vm := goja.New()
	vm.SetFieldNameMapper(&customFieldNameMapper{})
	return &mokapi{app: app, vm: vm}
}

func (m *mokapi) run(code string) (any, error) {
	obj := m.vm.NewObject()
	m.init(obj)
	_ = m.vm.Set("mokapi", obj)
	v, err := m.vm.RunString(code)
	if err != nil {
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
				Name: name,
				Type: "http",
				info: api,
			}
		}
	}
	return nil
}

type OpenAPI struct {
	Name string `json:"name"`
	Type string `json:"type"`

	info *runtime.HttpInfo
}

type OperationSummary struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Summary string `json:"summary"`
}

type OperationDetails struct {
	OperationId string `json:"operationId"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

func (o *OpenAPI) GetOperations() []OperationSummary {
	var result []OperationSummary
	for _, p := range o.info.Paths {
		if p.Value == nil {
			continue
		}
		for method, op := range p.Value.Operations() {
			summary := op.Summary
			if summary == "" {
				summary = p.Value.Summary
			}
			result = append(result, OperationSummary{
				Method:  method,
				Path:    p.Value.Path,
				Summary: summary,
			})
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

func (o *OpenAPI) GetOperationDetails(path, method string) *OperationDetails {
	for _, p := range o.info.Paths {
		if p.Value == nil || p.Value.Path != path {
			continue
		}
		op := p.Value.Operation(method)
		if op == nil {
			continue
		}
		return &OperationDetails{
			OperationId: op.OperationId,
			Method:      method,
			Path:        p.Value.Path,
			Summary:     op.Summary,
			Description: op.Description,
		}
	}
	return nil
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
