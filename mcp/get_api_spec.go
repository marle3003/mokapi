package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	log "github.com/sirupsen/logrus"
)

type GetApiSpecInput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type GetApiSpecOutput struct {
	Apis []ApiSpec `json:"apis"`
}

type ApiSpec struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Spec any    `json:"spec,omitempty"`
}

func (s *Service) registerGetSpecTool(server *mcp.Server) {
	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "The exact name of the API",
				"optional":    true,
			},
			"type": map[string]any{
				"type":        "string",
				"description": "Filter APIs by type. Use 'http' for REST/OpenAPI APIs, 'kafka' for AsyncAPI topics, 'ldap' for directory services, or 'mail' for SMTP/IMAP.",
				"enum":        []string{"http", "kafka", "ldap", "mail"},
				"optional":    true,
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
						"spec": map[string]any{
							"type":        "object",
							"description": "The specification of the API (e.g. OpenAPI or AsyncAPI",
						},
					},
					"required": []any{"name", "type"},
				},
			},
		},
	}

	registerTool(server, &mcp.Tool{
		Name: "mokapi_get_api_spec",
		Description: `Retrieve API specifications from Mokapi.

- DISCOVERY: Call without 'name' to get an overview of all available APIs (names and types). 
- DETAILS: Call with a specific 'name' and 'type' to get the full specification (OpenAPI, AsyncAPI, etc.) including endpoints, schemas, and operations.

Use discovery first if you are unsure which APIs are currently mocked.`,
		InputSchema:  inputSchema,
		OutputSchema: outputSchema,
	}, s.GetApiSpec)
}

func (s *Service) GetApiSpec(_ context.Context, in GetApiSpecInput) (GetApiSpecOutput, error) {
	var result []ApiSpec

	switch in.Type {
	case "", "http", "kafka", "ldap", "mail":
		break
	default:
		return GetApiSpecOutput{}, fmt.Errorf("unknown type: %s", in.Type)
	}

	if in.Name == "" {
		if in.Type == "http" || len(in.Type) == 0 {
			for _, api := range s.app.ListHttp() {
				if api.Info.Name == "" {
					log.Warnf("mcp tool mokapi_get_api_spec: skip empty HTTTP API name")
					continue
				}
				result = append(result, ApiSpec{
					Name: api.Info.Name,
					Type: "http",
				})
			}
		}

		if in.Type == "kafka" || len(in.Type) == 0 {
			for _, api := range s.app.Kafka.List() {
				if api.Info.Name == "" {
					log.Warnf("mcp tool mokapi_get_api_spec: skip empty Kafka API name")
					continue
				}
				result = append(result, ApiSpec{
					Name: api.Info.Name,
					Type: "kafka",
				})
			}
		}

		if in.Type == "ldap" || len(in.Type) == 0 {
			for _, api := range s.app.Ldap.List() {
				if api.Info.Name == "" {
					log.Warnf("mcp tool mokapi_get_api_spec: skip empty LDAP API name")
					continue
				}
				result = append(result, ApiSpec{
					Name: api.Info.Name,
					Type: "ldap",
				})
			}
		}

		if in.Type == "mail" || len(in.Type) == 0 {
			for _, api := range s.app.Mail.List() {
				if api.Info.Name == "" {
					log.Warnf("mcp tool mokapi_get_api_spec: skip empty Mail API name")
					continue
				}
				result = append(result, ApiSpec{
					Name: api.Info.Name,
					Type: "mail",
				})
			}
		}
		return GetApiSpecOutput{Apis: result}, nil
	}

	if in.Type == "http" || len(in.Type) == 0 {
		info := s.app.GetHttp(in.Name)
		if info != nil {
			result = append(result, ApiSpec{
				Name: in.Name,
				Type: "http",
				Spec: info.Config,
			})
		}
	}

	if in.Type == "kafka" || len(in.Type) == 0 {
		info := s.app.Kafka.Get(in.Name)
		if info != nil {
			result = append(result, ApiSpec{
				Name: in.Name,
				Type: "kafka",
				Spec: info.Config,
			})
		}
	}

	if in.Type == "ldap" || len(in.Type) == 0 {
		info := s.app.Ldap.Get(in.Name)
		if info != nil {
			result = append(result, ApiSpec{
				Name: in.Name,
				Type: "ldap",
				Spec: info.Config,
			})
		}
	}

	if in.Type == "mail" || len(in.Type) == 0 {
		info := s.app.Mail.Get(in.Name)
		if info != nil {
			result = append(result, ApiSpec{
				Name: in.Name,
				Type: "mail",
				Spec: info.Config,
			})
		}
	}

	return GetApiSpecOutput{Apis: result}, nil
}
