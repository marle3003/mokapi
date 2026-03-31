package mcp

import (
	"context"
	"fmt"
	"mokapi/media"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema"
	"mokapi/schema/json/generator"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GenerateHttpResponseInput struct {
	ApiName     string `json:"apiName"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	StatusCode  int    `json:"statusCode"`
	ContentType string `json:"contentType,omitempty"`
}

type GenerateHttpResponseOutput struct {
	StatusCode int            `json:"statusCode"`
	Data       any            `json:"data"`
	Headers    map[string]any `json:"headers"`
}

func (s *Service) registerGenerateHttpResponseTool(server *mcp.Server) {
	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"apiName": map[string]any{
				"type":        "string",
				"description": "The exact name of the API as returned by 'get_api_list'",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "The path template of the endpoint (e.g. /pets/{id})",
			},
			"method": map[string]any{
				"type":        "string",
				"description": "The HTTP method (GET, POST, PUT, DELETE, etc.)",
			},
			"statusCode": map[string]any{
				"type":        "integer",
				"description": "The HTTP status code to generate the response for",
			},
			"contentType": map[string]any{
				"type": "string",
				"description": `The HTTP content type of the response body. Optional: 
					If provided, this content type is used.
					If the endpoint has only one content type, it will be used automatically.
					Otherwise defaults to 'application/json'`,
				"default": "application/json",
			},
		},
	}

	outputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"statusCode": map[string]any{
				"type":        "integer",
				"description": "HTTP status code for the response",
			},
			"headers": map[string]any{
				"type":        "object",
				"description": "response headers defined by the API specification",
			},
			"data": map[string]any{
				"type":        "any",
				"description": "structured response body that matches the OpenAPI schema",
			},
		},
	}

	registerTool(server, &mcp.Tool{
		Name: "generate_http_response",
		Description: `Generate a valid HTTP response for a specific API endpoint.

This tool returns a complete response object that already conforms to the OpenAPI specification.
The generated data strictly matches the response schema, including all required fields and correct types.

Use this tool when writing HTTP mock scripts instead of manually constructing response bodies.

The returned object can be used directly in the mock script:

on('http', (request) => {
return GENERATED_RESPONSE
})

The "data" field is preferred and will be automatically encoded based on the API specification.
The "body" field is not returned by this tool and should only be used for raw responses.

Rules:
- Do NOT manually construct complex response objects
- Always prefer this tool to ensure schema-correct responses
- The "data" field contains structured data and will be encoded automatically
- The "statusCode" and "headers" are already set correctly

Example:
Generated response:
{
  "statusCode": 200,
  "headers": {
    "Content-Type": "application/json"
  },
  "data": {
    "id": 1,
    "name": "dog",
    "status": "available"
  }
}

Usage in mock script:
on('http', (request, response) => {
  response.statusCode = 200
  response.headers["Content-Type"] = "application/json"
  response.data = {
    "id": 1,
    "name": "dog",
    "status": "available"
  }
})
`,
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
	}, s.GetHttpResponseSchema)
}

func (s *Service) GenerateHttpResponse(_ context.Context, in GenerateHttpResponseInput) (GenerateHttpResponseOutput, error) {
	result := GenerateHttpResponseOutput{StatusCode: in.StatusCode, Headers: make(map[string]any)}

	info := s.app.GetHttp(in.ApiName)
	if info == nil {
		return result, fmt.Errorf("http api not found")
	}
	p, ok := info.Paths[in.Path]
	if !ok || p.Value == nil {
		return result, fmt.Errorf("path not found")
	}
	o := p.Value.Operation(in.Method)
	if o == nil {
		return result, fmt.Errorf("operation not found")
	}
	r := o.Responses.GetResponse(in.StatusCode)
	if r == nil {
		return result, fmt.Errorf("response not found")
	}

	n := len(r.Content)
	if n == 0 {
		return result, fmt.Errorf("response has no content")
	}

	var mt *openapi.MediaType
	if n == 1 && in.ContentType == "" {
		for _, v := range r.Content {
			mt = v
			break
		}
	} else {
		contentType := "application/json"
		if in.ContentType != "" {
			contentType = in.ContentType
		}
		accept := media.ParseContentType(contentType)
		for k, v := range r.Content {
			key := media.ParseContentType(k)
			if accept.Match(key) {
				mt = v
				break
			}
		}
	}

	if mt == nil {
		return result, fmt.Errorf("response not found")
	}

	segments := strings.Split(p.Value.Path, "/")
	var names []string
	for _, seg := range segments[1:] {
		if !strings.HasPrefix(seg, "{") {
			names = append(names, seg)
		}
	}
	req := generator.NewRequest(
		names,
		schema.ConvertToJsonSchema(mt.Schema),
		nil,
	)

	var err error
	result.Data, err = generator.New(req)
	if err != nil {
		return result, err
	}

	return result, nil
}
