package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

type OperationHandler struct {
	handlers map[string]http.Handler
}

func NewOperationHandler() *OperationHandler {
	return &OperationHandler{handlers: make(map[string]http.Handler)}
}

func (o *OperationHandler) AddHandler(contentType string, handler http.Handler) {
	o.handlers[contentType] = handler
}

func (o *OperationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, error := o.resolveHandler(r.Header.Get("accept"))
	if error != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "No supporting content type found")
		return
	}

	handler.ServeHTTP(w, r)
}

func (o *OperationHandler) resolveHandler(accept string) (http.Handler, error) {
	for _, s := range strings.Split(accept, ",") {
		contentType := strings.TrimSpace(s)
		if handler, ok := o.handlers[contentType]; ok {
			return handler, nil
		}
	}
	return nil, fmt.Errorf("No supporting content type found")
}
