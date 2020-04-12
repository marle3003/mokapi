package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

type ServiceHandler struct {
	endpoints map[string]*EndpointHandler
}

func NewServiceHandler() *ServiceHandler {
	return &ServiceHandler{endpoints: make(map[string]*EndpointHandler)}
}

func (s *ServiceHandler) AddEndpoint(path string, endpoint *EndpointHandler) error {
	if _, ok := s.endpoints[path]; ok {
		return fmt.Errorf("Endpoint is already defined '%v'", path)
	}

	s.endpoints[path] = endpoint
	return nil
}

func (s *ServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for path, endpoint := range s.endpoints {
		if strings.HasPrefix(r.URL.Path, path) {
			endpoint.ServeHTTP(w, r)
			return
		}
	}

	w.WriteHeader(404)
	fmt.Fprintf(w, "No endpoint found %v", r.URL.String())
}
