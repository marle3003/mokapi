package handlers

import (
	"fmt"
	"mokapi/config"
	"net/http"
)

type EndpointHandler struct {
	handlers map[string]*OperationHandler
}

func NewEndpointHandler(endpoint *config.Endpoint) *EndpointHandler {
	handler := &EndpointHandler{}
	handler.setOperations(endpoint)
	return handler
}

func (e *EndpointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := e.handlers[r.Method]; ok {
		handler.ServeHTTP(w, r)
	} else {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Method %s on endpoint %v not found", r.Method, r.URL.String())
	}
}

func (e *EndpointHandler) setOperations(endpoint *config.Endpoint) {
	e.handlers = make(map[string]*OperationHandler)

	if endpoint.Get != nil {
		e.handlers["GET"] = NewOperationHandler(endpoint.Get)
	}
}
