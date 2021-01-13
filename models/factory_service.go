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
		webServiceInfo.apply(item, filePath)
	}
}

func (w *WebServiceInfo) apply(config *dynamic.OpenApi, filePath string) {
	context := &serviceContext{config: config, Errors: make([]string, 0), path: filePath}

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
		server, err := createServers(v)
		if err != nil {
			log.Error(err.Error())
			context.Errors = append(context.Errors, err.Error())
			continue
		}
		w.Data.AddServer(server)
	}

	for path, v := range config.EndPoints {
		var endpoint *Endpoint
		if e, ok := w.Data.Endpoint[path]; ok {
			endpoint = e
		} else {
			endpoint = NewEndpoint(path, w.Data)
			endpoint.Pipeline = v.Pipeline
			w.Data.Endpoint[path] = endpoint
		}
		endpoint.update(v, context)

	}

	if len(config.Info.MokapiFile) > 0 {
		mokapiFile := config.Info.MokapiFile
		if strings.HasPrefix(mokapiFile, "./") {
			dir := filepath.Dir(context.path)
			mokapiFile = strings.Replace(mokapiFile, ".", dir, 1)
		} else if !filepath.IsAbs(mokapiFile) {
			dir := filepath.Dir(context.path)
			mokapiFile = filepath.Join(dir, mokapiFile)
		}
		w.Data.MokapiFile = mokapiFile
	}

	w.Errors = append(w.Errors, context.Errors...)
	for _, err := range context.Errors {
		log.Errorf("error in config %v: %v", filePath, err)
	}
}

func (e *Endpoint) update(config *dynamic.Endpoint, context *serviceContext) {
	if len(config.Summary) > 0 {
		e.Summary = config.Summary
	}
	if len(config.Description) > 0 {
		e.Description = config.Description
	}

	if config.Get != nil {
		e.Get = createOperation(config.Get, e, context)
	}
	if config.Post != nil {
		e.Post = createOperation(config.Post, e, context)
	}
	if config.Put != nil {
		e.Put = createOperation(config.Put, e, context)
	}
	if config.Patch != nil {
		e.Patch = createOperation(config.Patch, e, context)
	}
	if config.Delete != nil {
		e.Delete = createOperation(config.Delete, e, context)
	}
	if config.Head != nil {
		e.Head = createOperation(config.Head, e, context)
	}
	if config.Options != nil {
		e.Options = createOperation(config.Options, e, context)
	}
	if config.Trace != nil {
		e.Trace = createOperation(config.Trace, e, context)
	}
	if config.Parameters != nil {
		if e.Parameters == nil {
			e.Parameters = make([]*Parameter, 0)
		}
		for _, v := range config.Parameters {
			p, err := createParameter(v, context)
			if err != nil {
				log.Error(err.Error())
				context.Errors = append(context.Errors, err.Error())
				continue
			}
			e.Parameters = append(e.Parameters, p)
		}
	}
}

func createOperation(config *dynamic.Operation, endpoint *Endpoint, context *serviceContext) *Operation {
	o := NewOperation(
		config.Summary,
		config.Description,
		config.OperationId,
		config.Pipeline,
		endpoint,
	)

	if config.RequestBody != nil {
		o.RequestBody = createRequestBody(config.RequestBody, context)
	}

	for k, r := range config.Responses {
		s, err := parseHttpStatus(k)
		if err != nil {
			log.Error(err.Error())
			context.Errors = append(context.Errors, err.Error())
			continue
		}
		o.Responses[s] = createResponse(r, context)
	}

	if config.Parameters != nil {
		if o.Parameters == nil {
			o.Parameters = make([]*Parameter, 0)
		}
		for _, v := range config.Parameters {
			p, err := createParameter(v, context)
			if err != nil {
				log.Error(err.Error())
				context.Errors = append(context.Errors, err.Error())
				continue
			}
			o.Parameters = append(o.Parameters, p)
		}
	}

	return o
}

