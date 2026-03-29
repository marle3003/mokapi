package mcp

import (
	"context"
	"mokapi/engine"
	"mokapi/engine/common"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ProduceKafkaMessageInput struct {
	APIName   string            `json:"apiName"`
	Topic     string            `json:"topic"`
	Partition int               `json:"partition"`
	Key       any               `json:"key,omitempty"`
	Value     any               `json:"value"`
	Headers   map[string]string `json:"headers,omitempty"`
	ClientId  string            `json:"clientId"`
}

type ProduceKafkaMessageResponse struct {
	Offset int64 `json:"offset"`
}

func (s *Service) registerProduceKafkaMessage(server *mcp.Server) {
	inputSchema := map[string]any{
		"type":     "object",
		"required": []string{"apiName", "topic", "value"},
		"properties": map[string]any{
			"apiName": map[string]any{
				"type":        "string",
				"description": "The name of the Kafka API as returned by 'get_api_list'",
			},
			"topic": map[string]any{
				"type":        "string",
				"description": "Kafka topic name",
			},
			"partition": map[string]any{
				"type":        "integer",
				"description": "Partition number where to write the message to",
			},
			"key": map[string]any{
				"description": "Optional message key",
			},
			"value": map[string]any{
				"description": "Message payload",
			},
			"headers": map[string]any{
				"type":        "object",
				"description": "Optional message headers",
				"additionalProperties": map[string]any{
					"type": "string",
				},
			},
			"clientId": map[string]any{
				"type":        "string",
				"description": "ClientId of the producer",
			},
		},
	}

	outputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"offset": map[string]any{
				"type":        "integer",
				"description": "The offset of the produced message",
			},
		},
	}

	registerTool(server, &mcp.Tool{
		Name: "send_http_request",
		Description: `Produce a message to a Kafka topic.

Use this tool after retrieving the API specification with 'get_api_spec' to understand available topics and message formats.

Allows sending messages with optional key and headers.`,
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
	}, s.ProduceKafkaMessage)
}

func (s *Service) ProduceKafkaMessage(_ context.Context, in ProduceKafkaMessageInput) (ProduceKafkaMessageResponse, error) {
	result := ProduceKafkaMessageResponse{}

	c := engine.NewKafkaClient(s.app)
	r, err := c.Produce(&common.KafkaProduceArgs{
		Cluster: in.APIName,
		Topic:   in.Topic,
		Messages: []common.KafkaMessage{
			{
				Key:       in.Key,
				Data:      in.Value,
				Headers:   in.Headers,
				Partition: in.Partition,
			},
		},
		Retry:    common.KafkaProduceRetry{},
		ClientId: in.ClientId,
	})
	if err != nil {
		return result, err
	}
	if len(r.Messages) > 0 {
		result.Offset = r.Messages[0].Offset
	}
	return result, nil
}
