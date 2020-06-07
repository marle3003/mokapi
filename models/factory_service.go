package models

import (
	"fmt"
	"mokapi/config/dynamic"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ServiceContext struct {
	schemas map[string]*dynamic.Schema
	path    string
	Errors  []string
}

func CreateService(config *dynamic.OpenApi) (*Service, []string) {
	service := &Service{Servers: make([]Server, 0), Endpoint: make(map[string]*Endpoint)}

	serverUrls := make(map[string]bool)
	context := &ServiceContext{schemas: getSchemas(config), Errors: make([]string, 0)}

	for filePath, part := range config.Parts {
		if len(service.Name) == 0 {
			service.Name = part.Info.Name
		}
		if part.Info.Description != "" {
			service.Description = part.Info.Description
		}
		if part.Info.Version != "" {
			service.Version = part.Info.Version
		}

		context.path = filePath

		for _, v := range part.Servers {
			if _, found := serverUrls[v.Url]; !found {
				serverUrls[v.Url] = true
				server, error := createServers(v)
				if error != nil {
					log.Error(error.Error())
					context.Errors = append(context.Errors, error.Error())
					continue
				}
				service.Servers = append(service.Servers, server)
			}

		}

		for k, v := range part.EndPoints {
			var endpoint *Endpoint
			if e, ok := service.Endpoint[k]; ok {
				endpoint = e
			} else {
				endpoint = &Endpoint{Path: k}
				service.Endpoint[k] = endpoint
			}
			endpoint.update(v, context)

		}

		dataProviders := part.Info.DataProviders

		if dataProviders != nil {
			if dataProviders.File != nil {
				provider := &FileDataProvider{}
				if len(dataProviders.File.Filename) > 0 {
					provider.Path = dataProviders.File.Filename
				} else {
					provider.Path = dataProviders.File.Directory
				}

				if strings.HasPrefix(provider.Path, "./") {
					dir := filepath.Dir(context.path)
					provider.Path = strings.Replace(provider.Path, ".", dir, 1)
				}

				service.DataProviders.File = provider
			}
		}
	}

	return service, context.Errors
}

func (e *Endpoint) update(config *dynamic.Endpoint, context *ServiceContext) {
	if config.Get != nil {
		e.Get = createOperation(config.Get, context)
	}
	if config.Post != nil {
		e.Post = createOperation(config.Post, context)
	}
	if config.Put != nil {
		e.Put = createOperation(config.Put, context)
	}
	if config.Patch != nil {
		e.Patch = createOperation(config.Patch, context)
	}
	if config.Delete != nil {
		e.Delete = createOperation(config.Delete, context)
	}
	if config.Head != nil {
		e.Head = createOperation(config.Head, context)
	}
	if config.Options != nil {
		e.Options = createOperation(config.Options, context)
	}
	if config.Trace != nil {
		e.Trace = createOperation(config.Trace, context)
	}
	if config.Parameters != nil {
		if e.Parameters == nil {
			e.Parameters = make([]*Parameter, 0)
		}
		for _, v := range config.Parameters {
			p, error := createParameter(v, context)
			if error != nil {
				log.Error(error.Error())
				continue
			}
			e.Parameters = append(e.Parameters, p)
		}
	}
}

func createOperation(config *dynamic.Operation, context *ServiceContext) *Operation {
	o := &Operation{Description: config.Description, OperationId: config.OperationId, Summary: config.Summary, Responses: make(map[HttpStatus]*Response)}

	for k, v := range config.Responses {
		status, error := parseHttpStatus(k)
		if error != nil {
			log.Error(error.Error())
			context.Errors = append(context.Errors, error.Error())
			continue
		}

		response := &Response{Description: v.Description, ContentTypes: make(map[string]*ResponseContent)}
		o.Responses[status] = response

		if v.Content != nil {
			for t, c := range v.Content {
				responseContent := &ResponseContent{}
				if c != nil && c.Schema != nil {
					responseContent.Schema = createSchema(c.Schema, context)
				}
				contentType := NewContentType(t)
				response.ContentTypes[contentType.Key()] = responseContent
			}
		}
	}

	if config.Parameters != nil {
		if o.Parameters == nil {
			o.Parameters = make([]*Parameter, 0)
		}
		for _, v := range config.Parameters {
			p, error := createParameter(v, context)
			if error != nil {
				log.Error(error.Error())
				context.Errors = append(context.Errors, error.Error())
				continue
			}
			o.Parameters = append(o.Parameters, p)
		}
	}

	if config.Middlewares != nil {
		o.Middleware = createMiddlewares(config.Middlewares, context)
	}

	if config.Resources != nil {
		o.Resources = make([]*Resource, 0)
		for _, i := range config.Resources {
			r := &Resource{Name: i.Name}
			if len(i.If) > 0 {
				r.If = NewFilter(i.If)
			}
			o.Resources = append(o.Resources, r)
		}
	}

	return o
}

func createSchema(config *dynamic.Schema, context *ServiceContext) *Schema {
	if config == nil {
		return nil
	}

	if config.Type == "array" {
		fmt.Print("")
	}

	// todo resolving schema: $ref
	schema := &Schema{
		Description:          config.Description,
		Faker:                config.Faker,
		Format:               config.Format,
		Type:                 config.Type,
		AdditionalProperties: config.AdditionalProperties,
		Reference:            config.Reference,
	}

	if config.Xml != nil {
		schema.Xml = &XmlEncoding{
			Attribute: config.Xml.Attribute,
			CData:     config.Xml.CData,
			Name:      config.Xml.Name,
			Namespace: config.Xml.Namespace,
			Prefix:    config.Xml.Prefix,
			Wrapped:   config.Xml.Wrapped,
		}
	}

	if config.Reference != "" {
		if strings.HasPrefix(config.Reference, "#/components/schemas/") {
			key := strings.TrimPrefix(config.Reference, "#/components/schemas/")
			ref := createSchema(context.schemas[key], context)
			ref.Reference = config.Reference
			return ref
		} else {
			// todo
		}
	} else if config.Items != nil {
		schema.Items = createSchema(config.Items, context)
	}

	for i, p := range config.Properties {
		if schema.Properties == nil {
			schema.Properties = make(map[string]*Schema)
		}
		schema.Properties[i] = createSchema(p, context)
	}

	return schema
}

func getSchemas(config *dynamic.OpenApi) map[string]*dynamic.Schema {
	schemas := make(map[string]*dynamic.Schema)
	for _, part := range config.Parts {
		for k, v := range part.Components.Schemas {
			schemas[k] = v
		}
	}
	return schemas
}

func createServers(config *dynamic.Server) (Server, error) {
	host, error := getHost(config)
	if error != nil {
		return Server{}, error
	}
	port, error := getPort(config)
	if error != nil {
		return Server{}, error
	}
	path, error := getPath(config)
	if error != nil {
		return Server{}, error
	}

	return Server{Host: host, Port: port, Path: path, Description: config.Description}, nil
}

func GetEndpoint(config *dynamic.Endpoint) *Endpoint {
	return &Endpoint{
		Get:     GetOperation(config.Get),
		Post:    GetOperation(config.Post),
		Put:     GetOperation(config.Put),
		Patch:   GetOperation(config.Patch),
		Delete:  GetOperation(config.Delete),
		Head:    GetOperation(config.Head),
		Options: GetOperation(config.Options),
		Trace:   GetOperation(config.Trace),
	}

}

func GetOperation(config *dynamic.Operation) *Operation {
	if config == nil {
		return nil
	}
	return &Operation{Description: config.Description, OperationId: config.OperationId, Summary: config.Summary}
}

func getHost(s *dynamic.Server) (string, error) {
	u, error := url.Parse(s.Url)
	if error != nil {
		return "", fmt.Errorf("Invalid format in url found: %v", s.Url)
	}
	return u.Hostname(), nil
}

func getPath(s *dynamic.Server) (string, error) {
	u, error := url.Parse(s.Url)
	if error != nil {
		return "", fmt.Errorf("Invalid format in url found: %v", s.Url)
	}
	if len(u.Path) == 0 {
		return "/", nil
	}
	return u.Path, nil
}

func getPort(s *dynamic.Server) (int, error) {
	u, error := url.Parse(s.Url)
	if error != nil {
		return -1, fmt.Errorf("Invalid format in url found: %v", s.Url)
	}
	portString := u.Port()
	if len(portString) == 0 {
		return 80, nil
	} else {
		port, error := strconv.ParseInt(portString, 10, 32)
		if error != nil {
			return -1, fmt.Errorf("Invalid port format in url found: %v", error.Error())
		}
		return int(port), nil
	}
}

func createParameter(config *dynamic.Parameter, context *ServiceContext) (*Parameter, error) {
	p := &Parameter{Name: config.Name, Description: config.Description, Required: config.Required, Schema: createSchema(config.Schema, context)}

	switch strings.ToLower(config.Type) {
	case "path":
		p.Type = PathParameter
	case "query":
		p.Type = QueryParameter
	case "header":
		p.Type = HeaderParameter
	case "cookie":
		p.Type = CookieParameter
	default:
		return nil, fmt.Errorf("Unsupported parameter type %v", config.Type)
	}

	return p, nil
}

func createMiddlewares(config []map[string]interface{}, context *ServiceContext) []interface{} {
	middlewares := make([]interface{}, 0)
	for _, c := range config {
		if t, ok := c["type"]; ok {
			if s, ok := t.(string); ok {
				switch strings.ToLower(s) {
				case "replacecontent":
					m := &ReplaceContent{}
					if s, ok := c["regex"].(string); ok {
						m.Regex = s
					}
					if replacement, ok := c["replacement"].(map[interface{}]interface{}); ok {
						if s, ok := replacement["from"].(string); ok {
							m.Replacement.From = s
						}
						if s, ok := replacement["selector"].(string); ok {
							m.Replacement.Selector = s
						}
					}
					middlewares = append(middlewares, m)
				case "filtercontent":
					m := &FilterContent{}
					if s, ok := c["filter"].(string); ok {
						m.Filter = NewFilter(s)
					}
					middlewares = append(middlewares, m)
				case "template":
					m := &Template{}
					if s, ok := c["filename"].(string); ok {
						if strings.HasPrefix(s, "./") {
							dir := filepath.Dir(context.path)
							s = strings.Replace(s, ".", dir, 1)
						}
						m.Filename = s
					}
					middlewares = append(middlewares, m)
				case "selection":
					m := &Selection{}
					if slice, ok := c["slice"].(map[interface{}]interface{}); ok {
						m.Slice = &Slice{Low: 0, High: -1}
						if s, ok := slice["low"].(int); ok {
							m.Slice.Low = s
						}
						if s, ok := slice["high"].(int); ok {
							m.Slice.High = s
						}
					}
					if b, ok := c["first"].(bool); ok {
						m.First = b
					}
					middlewares = append(middlewares, m)
				default:
					error := fmt.Errorf("Unsupported middleware %v", s)
					log.Error(error.Error())
					context.Errors = append(context.Errors, error.Error())

				}
				continue
			}

		}
		error := fmt.Errorf("No type definition found in middleware configuration")
		log.Error(error.Error())
		context.Errors = append(context.Errors, error.Error())
	}

	return middlewares
}
