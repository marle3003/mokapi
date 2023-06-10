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

type converter struct {
	config *Config
}

func Convert(config *Config) (*openapi.Config, error) {
	c := &converter{config: config}
	return c.Convert()
}

func (c *converter) Convert() (*openapi.Config, error) {
	result := &openapi.Config{
		OpenApi: "3.0.1",
		Info:    c.config.Info,
		Paths:   openapi.EndpointsRef{Value: make(map[string]*openapi.EndpointRef)},
	}

	if len(c.config.Schemes) == 0 {
		if len(c.config.Host) == 0 {
			result.Servers = append(result.Servers, &openapi.Server{Url: c.config.BasePath})
		} else {
			result.Servers = append(result.Servers, &openapi.Server{Url: fmt.Sprintf("http://%v%v", c.config.Host, c.config.BasePath)})
		}
	}
	for _, scheme := range c.config.Schemes {
		server := fmt.Sprintf("%v://%v%v", scheme, c.config.Host, c.config.BasePath)
		result.Servers = append(result.Servers, &openapi.Server{Url: server})
	}

	for path, item := range c.config.Paths {
		converted, err := c.convertPath(item)
		if err != nil {
			return nil, err
		}
		result.Paths.Value[path] = converted
	}

	if len(c.config.Definitions) > 0 {
		result.Components.Schemas = &schema.Schemas{}
		for k, v := range c.config.Definitions {
			result.Components.Schemas.Set(k, c.convertSchema(v))
		}
	}

	return result, nil
}

func (c *converter) convertPath(p *PathItem) (*openapi.EndpointRef, error) {
	if len(p.Ref) > 0 {
		return &openapi.EndpointRef{Reference: ref.Reference{Ref: convertRef(p.Ref)}}, nil
	}

	result := &openapi.Endpoint{}

	var body *openapi.RequestBodyRef
	var bodySchema *schema.Ref
	for _, p := range p.Parameters {
		switch p.In {
		case "body":
			body = &openapi.RequestBodyRef{Value: &openapi.RequestBody{
				Description: p.Description,
				Required:    p.Required,
				Content:     make(map[string]*openapi.MediaType),
			}}
			bodySchema = c.convertSchema(p.Schema)
		default:
			result.Parameters = append(result.Parameters, convertParameter(p))
		}
	}

	for m, o := range p.Operations() {
		converted, err := c.convertOperation(o)
		if err != nil {
			return nil, err
		}
		if body != nil && converted.RequestBody == nil {
			for _, consume := range o.Consumes {
				b := &openapi.RequestBodyRef{Value: &openapi.RequestBody{
					Description: body.Value.Description,
					Required:    body.Value.Required,
					Content:     make(map[string]*openapi.MediaType),
				}}
				b.Value.Content[consume] = &openapi.MediaType{
					Schema:      bodySchema,
					ContentType: media.ParseContentType(consume),
				}
				converted.RequestBody = b
			}
		}
		result.SetOperation(m, converted)
	}

	return &openapi.EndpointRef{Value: result}, nil
}

func (c *converter) convertOperation(o *Operation) (*openapi.Operation, error) {
	if o == nil {
		return nil, nil
	}
	result := &openapi.Operation{
		Tags:        o.Tags,
		Summary:     o.Summary,
		Description: o.Description,
		OperationId: o.OperationID,
		Responses:   &openapi.Responses{},
		Deprecated:  o.Deprecated,
	}

	if o.Consumes == nil {
		o.Consumes = c.config.Consumes
	}

	if len(o.Consumes) == 0 {
		o.Consumes = append(o.Consumes, "*/*")
	}

	for _, p := range o.Parameters {
		switch p.In {
		case "body":
			body := &openapi.RequestBody{
				Description: p.Description,
				Required:    p.Required,
				Content:     make(map[string]*openapi.MediaType),
			}
			for _, consume := range o.Consumes {
				body.Content[consume] = &openapi.MediaType{
					Schema:      c.convertSchema(p.Schema),
					ContentType: media.ParseContentType(consume),
				}
			}
			result.RequestBody = &openapi.RequestBodyRef{Value: body}
		default:
			result.Parameters = append(result.Parameters, convertParameter(p))
		}
	}

	for statusCode, r := range o.Responses {
		converted, err := c.convertResponse(r, o.Produces)
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

func (c *converter) convertResponse(r *Response, produces []string) (*openapi.ResponseRef, error) {
	if len(r.Ref) > 0 {
		return &openapi.ResponseRef{Reference: ref.Reference{Ref: convertRef(r.Ref)}}, nil
	}
	result := &openapi.Response{
		Description: r.Description,
		Content:     make(map[string]*openapi.MediaType),
	}

	for _, produce := range produces {
		result.Content[produce] = &openapi.MediaType{
			Schema:      c.convertSchema(r.Schema),
			ContentType: media.ParseContentType(produce),
		}
	}
	return &openapi.ResponseRef{Value: result}, nil
}

func (c *converter) convertSchema(s *schema.Ref) *schema.Ref {
	if s == nil {
		return nil
	}

	if len(s.Ref) > 0 {
		return &schema.Ref{Reference: ref.Reference{Ref: convertRef(s.Ref)}}
	}

	if s.Value == nil {
		return s
	}

	if s.Value.Items != nil {
		s.Value.Items = c.convertSchema(s.Value.Items)
	}

	if s.Value.Properties != nil {
		if len(s.Value.Properties.Ref) > 0 {
			s.Value.Properties.Ref = convertRef(s.Ref)
		} else {
			for it := s.Value.Properties.Value.Iter(); it.Next(); {
				s.Value.Properties.Value.Set(it.Key(), c.convertSchema(it.Value().(*schema.Ref)))
			}
		}
	}

	if s.Value.AdditionalProperties != nil && s.Value.AdditionalProperties.Ref != nil {
		s.Value.AdditionalProperties.Ref.Ref = convertRef(s.Value.AdditionalProperties.Ref.Ref)
	}
	for i, v := range s.Value.AllOf {
		s.Value.AllOf[i] = c.convertSchema(v)
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

func convertParameter(p *Parameter) *parameter.Ref {
	return &parameter.Ref{Value: &parameter.Parameter{
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
		Deprecated:  p.Deprecated,
		Description: p.Description,
	}}
}