func createRequestBody(config *dynamic.RequestBody, context *serviceContext) *RequestBody {
	if len(config.Reference) > 0 {
		if ref, ok := context.getRequestBody(config.Reference); ok {
			return createRequestBody(ref, context)
		} else {
			context.Errors = append(context.Errors, fmt.Sprintf("unable to resolve reference '%v'", config.Reference))
			return nil
		}
	}

	r := &RequestBody{Description: config.Description, ContentTypes: make(map[string]*MediaType), Required: config.Required}
	if config.Content != nil {
		for t, c := range config.Content {
			if c == nil {
				continue
			}
			m := getMediaType(c, context)
			contentType := ParseContentType(t)
			r.ContentTypes[contentType.Key()] = m
		}
	}
	return r
}

func createResponse(config *dynamic.Response, context *serviceContext) *Response {
	if len(config.Reference) > 0 {
		if ref, ok := context.getResponse(config.Reference); ok {
			return createResponse(ref, context)
		} else {
			context.Errors = append(context.Errors, fmt.Sprintf("unable to resolve reference '%v'", config.Reference))
			return nil
		}
	}

	r := &Response{Description: config.Description, ContentTypes: make(map[string]*MediaType)}
	if config.Content != nil {
		for t, c := range config.Content {
			if c == nil {
				continue
			}
			m := getMediaType(c, context)
			contentType := ParseContentType(t)
			r.ContentTypes[contentType.Key()] = m
		}
	}
	return r
}

func getMediaType(config *dynamic.MediaType, context *serviceContext) *MediaType {
	m := &MediaType{}
	if config.Schema != nil {
		m.Schema = createSchema(config.Schema, context)
	}
	return m
}

func createSchema(config *dynamic.Schema, context *serviceContext) *Schema {
	if config == nil {
		return nil
	}

	if len(config.Reference) > 0 {
		schema, ok := context.getSchema(config.Reference)
		if !ok {
			context.Errors = append(context.Errors, fmt.Sprintf("unable to resolve reference '%v'", config.Reference))
			return nil
		}
		ref := createSchema(schema, context)
		ref.Reference = config.Reference
		return ref
	}

	schema := &Schema{
		Description:          config.Description,
		Faker:                config.Faker,
		Format:               config.Format,
		Type:                 config.Type,
		AdditionalProperties: config.AdditionalProperties,
		Reference:            config.Reference,
		Required:             config.Required,
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

	if config.Items != nil {
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

func createParameter(config *dynamic.Parameter, context *serviceContext) (*Parameter, error) {
	p := &Parameter{
		Name:        config.Name,
		Description: config.Description,
		Required:    config.Required,
		Schema:      createSchema(config.Schema, context),
		Style:       config.Style,
	}

	if len(config.Explode) == 0 {
		p.Explode = true
	} else {
		b, err := strconv.ParseBool(config.Explode)
		if err != nil {
			return nil, err
		}
		p.Explode = b
	}

	switch strings.ToLower(config.Type) {
	case "path":
		p.Location = PathParameter
	case "query":
		p.Location = QueryParameter
	case "header":
		p.Location = HeaderParameter
	case "cookie":
		p.Location = CookieParameter
	default:
		return nil, fmt.Errorf("Unsupported parameter type %v", config.Type)
	}

	return p, nil
}

type serviceContext struct {
	config *dynamic.OpenApi
	path   string
	Errors []string
}

func (c *serviceContext) getRequestBody(ref string) (*dynamic.RequestBody, bool) {
	if len(ref) == 0 {
		return nil, false
	}
	if ref[0] == '#' {
		key := strings.TrimPrefix(ref, "#/components/requestBodies/")
		if len(key) == 0 {
			return nil, false
		}
		s, ok := c.config.Components.RequestBodies[key]
		return s, ok
	}
	return nil, false
}

func (c *serviceContext) getSchema(ref string) (*dynamic.Schema, bool) {
	if len(ref) == 0 {
		return nil, false
	}
	if ref[0] == '#' {
		key := strings.TrimPrefix(ref, "#/components/schemas/")
		if len(key) == 0 {
			return nil, false
		}
		s, ok := c.config.Components.Schemas[key]
		return s, ok
	}
	return nil, false
}

func (c *serviceContext) getResponse(ref string) (*dynamic.Response, bool) {
	if len(ref) == 0 {
		return nil, false
	}
	if ref[0] == '#' {
		key := strings.TrimPrefix(ref, "#/components/responses/")
		if len(key) == 0 {
			return nil, false
		}
		s, ok := c.config.Components.Responses[key]
		return s, ok
	}
	return nil, false
}
