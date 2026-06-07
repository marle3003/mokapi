package mcp

import (
	"context"
	_ "embed"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed data/automation.md
var automation string

//go:embed data/automation-core.md
var automationCore string

//go:embed data/automation-http.md
var automationHttp string

//go:embed data/automation-kafka.md
var automationKafka string

//go:embed data/automation-mail.md
var automationMail string

//go:embed data/automation-ldap.md
var automationLdap string

//go:embed data/automation-event.md
var automationEvent string

type AutomationDefinitionsInput struct {
	Category string `json:"category"`
}

type AutomationDefinitionsOutput struct {
	Text string `json:"text"`
}

func (s *Service) registerGetAutomationDefinitions(server *mcp.Server) {
	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"category": map[string]any{
				"type":        "string",
				"description": "The category of type definition to retrieve. If omitted, a general overview is returned.",
				"enum":        []string{"core", "http", "kafka", "event"},
			},
		},
		"required": []string{},
	}

	registerTool(server, &mcp.Tool{
		Name: "mokapi_get_automation_definitions",
		Description: `Returns the required TypeScript definitions ONLY for the Mokapi Automation API (tool mokapi_execute_code).

MANDATORY: Call this tool BEFORE using 'mokapi_execute_code'. If you are unsure which category you need, call this tool without any parameters to receive a general overview of all available categories.
- Learn how to query API specifications (OpenAPI, AsyncAPI)
- Access methods for inspecting live logs and events
- Get correct syntax for the global 'mokapi' object

This ensures your generated code is valid and uses the correct library methods.`,
		InputSchema: inputSchema,
	}, s.GetAutomationDefinitions)
}

func (s *Service) GetAutomationDefinitions(_ context.Context, in AutomationDefinitionsInput) (AutomationDefinitionsOutput, error) {
	var text string
	switch in.Category {
	case "core":
		text = automationCore
	case "http":
		text = automationHttp
	case "kafka":
		text = automationKafka
	case "mail":
		text = automationMail
	case "ldap":
		text = automationLdap
	case "event":
		text = automationEvent
	default:
		text = automation
	}

	return AutomationDefinitionsOutput{Text: text}, nil
}
