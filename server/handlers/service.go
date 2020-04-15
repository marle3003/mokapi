package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ServiceHandler struct {
	handlers map[string]http.Handler
}

func NewServiceHandler() *ServiceHandler {
	return &ServiceHandler{handlers: make(map[string]http.Handler)}
}

func (s *ServiceHandler) AddHandler(path string, handler http.Handler) error {
	if _, ok := s.handlers[path]; ok {
		return fmt.Errorf("Endpoint is already defined '%v'", path)
	}

	s.handlers[path] = handler
	return nil
}

func (s *ServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, error := s.resolveEndpoint(r.URL)
	if error != nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, error.Error())
		return
	}

	handler.ServeHTTP(w, r)
}

func (s *ServiceHandler) resolveEndpoint(u *url.URL) (http.Handler, error) {
	for path, handler := range s.handlers {
		if strings.HasPrefix(u.Path, path) {
			return handler, nil
		}
	}

	return nil, fmt.Errorf("There was no endpoint listening at %s", u)
}
