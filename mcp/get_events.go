package mcp

import (
	"context"
	"mokapi/runtime/events"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetEventsInput struct {
	APIName string            `json:"apiName"`
	Type    string            `json:"type"`
	Limit   *int              `json:"limit"`
	Traits  map[string]string `json:"traits"`
}

type GetEventsResponse struct {
	Events []events.Event `json:"events"`
}

func (s *Service) registerGetEvents(server *mcp.Server) {
	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"apiName": map[string]any{
				"type":        "string",
				"description": "Filter events by API name",
			},
			"type": map[string]any{
				"type":        "string",
				"description": "Filter by event type",
				"enum":        []string{"http", "kafka"},
			},
			"traits": map[string]any{
				"type":        "object",
				"description": "Filter events by traits",
				"additionalProperties": map[string]interface{}{
					"type": "string",
				},
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of events to return",
				"default":     10,
			},
		},
	}

	outputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"events": map[string]any{
				"type":        "array",
				"description": "List of events",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{
							"type":        "string",
							"description": "ID of the event",
						},
						"traits": map[string]any{
							"type":        "object",
							"description": "List of traits",
							"additionalProperties": map[string]interface{}{
								"type": "string",
							},
						},
						"data": map[string]any{
							"type":        "object",
							"description": "The data of the event",
						},
						"time": map[string]any{
							"type":        "string",
							"description": "Time of the event",
							"format":      "date-time",
						},
					},
				},
			},
		},
	}

	registerTool(server, &mcp.Tool{
		Name: "get_events",
		Description: `Returns recorded events from Mokapi including HTTP requests/responses and Kafka messages.

Use this tool when the user asks:
- "What requests were made?"
- "Why did my request fail?"
- "Show recent API activity"
- "What messages were produced to Kafka?"

Each event contains:
- metadata: id, time, traits
- HTTP data: request (method, URL, parameters, body) and response (status, headers, body, duration)
- Kafka data: message payload, key, headers, partition, offset

Call this tool after sending requests or producing messages to inspect results and debug behavior.`,
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
	}, s.ProduceKafkaMessage)
}

func (s *Service) GetEvents(_ context.Context, in GetEventsInput) (GetEventsResponse, error) {
	result := GetEventsResponse{}

	traits, err := bindInput[events.Traits](in.Traits)
	if err != nil {
		return result, err
	}
	if traits == nil {
		traits = events.NewTraits()
	}
	if in.Type != "" {
		traits.WithNamespace(in.Type)
	}

	evts := s.app.Events.GetEvents(traits)

	limit := 10
	if in.Limit != nil {
		limit = *in.Limit
	}
	if len(evts) > limit {
		result.Events = evts[0:limit]
	} else {
		result.Events = evts
	}
	return result, nil
}
