package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetHttpMockTemplateInput struct {
	Scenario string `json:"scenario"`
}

type GetHttpMockTemplateOutput struct {
	Scenario    string `json:"scenario"`
	Description string `json:"description"`
	Code        string `json:"code"`
}

func (s *Service) registerGetHttpMockTemplate(server *mcp.Server) {
	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"scenario": map[string]any{
				"type":        "string",
				"description": "Scenario to generate boilerplate",
			},
		},
		"required": []string{"scenario"},
	}

	registerTool(server, &mcp.Tool{
		Name: "mokapi_get_http_mock_template",
		Description: `Templates demonstrate how to implement logic for HTTP mocks.

This tool should be used AFTER calling 'get_scenarios' to find available scenarios.

Use "generate_http_mock_response" to:
- Get a valid response structure based on OpenAPI
- Avoid guessing response formats

You can:
- Replace result.data with custom data (e.g., from a list or database)
- Modify the response based on request parameters

Typical pattern:
1. Apply your business logic (e.g., find a resource)
2. Call generate_http_mock_response with the correct status code
3. Override result.data if needed
4. Assign result to response
`,
		InputSchema: inputSchema,
	}, s.GetHttpMockTemplate)
}

func (s *Service) GetHttpMockTemplate(_ context.Context, in GetHttpMockTemplateInput) (GetHttpMockTemplateOutput, error) {
	switch in.Scenario {
	case "dynamic-path-params":
		return GetHttpMockTemplateOutput{
			Scenario: "dynamic-path-params",
			Description: `HTTP mock handler to get a pet stored in a array list.
Demonstrates how to:
- Access request parameters
- Apply custom logic (e.g., lookup, filtering)`,
			Code: `
import { on } from "mokapi"

let pets = [
  { id: 1, name: 'Fluffy', status: 'available', category: { id: 1, name: 'Dogs' }, photoUrls: [], tags: [] },
  { id: 3, name: 'Hedgie', status: 'pending', category: { id: 2, name: 'Small Animals' }, photoUrls: [], tags: [] }
];

export default function () {
  on('http', async (request, response) => {
    switch(request.key) {
      case '/pets/{id}':
        if (request.method !== 'GET') {
          return
        }
        const pet = pets.find(x => x.id === request.path.id)
        if (pet) {
          response.data = pet
        } else {
          console.log('pet not found', request)
          response.rebuild(404)
        }
	}
  })
}
`,
		}, nil
	case "conditional-response":
		return GetHttpMockTemplateOutput{
			Scenario: "conditional-response",
			Description: `HTTP mock handler for terminals.
Demonstrates how to:
- Access request parameters
- Apply custom logic (e.g., lookup, filtering, updates)
`,
			Code: `
import { on } from "mokapi"

interface Terminal {
	id: string
	compartments: {
		id: string
		doorState: 'open' | 'closed'
	}[]
}

const terminals: Terminal[] = []

export default function () {
	on('http', (request, response) => {
		switch(request.key) {
			case '/terminals/{id}': {
				const terminal = terminals.find(x => x.id === request.path.id)
				if (!terminal) {
					response.rebuild(404)
					response.data = { error: 'terminal not found' }
					return
				}

				if (request.method === 'GET') {
					response.data = terminal
				} else if (request.method === 'POST') {
					// update the terminal
					Object.assign(terminal, request.body)
					// mokapi already set the success response, nothing to do
				}
				// do not raise an error if different method is used,
				// maybe there is another event handler in a different file defined
				return
			}
			case '/terminals': {
				if (request.method === 'GET') {
					response.data = terminals
				} else if (request.method === 'POST') {
					const terminal = terminals.find(x => x.id === request.path.id)
					if (terminal) {
						// console output will be displayed in the Mokapi's' dashboard
						console.log('terminal already exists', request.body)
						response.rebuild(400)
					} else {
						terminals.push(request.body)
					}
				}
				return
			}
		}
	})
}
`,
		}, nil
	case "static-error-simulation":
		return GetHttpMockTemplateOutput{
			Scenario:    "static-error-simulation",
			Description: `Return predefined error responses (e.g., 400, 404, 500) for specific endpoints or conditions without dynamic logic.`,
			Code: `
import { on } from "mokapi"

export default function () {
	on('http', (request, response) => {
		switch(request.key) {
			case '/bookings': {
				if (request.method === 'POST') {
					if (request.header['Api-Key'] === 'invalid') {
						// console output will be displayed in the Mokapi's' dashboard
						console.log('api-key is not valid')
						response.rebuild(401)
						return
					}
					if (request.body?.hotel?.code === 'NOT_FOUND') {
						console.log('hotel not found')
						response.rebuild(404)
						return
					}
					if (request.body.hotel.name === 'INVALID') {
						console.log('hotel name is not valid')
						response.rebuild(400)
						return
					}
				}
			}
		}
	})
}
`,
		}, nil
	case "dynamic-error-simulation":
		return GetHttpMockTemplateOutput{
			Scenario:    "dynamic-error-simulation",
			Description: `Return error responses based on runtime conditions, such as missing resources, validation failures, or conflicting state.`,
			Code: `
import { on } from "mokapi"

const hotels = []

export default function () {
	on('http', (request, response) => {
		switch(request.key) {
			case '/bookings': {
				const hotel = hotels.find(x => x.code === request.body?.hotel?.code)

				if (!hotel) {
				  console.log('hotel not found')
				  response.rebuild(404)
				  response.data = { error: 'hotel not found' }
				  return
				}

				// simulate dynamic errors based on hotel simulation config
				const type = hotel.simulation?.responseType
				switch (type) {
					case 'bad-request':
						response.rebuild(400)
						return
					case 'unauthorized':
						response.rebuild(401)
						return
					case 'forbidden':
						response.rebuild(403)
						return
					case 'internal-server-error':
						response.rebuild(500)
						return
				}
				// success path: generate valid response
				// ...
			}
		}
	})
}
`,
		}, nil
	case "delay-latency":
		return GetHttpMockTemplateOutput{
			Scenario:    "delay-latency",
			Description: `Simulate server latency by delaying the response. Useful to test frontend loading states, timeouts, or high-load scenarios. Use generate_http_mock_response for schema-compliant response data.`,
			Code: `
import { on } from "mokapi"

let pets = [
  { id: 1, name: 'Fluffy', status: 'available', category: { id: 1, name: 'Dogs' }, photoUrls: [], tags: [] },
  { id: 3, name: 'Hedgie', status: 'pending', category: { id: 2, name: 'Small Animals' }, photoUrls: [], tags: [] }
];

export default function () {
  on('http', async (request, response) => {
    switch(request.key) {
      case '/pets': {
        if (request.method !== 'GET') return

        // simulate network latency (e.g., 2 seconds)
        sleep('2s')

        response.data = pets
      }
    }
  })
}
`,
		}, nil
	case "forward-request-to-real-backend":
		return GetHttpMockTemplateOutput{
			Scenario: "forward-request-to-real-backend",
			Description: `Stop API drift in its tracks. Use Mokapi as a validation layer to enforce OpenAPI contracts between clients and backends,
regardless of who's calling or what they're building. This scenario forwards incoming requests to real backend services while
validating both requests and responses against the OpenAPI specification.`,
			Code: `import { on } from 'mokapi';
import { fetch } from 'mokapi/http';

export default async function () {
    
    on('http', async (request, response) => {

        // Map request to backend URL based on OpenAPI spec name
        const url = getForwardUrl(request)

        // If no URL could be determined, return an error immediately
        if (!url) {
            response.statusCode = 500;
            response.body = 'Failed to forward request: unknown backend';
            return;
        } 
            
        try {
            // Forward the request to the backend
            const res = await fetch(url, {
                method: request.method,
                body: request.body,
                headers: request.header,
                timeout: '30s'
            });

            // Copy status code and headers
            response.statusCode = res.statusCode;
            response.headers = res.headers

            // Check the content type to decide whether to validate the response
            const contentType = res.headers['Content-Type']?.[0] || '';

            if (contentType.includes('application/json')) {
                // Mokapi can validate JSON responses automatically
                response.data = res.json();
            } else {
                // For other content types, skip validation
                response.body = res.body;
            }
            
        } catch (e) {
            // Handle any errors that occur while forwarding
            response.statusCode = 500;
            response.body = e.toString();
        }
    });

    function getForwardUrl(request: HttpRequest): string | undefined {
        switch (request.api) {
            case 'backend-1': {
                return 'https://backend1.example.com' + request.url.path + '?' + request.url.query;
			}
			case 'backend-2': {
				return 'https://backend2.example.com' + request.url.path + '?' + request.url.query;
			}
			default:
				return undefined;
		}
	}
}`,
		}, nil
	}

	return GetHttpMockTemplateOutput{}, fmt.Errorf("unknown scenario")
}
