package handlers

import (
	"fmt"
	"strings"
)

type ServiceHandler struct {
	handlers map[string]*EndpointHandler
}

func NewServiceHandler() *ServiceHandler {
	return &ServiceHandler{handlers: make(map[string]*EndpointHandler)}
}

func (s *ServiceHandler) AddHandler(path string, handler *EndpointHandler) error {
	if _, ok := s.handlers[path]; ok {
		return fmt.Errorf("Endpoint is already defined '%v'", path)
	}

	s.handlers[path] = handler
	return nil
}

func (s *ServiceHandler) ServeHTTP(context *Context) {
	handler, error := s.resolveEndpoint(context)
	if error != nil {
		context.Response.WriteHeader(404)
		fmt.Fprintf(context.Response, error.Error())
		return
	}

	handler.ServeHTTP(context)
}

func (s *ServiceHandler) resolveEndpoint(context *Context) (*EndpointHandler, error) {
	endpointPath := context.Request.URL.Path
	if context.ServiceUrl != "/" {
		endpointPath = context.Request.URL.Path[len(context.ServiceUrl):]
	}
	for path, handler := range s.handlers {
		if strings.HasPrefix(endpointPath, path) {
			return handler, nil
		}
	}

	return nil, fmt.Errorf("There was no endpoint listening at %s", context.Request.URL)
}
