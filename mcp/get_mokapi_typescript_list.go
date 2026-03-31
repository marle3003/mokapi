package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetMokapiTypeScriptApiListOutput struct {
	Packages []MokapiPackage `json:"packages"`
}

type MokapiPackage struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Service) registerGetMokapiTypeScriptList(server *mcp.Server) {
	registerTool(server, &mcp.Tool{
		Name: "get_mokapi_typescript_list",
		Description: `Lists available Mokapi TypeScript API packages.

Use this tool to discover which package to use before requesting detailed type definitions.

Typical workflow:
1. Call this tool to find the correct package (e.g., mokapi/http)
2. Call "get_mokapi_typescript_api" with the selected package
3. Use the returned types to implement the mock script`,
	}, s.GetMokapiTypeScriptApi)
}

func (s *Service) GetMokapiTypeScriptList(_ context.Context, _ any) (GetMokapiTypeScriptApiListOutput, error) {
	return GetMokapiTypeScriptApiListOutput{
		Packages: []MokapiPackage{
			{
				Name: "mokapi",
				Description: `Mokapi JavaScript API
This module exposes the core scripting API for Mokapi.
It allows you to intercept and manipulate protocol events (HTTP, Kafka, LDAP, SMTP),
schedule jobs, generate mock data, and share state between scripts.`,
			},
			{
				Name: "mokapi/http",
				Description: `Utilities for sending HTTP requests and handling HTTP interactions within Mokapi scripts.
Use these functions to simulate client calls, test API integrations, or trigger endpoints from scripts.`,
			},
			{
				Name: "mokapi/kafka",
				Description: `Utilities for producing and consuming messages on Kafka topics.
This package allows you to mock message streams, inspect events, and simulate Kafka-based workflows.`,
			},
			{
				Name: "mokapi/faker",
				Description: `Generates realistic random test data based on JSON schemas or attribute names.
Use this to populate mock responses or generate dynamic content for API responses.`,
			},
			{
				Name: "mokapi/file",
				Description: `File system utilities for reading, writing, and manipulating files within Mokapi scripts.
Useful for mocking file-based APIs, loading fixtures, or storing script state.`,
			},
		},
	}, nil
}
