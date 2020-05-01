package handlers

import (
	"fmt"
)

type EndpointHandler struct {
	handlers map[string]*OperationHandler
}

func NewEndpointHandler() *EndpointHandler {
	return &EndpointHandler{handlers: make(map[string]*OperationHandler)}
}

func (e *EndpointHandler) AddHandler(method string, handler *OperationHandler) {
	e.handlers[method] = handler
}

func (e *EndpointHandler) ServeHTTP(context *Context) {
	if handler, ok := e.handlers[context.Request.Method]; ok {
		handler.ServeHTTP(context)
	} else {
		context.Response.WriteHeader(404)
		fmt.Fprintf(context.Response, "Method %s on endpoint %v not found", context.Request.Method, context.Request.URL.String())
	}
}
