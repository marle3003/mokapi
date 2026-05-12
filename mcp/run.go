package mcp

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"mokapi/js/compiler"
	"mokapi/js/faker"
	"mokapi/providers/openapi"
	"mokapi/runtime"
	"mokapi/runtime/search"
	"mokapi/schema/json/generator"
	"reflect"
	"slices"
	"strings"

	"github.com/dop251/goja"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	log "github.com/sirupsen/logrus"
)

const errorHint = `Tip for Correction:
It seems there is a syntax error or a misunderstanding of the API. 
To ensure you are using the correct global variables and methods:
1. Call 'mokapi_get_automation_definitions' without parameters to see the general overview.
2. Check 'category="core"' to verify the syntax of the global 'mokapi' object.`

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
Do not guess the API. Always call mokapi_get_automation_definitions first.

MANDATORY WORKFLOW:
1. FIRST: Call 'mokapi_get_automation_definitions' to get the latest API types.
2. SECOND: Use this tool to query live API data (endpoints, schemas, events).

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
			ue := ex.Unwrap()
			if ue == nil {
				return nil, fmt.Errorf("%w\n\n%s", err, errorHint)
			}
			return nil, ue
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
	_ = obj.Set("getEvent", m.getEvent)
	_ = obj.Set("search", m.search)
}

func (m *mokapi) getApis() []ApiSummary {
	var result []ApiSummary
	for _, api := range m.app.Http.List() {
		if api.Info.Name == "" {
			log.Warnf("mcp tool mokapi_execute_code: skip empty HTTTP API name")
			continue
		}
		result = append(result, ApiSummary{
			Name: api.Info.Name,
			Type: "http",
		})
	}
	for _, api := range m.app.Kafka.List() {
		if api.Info.Name == "" {
			log.Warnf("mcp tool mokapi_execute_code: skip empty Kafka API name")
			continue
		}
		result = append(result, ApiSummary{
			Name: api.Info.Name,
			Type: "kafka",
		})
	}
	slices.SortStableFunc(result, func(a, b ApiSummary) int {
		return strings.Compare(a.Name, b.Name)
	})
	return result
}

func (m *mokapi) getApi(name string) any {
	var api any
	api = m.getHttpApi(name)
	if api != nil {
		return api
	}
	api = m.getKafkaApi(name)
	return api
}

func (m *mokapi) fake(v goja.Value) (any, error) {
	js, err := faker.ToJsonSchema(v, m.vm)
	if err != nil {
		return nil, err
	}
	return generator.New(&generator.Request{Schema: js})
}

type SearchResult struct {
	Items []SearchResultItem `json:"items"`
	Total uint64             `json:"total"`
}

type SearchResultItem struct {
	Type      string            `json:"type"`
	Domain    string            `json:"domain,omitempty"`
	Title     string            `json:"title"`
	Fragments []string          `json:"fragments,omitempty"`
	Metadata  map[string]string `json:"metadata"`
	Time      string            `json:"time,omitempty"`
}

func (m *mokapi) search(queryText string, index int, limit int) (SearchResult, error) {
	if limit == 0 {
		limit = 10
	}
	r := search.Request{
		QueryText: queryText,
		Index:     index,
		Limit:     limit,
	}

	sr, err := m.app.Search(r)
	if err != nil {
		return SearchResult{}, err
	}
	result := SearchResult{
		Total: sr.Total,
	}
	for _, item := range sr.Results {
		result.Items = append(result.Items, SearchResultItem{
			Type:      item.Type,
			Domain:    item.Domain,
			Title:     item.Title,
			Fragments: item.Fragments,
			Metadata:  item.Params,
			Time:      item.Time,
		})
	}
	return result, nil
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
