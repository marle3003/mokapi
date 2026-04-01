package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	log "github.com/sirupsen/logrus"
)

type ListApisInput struct {
	Type string `json:"type,omitempty"`
}

type Api struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ListApiResponse struct {
	Apis []Api `json:"apis"`
}

func (s *Service) registerListApiTool(server *mcp.Server) {
	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"type": map[string]any{
				"type":        "string",
				"description": "Filter APIs by type. Use 'http' for REST/OpenAPI APIs, 'kafka' for AsyncAPI topics, 'ldap' for directory services, or 'mail' for SMTP/IMAP.",
				"enum":        []string{"http", "kafka", "ldap", "mail"},
			},
		},
	}

	outputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"apis": map[string]any{
				"type":        "array",
				"description": "The list of mocked apis",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type":        "string",
							"description": "The name of the API",
						},
						"type": map[string]any{
							"type":        "string",
							"description": "The type of the API",
							"enum":        []string{"http", "kafka", "ldap", "mail"},
						},
					},
				},
			},
		},
	}

	registerTool(server, &mcp.Tool{
		Name: "get_api_list",
		Description: `Returns all available APIs with their name and type. 
		Use this to discover APIs before calling 'get_api_spec' to retrieve detailed specifications.`,
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
	}, s.ListApis)
}

func (s *Service) ListApis(_ context.Context, in ListApisInput) (*ListApiResponse, error) {
	var result []Api

	if in.Type == "http" || len(in.Type) == 0 {
		for _, api := range s.app.ListHttp() {
			if api.Info.Name == "" {
				log.Warnf("mcp tool get_api_list: skip empty HTTTP API name")
				continue
			}
			result = append(result, Api{
				Name: api.Info.Name,
				Type: "http",
			})
		}
	}

	if in.Type == "kafka" || len(in.Type) == 0 {
		for _, api := range s.app.Kafka.List() {
			if api.Info.Name == "" {
				log.Warnf("mcp tool get_api_list: skip empty Kafka API name")
				continue
			}
			result = append(result, Api{
				Name: api.Info.Name,
				Type: "kafka",
			})
		}
	}

	if in.Type == "ldap" || len(in.Type) == 0 {
		for _, api := range s.app.Ldap.List() {
			if api.Info.Name == "" {
				log.Warnf("mcp tool get_api_list: skip empty LDAP API name")
				continue
			}
			result = append(result, Api{
				Name: api.Info.Name,
				Type: "ldap",
			})
		}
	}

	if in.Type == "mail" || len(in.Type) == 0 {
		for _, api := range s.app.Mail.List() {
			if api.Info.Name == "" {
				log.Warnf("mcp tool get_api_list: skip empty Mail API name")
				continue
			}
			result = append(result, Api{
				Name: api.Info.Name,
				Type: "mail",
			})
		}
	}

	return &ListApiResponse{Apis: result}, nil
}
