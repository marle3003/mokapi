package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetSpecInput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (s *Service) registerGetSpecTool(server *mcp.Server) {
	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "The exact name of the API as returned by 'get_api_list'",
			},
			"type": map[string]any{
				"type":        "string",
				"description": "The type of the API as returned by 'get_api_list'",
				"enum":        []string{"http", "kafka", "ldap", "mail"},
			},
		},
	}

	registerTool(server, &mcp.Tool{
		Name: "get_api_spec",
		Description: `Get the full API specification for a specific API.

		This tool should be used AFTER calling 'get_api_list' to find available APIs, then call this tool with the exact 'name' and 'type'.

		Returns the complete specification including endpoints, operations, and schemas.`,
		InputSchema: inputSchema,
	}, s.GetApiSpec)
}

func (s *Service) GetApiSpec(_ context.Context, in GetSpecInput) (any, error) {
	switch in.Type {
	case "http":
		info := s.app.GetHttp(in.Name)
		if info == nil {
			return nil, fmt.Errorf("http api spec not found")
		}
		return info.Config, nil
	case "kafka":
		info := s.app.Kafka.Get(in.Name)
		if info == nil {
			return nil, fmt.Errorf("kafka api spec not found")
		}
		return info.Config, nil
	case "ldap":
		info := s.app.Ldap.Get(in.Name)
		if info == nil {
			return nil, fmt.Errorf("ldap api spec not found")
		}
		return info.Config, nil
	case "mail":
		info := s.app.Mail.Get(in.Name)
		if info == nil {
			return nil, fmt.Errorf("mail api spec not found")
		}
		return info.Config, nil
	}

	return nil, fmt.Errorf("invalid type: %s", in.Type)
}
