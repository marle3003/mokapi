package service

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/providers/parser"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ServiceContext struct {
	schemas map[string]*dynamic.Schema
	path    string
}

func CreateService(config *dynamic.OpenApi) *Service {
	service := &Service{Servers: make([]Server, 0), Endpoint: make(map[string]*Endpoint)}

	serverUrls := make(map[string]bool)
	context := &ServiceContext{schemas: getSchemas(config)}

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
				server := createServers(v)
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

	return service
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
		i, error := strconv.Atoi(k)
		if error != nil {
			log.Error("Unsupport status code", k)
			continue
		}
		status := HttpStatus(i)
		if !IsValidHttpStatus(status) {
			log.Error("Unsupport status code", k)
		}

		response := &Response{Description: v.Description, ContentTypes: make(map[ContentType]*ResponseContent)}
		o.Responses[status] = response

		for t, c := range v.Content {
			responseContent := &ResponseContent{Schema: createSchema(c.Schema, context)}
			contentType, error := ParseContentType(t)
			if error != nil {
				log.Error("Error in parsing content type", error)
				continue
			}
			response.ContentTypes[contentType] = responseContent
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
				continue
			}
			o.Parameters = append(o.Parameters, p)
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

	if config.Resource != nil {
		r := &Resource{Name: config.Resource.Name}
		expr, error := parser.ParseFilter(config.Resource.Filter)
		if error != nil {
			log.Errorf("Error in parsing filter: %v", error.Error())
		}
		r.Filter = expr
		schema.Resource = r
	}

	if config.Reference != "" {
		if strings.HasPrefix(config.Reference, "#/components/schemas/") {
			key := strings.TrimPrefix(config.Reference, "#/components/schemas/")
			return createSchema(context.schemas[key], context)
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

func createServers(config *dynamic.Server) Server {
	return Server{Host: getHost(config), Port: getPort(config), Path: getPath(config), Description: config.Description}
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

func getHost(s *dynamic.Server) string {
	u, error := url.Parse(s.Url)
	if error != nil {
		log.WithField("url", s.Url).Error("Invalid format in url found.")
		return ""
	}
	return u.Hostname()
}

func getPath(s *dynamic.Server) string {
	u, error := url.Parse(s.Url)
	if error != nil {
		log.WithField("url", s.Url).Error("Invalid format in url found.")
		return ""
	}
	if len(u.Path) == 0 {
		return "/"
	}
	return u.Path
}

func getPort(s *dynamic.Server) int {
	u, error := url.Parse(s.Url)
	if error != nil {
		log.WithField("url", s.Url).Error("Invalid format in url found.")
		return -1
	}
	portString := u.Port()
	if len(portString) == 0 {
		return 80
	} else {
		port, error := strconv.ParseInt(portString, 10, 32)
		if error != nil {
			log.WithField("url", s.Url).Error("Invalid port format in url found.")
		}
		return int(port)
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
