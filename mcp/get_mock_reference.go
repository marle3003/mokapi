package mcp

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed mock_reference.md
var mockReference string

type GetMockReferenceInput struct {
	Category string `json:"category"`
	Name     string `json:"name"`
}

type GetMockReferenceOutput struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Text     string
}

func (s *Service) registerGetMockReference(server *mcp.Server) {
	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"category": map[string]any{
				"type":        "string",
				"description": "The category of reference material to retrieve.",
				"enum":        []string{"types", "scenarios"},
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "The specific library or scenario name (e.g., 'http', 'kafka', 'delay-latency').",
				"enum": []string{
					// Types
					"mokapi", "http", "kafka", "faker", "mustache", "yaml",
					// Scenarios
					"dynamic-path-params",
					"conditional-response",
					"static-error-simulation",
					"dynamic-error-simulation",
					"delay-latency",
					"forward-request-to-real-backend",
				},
			},
		},
		"required": []string{},
	}

	registerTool(server, &mcp.Tool{
		Name: "mokapi_get_mock_reference",
		Description: `Retrieves technical reference material for writing mock scripts.
You can request either 'types' (API definitions) or 'scenarios' (example blueprints).

Use this tool to:
- See how to import and use 'mokapi', 'mokapi/http', or 'mokapi/kafka'.
- Get boilerplate code for specific use cases (e.g., 'rest-auth').

MANDATORY: Use this before generating a new mock script to ensure correct syntax.`,
		InputSchema: inputSchema,
	}, s.GetMockReference)
}

func (s *Service) GetMockReference(_ context.Context, in *GetMockReferenceInput) (GetMockReferenceOutput, error) {
	if in.Name == "" || in.Category == "" {
		return GetMockReferenceOutput{Text: mockReference}, nil
	}

	if in.Category == "types" {
		text, ok := mockTypes[in.Name]
		if !ok {
			return GetMockReferenceOutput{}, fmt.Errorf("mock reference type not found: %s", in.Name)
		}
		return GetMockReferenceOutput{Text: text, Category: in.Category, Name: in.Name}, nil
	}

	text, ok := scenarios[in.Name]
	if !ok {
		return GetMockReferenceOutput{}, fmt.Errorf("mock reference scenario not found: %s", in.Name)
	}
	return GetMockReferenceOutput{Text: text, Category: in.Category, Name: in.Name}, nil
}
