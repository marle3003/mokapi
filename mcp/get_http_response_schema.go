package mcp

import (
	"context"
	"fmt"
	"mokapi/media"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetHttpResponseSchemaInput struct {
	ApiName     string `json:"apiName"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	StatusCode  int    `json:"statusCode"`
	ContentType string `json:"contentType,omitempty"`
}

func (s *Service) registerGetHttpResponseSchemaTool(server *mcp.Server) {
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
				"description": "The HTTP status code",
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

	registerTool(server, &mcp.Tool{
		Name: "get_http_response_schema",
		Description: `Get the HTTP response body schema for a specific API endpoint.

Use this tool **before generating any HTTP mock script**.
The returned schema defines all required fields, types, and nested structures.
All mock responses must strictly conform to this schema. Do not omit required fields or invent extra ones.
`,
		InputSchema: inputSchema,
	}, s.GetHttpResponseSchema)
}

func (s *Service) GetHttpResponseSchema(_ context.Context, in GetHttpResponseSchemaInput) (any, error) {
	info := s.app.GetHttp(in.ApiName)
	if info == nil {
		return nil, fmt.Errorf("http api not found")
	}
	p, ok := info.Paths[in.Path]
	if !ok || p.Value == nil {
		return nil, fmt.Errorf("path not found")
	}
	o := p.Value.Operation(in.Method)
	if o == nil {
		return nil, fmt.Errorf("operation not found")
	}
	r := o.Responses.GetResponse(in.StatusCode)
	if r == nil {
		return nil, fmt.Errorf("response not found")
	}

	n := len(r.Content)
	if n == 0 {
		return nil, fmt.Errorf("response has no content")
	}
	if n == 1 && in.ContentType == "" {
		for _, v := range r.Content {
			return v.Schema, nil
		}
	}
	contentType := "application/json"
	if in.ContentType != "" {
		contentType = in.ContentType
	}
	mt := media.ParseContentType(contentType)
	for k, v := range r.Content {
		key := media.ParseContentType(k)
		if mt.Match(key) {
			return v.Schema, nil
		}
	}

	return nil, fmt.Errorf("content type not found")
}
