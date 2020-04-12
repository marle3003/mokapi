package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

type EntryPointHandler struct {
	services map[string]*ServiceHandler
}

func NewEntryPointHandler() *EntryPointHandler {
	handler := &EntryPointHandler{services: make(map[string]*ServiceHandler)}

	return handler
}

func (e *EntryPointHandler) AddService(path string, service *ServiceHandler) error {
	if _, ok := e.services[path]; ok {
		return fmt.Errorf("Already service defined on path '%v'", path)
	}

	e.services[path] = service
	return nil
}

func (e *EntryPointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for path, service := range e.services {
		if strings.HasPrefix(r.URL.Path, path) {
			service.ServeHTTP(w, r)
			return
		}
	}

	w.WriteHeader(404)
	fmt.Fprintf(w, "No service found %v", r.URL.String())
}
