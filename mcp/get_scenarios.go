package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Service) registerGetScenarios(server *mcp.Server) {
	registerTool(server, &mcp.Tool{
		Name: "get_scenarios",
		Description: `Lists available scenarios for generating Mokapi scripts.

Use this tool BEFORE calling template tools (e.g., "get_http_mock_template") 
to discover supported scenarios.

Typical workflow:
1. Call this tool to find a suitable scenario
2. Call the corresponding template tool with the selected scenario
3. Adapt the template to your use case`,
	}, s.GetScenarios)
}

func (s *Service) GetScenarios(_ context.Context, _ any) (map[string]any, error) {
	return map[string]any{
		"http": []map[string]any{
			{
				"name":        "dynamic-path-params",
				"description": "Access and use path parameters (e.g., /pets/{petId}) to retrieve or process specific resources based on request.path values.",
			},
			{
				"name":        "conditional-response",
				"description": "Return different responses based on request data (path, query, headers, or body), such as selecting resources, updating state, or handling different HTTP methods.",
			},
			{
				"name":        "static-error-simulation",
				"description": "Return predefined error responses (e.g., 400, 404, 500) for specific endpoints or conditions without dynamic logic.",
			},
			{
				"name":        "dynamic-error-simulation",
				"description": "Return error responses based on runtime conditions, such as missing resources, validation failures, or conflicting state.",
			},
			{
				"name":        "delay-latency",
				"description": "Simulate network latency or slow backend processing by delaying the response before returning data or errors.",
			},
			{
				"name":        "forward-request-to-real-backend",
				"description": "Stop API drift in its tracks. Use Mokapi as a validation layer to enforce OpenAPI contracts between clients and backends,\nregardless of who's calling or what they're building. This scenario forwards incoming requests to real backend services while\nvalidating both requests and responses against the OpenAPI specification.",
			},
		},
	}, nil
}
