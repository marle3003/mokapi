package mcp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SendHttpRequestInput struct {
	APIName string            `json:"apiName"`
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   map[string]string `json:"query,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

type SendHttpRequestResponse struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    string              `json:"body,omitempty"`
}

func (s *Service) registerSendHttpRequest(server *mcp.Server) {
	inputSchema := map[string]any{
		"type":     "object",
		"required": []string{"apiName", "method", "path"},
		"properties": map[string]any{
			"apiName": map[string]any{
				"type":        "string",
				"description": "The name of the API as returned by 'get_api_list'",
			},
			"method": map[string]any{
				"type":        "string",
				"description": "HTTP method to use",
				"enum":        []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
			},
			"path": map[string]any{
				"type":        "string",
				"description": "The endpoint path, e.g. /pets or /pets/{id}",
			},
			"query": map[string]any{
				"type":        "object",
				"description": "Query parameters as key-value pairs",
				"additionalProperties": map[string]any{
					"type": "string",
				},
			},
			"headers": map[string]any{
				"type":        "object",
				"description": "HTTP headers as key-value pairs",
				"additionalProperties": map[string]any{
					"type": "string",
				},
			},
			"body": map[string]any{
				"type":        "string",
				"description": "Request body (JSON object, string, number, etc.)",
			},
		},
	}

	outputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"status": map[string]any{
				"type":        "integer",
				"description": "HTTP status code",
			},
			"headers": map[string]any{
				"type":        "object",
				"description": "Response headers",
				"additionalProperties": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "string",
					},
				},
			},
			"body": map[string]any{
				"description": "Response body",
			},
		},
	}

	registerTool(server, &mcp.Tool{
		Name: "mokapi_send_http_request",
		Description: `Send an HTTP request to a mocked API.

Use this tool AFTER retrieving the API specification with 'mokapi_get_api_spec' to understand available endpoints.

Supports GET, POST, PUT, PATCH, and DELETE requests.
Returns the full response including status code, headers, and body.`,
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
	}, s.SendHttpRequest)
}

func (s *Service) SendHttpRequest(_ context.Context, in SendHttpRequestInput) (SendHttpRequestResponse, error) {
	result := SendHttpRequestResponse{Headers: make(map[string][]string)}

	info := s.app.GetHttp(in.APIName)
	if info == nil {
		return result, fmt.Errorf("API '%s' not found", in.APIName)
	}

	h := info.Handler(s.app.Monitor.Http, s.app.Engine, s.app.Events)

	var body io.Reader
	if in.Body != "" {
		body = strings.NewReader(in.Body)
	}

	r, err := http.NewRequest(in.Method, in.Path, body)
	if err != nil {
		return result, fmt.Errorf("error creating request: %w", err)
	}

	he := h.ServeHTTP(&result, r)
	if he != nil {
		result.Status = he.StatusCode
		if he.StatusCode == http.StatusNotFound && strings.HasPrefix(he.Message, "no matching endpoint found") {
			result.Body = fmt.Sprintf("path '%v' not found", in.Path)
		} else {
			result.Body = he.Message
		}
	}
	return result, nil
}

func (r *SendHttpRequestResponse) Header() http.Header {
	return r.Headers
}

func (r *SendHttpRequestResponse) WriteHeader(statusCode int) {
	r.Status = statusCode
}

func (r *SendHttpRequestResponse) Write(body []byte) (int, error) {
	r.Body = string(body)
	return len(body), nil
}
