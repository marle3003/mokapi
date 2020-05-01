package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

type EntryPointHandler struct {
	handlers map[string]*ServiceHandler
	Host     string
	Port     int
}

func NewEntryPointHandler(host string, port int) *EntryPointHandler {
	return &EntryPointHandler{handlers: make(map[string]*ServiceHandler), Host: host, Port: port}
}

func (e *EntryPointHandler) AddHandler(path string, handler *ServiceHandler) error {
	if _, ok := e.handlers[path]; ok {
		return fmt.Errorf("A service is already defined on path '%v'", path)
	}

	e.handlers[path] = handler
	return nil
}

func (e *EntryPointHandler) RemoveHandler(path string) {
	delete(e.handlers, path)
}

func (e *EntryPointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service, path := e.resolveService(r.URL)
	if service == nil {
		w.WriteHeader(404)
		log.Errorf("There was no service listening at %v", r.URL)
		return
	}

	service.ServeHTTP(NewContext(path, w, r))
}

func (e *EntryPointHandler) resolveService(u *url.URL) (*ServiceHandler, string) {
	var matchedPath string
	var matchedHandler *ServiceHandler
	for path, handler := range e.handlers {
		if strings.HasPrefix(u.Path, path) {
			if matchedPath == "" || len(matchedPath) < len(path) {
				matchedPath = path
				matchedHandler = handler
			}
		}
	}

	if matchedHandler != nil {
		return matchedHandler, matchedPath
	}

	return nil, ""
}
