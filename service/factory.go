package service

import (
	"mokapi/config/dynamic"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ServiceContext struct {
	schemas map[string]*dynamic.Schema
}

func CreateService(config *dynamic.OpenApi) *Service {
	service := &Service{Servers: make([]Server, 0), Endpoint: make(map[string]*Endpoint)}

	serverUrls := make(map[string]bool)
	context := &ServiceContext{schemas: getSchemas(config)}

	for _, part := range config.Parts {
		if len(service.Name) == 0 {
			service.Name = part.Info.Name
		}
		if part.Info.Description != "" {
			service.Description = part.Info.Description
		}
		if part.Info.Version != "" {
			service.Version = part.Info.Version
		}

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
				endpoint = &Endpoint{}
				service.Endpoint[k] = endpoint
			}
			endpoint.update(v, context)

		}

		dataProviders := part.Info.ServerConfiguration.DataProviders

		if dataProviders != nil {
			if dataProviders.File != nil {
				provider := &FileDataProvider{}
				if len(dataProviders.File.Filename) > 0 {
					provider.Path = dataProviders.File.Filename
				} else {
					provider.Path = dataProviders.File.Directory
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
		e.Get = createOperation(config.Post, context)
	}
	if config.Put != nil {
		e.Get = createOperation(config.Put, context)
	}
	if config.Patch != nil {
		e.Get = createOperation(config.Patch, context)
	}
	if config.Delete != nil {
		e.Get = createOperation(config.Delete, context)
	}
	if config.Head != nil {
		e.Get = createOperation(config.Head, context)
	}
	if config.Options != nil {
		e.Get = createOperation(config.Options, context)
	}
	if config.Trace != nil {
		e.Get = createOperation(config.Trace, context)
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

	return o
}

func createSchema(config *dynamic.Schema, context *ServiceContext) *Schema {
	// todo resolving schema: $ref
	schema := &Schema{
		Description: config.Description,
		Faker:       config.Faker,
		Format:      config.Format,
		Resource:    config.Resource,
		Type:        config.Type,
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
