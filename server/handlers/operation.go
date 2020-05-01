package handlers

import (
	"fmt"
	"mokapi/service"
	"strings"
)

type OperationHandler struct {
	handlers map[string]*ResponseHandler
}

func NewOperationHandler() *OperationHandler {
	return &OperationHandler{handlers: make(map[string]*ResponseHandler)}
}

func (o *OperationHandler) AddHandler(contentType service.ContentType, handler *ResponseHandler) {
	o.handlers[contentType.String()] = handler
}

func (o *OperationHandler) ServeHTTP(context *Context) {
	handler, error := o.resolveHandler(context.Request.Header.Get("accept"))
	if error != nil {
		context.Response.WriteHeader(500)
		fmt.Fprintf(context.Response, "No supporting content type found")
		return
	}

	handler.ServeHTTP(context)
}

func (o *OperationHandler) resolveHandler(accept string) (*ResponseHandler, error) {
	if accept == "" {
		for _, handler := range o.handlers {
			// return first element
			return handler, nil
		}
	}
	for _, s := range strings.Split(accept, ",") {
		contentType := strings.TrimSpace(s)
		if handler, ok := o.handlers[contentType]; ok {
			return handler, nil
		}
	}
	return nil, fmt.Errorf("No supporting content type found")
}
