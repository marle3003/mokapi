package handlers

import (
	"fmt"
	"mokapi/config"
	"net/http"
)

type EndpointHandler struct {
	handlers map[string]http.Handler
}

func NewEndpointHandler(endpoint *config.Endpoint) *EndpointHandler {
	return &EndpointHandler{handlers: make(map[string]http.Handler)}
}

func (e *EndpointHandler) AddHandler(method string, handler http.Handler) {
	e.handlers[method] = handler
}

func (e *EndpointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := e.handlers[r.Method]; ok {
		handler.ServeHTTP(w, r)
	} else {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Method %s on endpoint %v not found", r.Method, r.URL.String())
	}
}
