package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type EntryPointHandler struct {
	handlers map[string]http.Handler
}

func NewEntryPointHandler() *EntryPointHandler {
	return &EntryPointHandler{handlers: make(map[string]http.Handler)}
}

func (e *EntryPointHandler) AddHandler(path string, handler http.Handler) error {
	if _, ok := e.handlers[path]; ok {
		return fmt.Errorf("A service is already defined on path '%v'", path)
	}

	e.handlers[path] = handler
	return nil
}

func (e *EntryPointHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service, error := e.resolveService(r.URL)
	if error != nil {
		w.WriteHeader(404)
		fmt.Fprint(w, error.Error())
		return
	}

	service.ServeHTTP(w, r)
}

func (e *EntryPointHandler) resolveService(u *url.URL) (http.Handler, error) {
	for path, handler := range e.handlers {
		if strings.HasPrefix(u.Path, path) {
			return handler, nil
		}
	}

	return nil, fmt.Errorf("There was no service listening at %v", u)
}
