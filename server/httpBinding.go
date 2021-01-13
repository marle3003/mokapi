package server

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"mokapi/models"
	"mokapi/server/web"
	"mokapi/server/web/handlers"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"net/http"
)

type HttpBinding struct {
	Address  string
	server   *http.Server
	handlers map[string]map[string]*handlers.WebServiceHandler
}

func NewHttpBinding(address string) *HttpBinding {
	httpBinding := &HttpBinding{
		Address:  address,
		handlers: make(map[string]map[string]*handlers.WebServiceHandler),
	}
	httpBinding.server = &http.Server{Addr: address, Handler: httpBinding}

	return httpBinding
}

func (binding *HttpBinding) Start() {
	go func() {
		log.Infof("Starting web binding %v", binding.Address)
		binding.server.ListenAndServe()
	}()
}

func (binding *HttpBinding) Stop() {
	go func() {
		log.Infof("Stopping server on %v", binding.Address)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		binding.server.SetKeepAlivesEnabled(false)
		if error := binding.server.Shutdown(ctx); error != nil {
			log.Errorf("Could not gracefully shutdown server %v", binding.Address)
		}
	}()
}

func (binding *HttpBinding) Apply(data interface{}) error {
	service, ok := data.(*models.WebService)
	if !ok {
		return errors.Errorf("unexpected parameter type %T in http binding", data)
	}

	for _, server := range service.Servers {
		address := fmt.Sprintf(":%v", server.Port)
		if binding.Address != address {
			continue
		}

		host, found := binding.handlers[server.Host]
		if !found {
			log.Infof("Adding new host '%v' on binding %v", server.Host, binding.Address)
			host = make(map[string]*handlers.WebServiceHandler)
			binding.handlers[server.Host] = host
		}

		if handler, found := host[server.Path]; found {
			if service.Name != handler.WebService.Name {
				return errors.Errorf("service '%v' is already defined on path '%v'", handler.WebService.Name, server.Path)
			}
		} else {
			log.Infof("Adding service %v on binding %v on path %v", service.Name, binding.Address, server.Path)
			handler = handlers.NewWebServiceHandler(service)
			host[server.Path] = handler
		}
	}

	return nil
}

func (binding *HttpBinding) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service, servicePath := binding.resolveHandler(r)
	if service == nil {
		m := fmt.Sprintf("There was no service listening at %v", r.URL)
		http.Error(w, m, http.StatusInternalServerError)
		log.Error(m)
		//e.requestChannel <- &models.RequestMetric{Method: r.Method, Url: r.URL.String(), Error: m, HttpStatus: http.StatusInternalServerError}
		return
	}

	service.ServeHTTP(web.NewHttpContext(r, w, servicePath))
}

func (binding *HttpBinding) resolveHandler(r *http.Request) (*handlers.WebServiceHandler, string) {
	var matchedPath string
	var matchedHandler *handlers.WebServiceHandler
	if host, ok := binding.handlers[r.Host]; ok {
		for path, handler := range host {
			if strings.HasPrefix(r.URL.Path, path) {
				if matchedPath == "" || len(matchedPath) < len(path) {
					matchedPath = path
					matchedHandler = handler
				}
			}
		}
	}

	if matchedHandler != nil {
		return matchedHandler, matchedPath
	}

	return nil, ""
}

func (binding *HttpBinding) getServicePath(service *models.WebService) (bool, string) {
	for _, server := range service.Servers {
		if fmt.Sprintf("%v:%v", server.Host, server.Port) == binding.Address {
			return true, server.Path
		}
	}
	return false, ""
}

func GetHost(s string) string {
	return strings.Split(s, ":")[0]
}
