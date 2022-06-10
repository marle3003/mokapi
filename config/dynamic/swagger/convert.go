package swagger

import (
	"fmt"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/openapi/parameter"
	"mokapi/config/dynamic/openapi/ref"
	"mokapi/config/dynamic/openapi/schema"
	"mokapi/media"
	"strconv"
	"strings"
)

func Convert(config *Config) (*openapi.Config, error) {
	result := &openapi.Config{
		OpenApi: "3.0.1",
		Info:    config.Info,
		Paths:   openapi.EndpointsRef{Value: make(map[string]*openapi.EndpointRef)},
	}

	if len(config.Schemes) == 0 {
		server := fmt.Sprintf("http://%v%v", config.Host, config.BasePath)
		result.Servers = append(result.Servers, &openapi.Server{Url: server})
	}
	for _, scheme := range config.Schemes {
		server := fmt.Sprintf("%v://%v%v", scheme, config.Host, config.BasePath)
		result.Servers = append(result.Servers, &openapi.Server{Url: server})
	}

	for path, item := range config.Paths {
		converted, err := convertPath(item)
		if err != nil {
			return nil, err
		}
		result.Paths.Value[path] = &openapi.EndpointRef{Value: converted}
	}

	if len(config.Definitions) > 0 {
		result.Components.Schemas = &schema.SchemasRef{Value: &schema.Schemas{}}
		for k, v := range config.Definitions {
			result.Components.Schemas.Value.Set(k, convertSchema(v))
		}
	}

	return result, nil
}

func convertPath(p *PathItem) (*openapi.Endpoint, error) {
	result := &openapi.Endpoint{}
	for m, o := range p.Operations() {
		converted, err := convertOperation(o)
		if err != nil {
			return nil, err
		}
		result.SetOperation(m, converted)
	}
	return result, nil
}

func convertOperation(o *Operation) (*openapi.Operation, error) {
	if o == nil {
		return nil, nil
	}
	result := &openapi.Operation{
		Tags:        o.Tags,
		Summary:     o.Summary,
		Description: o.Description,
		OperationId: o.OperationID,
		Responses:   &openapi.Responses{},
	}

	for _, p := range o.Parameters {
		switch p.In {
		case "body":
			body := &openapi.RequestBody{
				Description: p.Description,
				Required:    false,
			}
			for _, consume := range o.Consumes {
				body.Content[consume] = &openapi.MediaType{
					Schema:      p.Schema,
					ContentType: media.ParseContentType(consume),
				}
			}
		default:
			result.Parameters = append(result.Parameters, &parameter.Ref{Value: &parameter.Parameter{
				Name: p.Name,
				Type: parameter.Location(p.In),
				Schema: &schema.Ref{Value: &schema.Schema{
					Type:             p.Type,
					Format:           p.Format,
					Pattern:          p.Pattern,
					Items:            p.Items,
					Minimum:          p.Minimum,
					Maximum:          p.Maximum,
					ExclusiveMinimum: &p.ExclusiveMin,
					ExclusiveMaximum: &p.ExclusiveMax,
					UniqueItems:      p.UniqueItems,
					MinItems:         &p.MinItems,
					MaxItems:         p.MaxItems,
				}},
				Required:    p.Required,
				Description: p.Description,
			}})
		}

	}

	for statusCode, r := range o.Responses {
		converted, err := convertResponse(r, o.Produces)
		if err != nil {
			return nil, err
		}
		i, err := strconv.Atoi(statusCode)
		if err != nil {
			return nil, fmt.Errorf("status code %v is not a valid integer", statusCode)
		}
		result.Responses.Set(i, converted)
	}

	return result, nil
}

func convertResponse(r *Response, produces []string) (*openapi.ResponseRef, error) {
	if len(r.Ref) > 0 {
		return &openapi.ResponseRef{Reference: ref.Reference{Ref: convertRef(r.Ref)}}, nil
	}
	result := &openapi.Response{
		Description: r.Description,
		Content:     make(map[string]*openapi.MediaType),
	}

	for _, produce := range produces {
		result.Content[produce] = &openapi.MediaType{
			Schema:      convertSchema(r.Schema),
			ContentType: media.ParseContentType(produce),
		}
	}
	return &openapi.ResponseRef{Value: result}, nil
}

func convertSchema(s *schema.Ref) *schema.Ref {
	if len(s.Ref) > 0 {
		return &schema.Ref{Reference: ref.Reference{Ref: convertRef(s.Ref)}}
	}
	if s.Value == nil {
		return s
	}

	if s.Value.Items != nil {
		s.Value.Items = convertSchema(s.Value.Items)
	}

	if s.Value.Properties != nil {
		if len(s.Value.Properties.Ref) > 0 {
			s.Value.Properties.Ref = convertRef(s.Ref)
		} else {
			for it := s.Value.Properties.Value.Iter(); it.Next(); {
				s.Value.Properties.Value.Set(it.Key(), convertSchema(it.Value().(*schema.Ref)))
			}
		}
	}

	if s.Value.AdditionalProperties != nil && len(s.Value.AdditionalProperties.Ref.Ref) > 0 {
		s.Value.AdditionalProperties.Ref.Ref = convertRef(s.Value.AdditionalProperties.Ref.Ref)
	}
	for i, v := range s.Value.AllOf {
		s.Value.AllOf[i] = convertSchema(v)
	}
	return s
}

var refMappings = map[string]string{
	"#/definitions/": "#/components/schemas/",
	"#/responses/":   "#/components/responses/",
	"#/parameters/":  "#/components/parameters/",
}

func convertRef(ref string) string {
	for old, new := range refMappings {
		if strings.HasPrefix(ref, old) {
			ref = strings.Replace(ref, old, new, 1)
		}
	}
	return ref
}
