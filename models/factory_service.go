package models

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

type ServiceContext struct {
	schemas map[string]*dynamic.Schema
	path    string
	Errors  []string
}

func (a *Application) ApplyWebService(config map[string]*dynamic.OpenApi) {
	for filePath, item := range config {
		key := filePath
		if len(item.Info.Name) > 0 {
			key = item.Info.Name
		}
		webServiceInfo, found := a.WebServices[key]
		if !found {
			webServiceInfo = NewServiceInfo()
			a.WebServices[key] = webServiceInfo
		}
		webServiceInfo.Apply(item, filePath)
	}
}

func (w *WebServiceInfo) Apply(config *dynamic.OpenApi, filePath string) {

	serverUrls := make(map[string]bool)
	context := &ServiceContext{schemas: getSchemas(config), Errors: make([]string, 0), path: filePath}

	if len(w.Data.Name) == 0 {
		w.Data.Name = config.Info.Name
	}
	if config.Info.Description != "" {
		w.Data.Description = config.Info.Description
	}
	if config.Info.Version != "" {
		w.Data.Version = config.Info.Version
	}

	for _, v := range config.Servers {
		if _, found := serverUrls[v.Url]; !found {
			serverUrls[v.Url] = true
			server, error := createServers(v)
			if error != nil {
				log.Error(error.Error())
				context.Errors = append(context.Errors, error.Error())
				continue
			}
			w.Data.Servers = append(w.Data.Servers, server)
		}

	}

	for path, v := range config.EndPoints {
		var endpoint *Endpoint
		if e, ok := w.Data.Endpoint[path]; ok {
			endpoint = e
		} else {
			endpoint = &Endpoint{Path: path}
			w.Data.Endpoint[path] = endpoint
		}
		endpoint.update(v, context)

	}

	if len(config.Info.MokapiFile) > 0 {
		mokapiFile := config.Info.MokapiFile
		if strings.HasPrefix(mokapiFile, "./") {
			dir := filepath.Dir(context.path)
			mokapiFile = strings.Replace(mokapiFile, ".", dir, 1)
		}
		w.Data.MokapiFile = mokapiFile
	}

	w.Errors = append(w.Errors, context.Errors...)
}

func (e *Endpoint) update(config *dynamic.Endpoint, context *ServiceContext) {
	if len(config.Summary) > 0 {
		e.Summary = config.Summary
	}
	if len(config.Description) > 0 {
		e.Description = config.Description
	}

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
	o := &Operation{Description: config.Description,
		OperationId: config.OperationId,
		Summary:     config.Summary,
		Responses:   make(map[HttpStatus]*Response),
		Pipeline:    config.Pipeline,
	}

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
				contentType := ParseContentType(t)
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
	for k, v := range config.Components.Schemas {
		schemas[k] = v
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
