package mcp

import (
	"context"
	_ "embed"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed run.ts
var automationTypes string

type AutomationDefinitions struct {
	Code string
}

func (s *Service) registerGetAutomationDefinitions(server *mcp.Server) {
	registerTool(server, &mcp.Tool{
		Name: "mokapi_get_automation_definitions",
		Description: `Returns the required TypeScript definitions for the Mokapi Automation API.

MANDATORY: Call this tool BEFORE using 'mokapi_execute_code' to:
- Learn how to query API specifications (OpenAPI, AsyncAPI)
- Access methods for inspecting live logs and events
- Get correct syntax for the global 'mokapi' object

This ensures your generated code is valid and uses the correct library methods.`,
	}, s.GetAutomationDefinitions)
}

func (s *Service) GetAutomationDefinitions(_ context.Context, _ any) (AutomationDefinitions, error) {
	return AutomationDefinitions{Code: automationTypes}, nil
}
