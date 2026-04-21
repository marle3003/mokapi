package mcp

import (
	"context"
	_ "embed"
	"mokapi/npm/go-mokapi/types"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed data/mocking-types.md
var overview string

var mockTypes = map[string]string{
	"mokapi":   types.Mokapi,
	"faker":    types.Faker,
	"http":     types.Http,
	"kafka":    types.Kafka,
	"mustache": types.Mustache,
	"yaml":     types.Yaml,
}

func addResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		URI:         "mokapi://lib/automation",
		Name:        "Automation API Reference",
		Description: "Types for querying Mokapi specs, logs, and events via Code Mode",
	}, func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      "mokapi://lib/automation",
					MIMEType: "application/typescript",
					Text:     automation,
				},
			},
		}, nil
	})

	server.AddResource(&mcp.Resource{
		URI:  "mokapi://lib/mocking/types",
		Name: "Mokapi Script API Overview",
		Description: `Overview of TypeScript definitions for mock event handlers.
Provides URIs to technical definitions (d.ts) for 'mokapi', 'http', 'kafka', etc.
- Use these types to ensure correct syntax for 'import { ... } from "mokapi/..."'.
- Refer to mokapi://lib/mocking/scenarios for boilerplate examples and usage patterns.
Mandatory for generating valid mock scripts.`,
	}, func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      "mokapi://lib/mocking/types",
					MIMEType: "text/markdown",
					Text:     overview,
				},
			},
		}, nil
	})

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "mokapi://lib/mocking/types/{name}",
		Name:        "Mocking Script API Reference",
		Description: `Types for writing event handlers and mock logic inside Mokapi. 
Use mokapi://lib/mocking/types to get an overview of all available APIs`,
	}, func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		name := extractName(request.Params.URI)
		text, ok := mockTypes[name]
		if !ok {
			return nil, mcp.ResourceNotFoundError(request.Params.URI)
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      "mokapi://lib/mocking/types/" + name,
					MIMEType: "application/typescript",
					Text:     text,
				},
			},
		}, nil
	})

	addMockScenarios(server)
}

//go:embed data/scenarios.md
var scenarioOverview string

//go:embed data/dynamic-path-params.md
var dynamicPathParams string

//go:embed data/conditional-response.md
var conditionalResponse string

//go:embed data/static-error-simulation.md
var staticErrorSimulation string

//go:embed data/dynamic-error-simulation.md
var dynamicErrorSimulation string

//go:embed data/delay-latency.md
var delayLatency string

//go:embed data/forward-request-to-real-backend.md
var forwardRequestToRealBackend string

var scenarios = map[string]string{
	"dynamic-path-params":             dynamicPathParams,
	"conditional-response":            conditionalResponse,
	"static-error-simulation":         staticErrorSimulation,
	"dynamic-error-simulation":        dynamicErrorSimulation,
	"delay-latency":                   delayLatency,
	"forward-request-to-real-backend": forwardRequestToRealBackend,
}

func addMockScenarios(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		URI:  "mokapi://lib/mocking/scenarios",
		Name: "Mokapi Script Blueprints & Scenarios",
		Description: `A directory of ready-to-use code examples and boilerplates. 
Use these to understand how to implement specific use cases like latency, error simulation or conditional behavior. 
Always check a scenario before writing a script from scratch to ensure you follow best practices.`,
	}, func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      "mokapi://lib/mocking/scenarios",
					MIMEType: "text/markdown",
					Text:     scenarioOverview,
				},
			},
		}, nil
	})

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "mokapi://lib/mocking/scenarios/{name}",
		Name:        "Mocking Scenario Detail",
		Description: "Full script template for a specific scenario. Use mokapi://lib/mocking/scenarios get an overview of all scenarios",
	}, func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		name := extractName(request.Params.URI)
		text, ok := scenarios[name]
		if !ok {
			return nil, mcp.ResourceNotFoundError(request.Params.URI)
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      "mokapi://lib/mocking/scenarios/" + name,
					MIMEType: "text/markdown",
					Text:     text,
				},
			},
		}, nil
	})
}
