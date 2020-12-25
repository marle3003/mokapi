package server

import (
	"context"
	"fmt"
	"mokapi/models"
	"mokapi/server/web"
	"mokapi/server/web/handlers"
	"strings"

	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type HttpBinding struct {
	Address  string
	Router   *mux.Router
	server   *http.Server
	handlers map[string]*handlers.WebServiceHandler
}

func NewHttpBinding(address string) *HttpBinding {
	router := mux.NewRouter()
	server := &http.Server{Addr: address, Handler: router}
	httpBinding := &HttpBinding{
		Address:  address,
		Router:   router,
		server:   server,
		handlers: make(map[string]*handlers.WebServiceHandler),
	}

	router.Host(GetHost(address)).Subrouter().NewRoute().Handler(httpBinding)
	//router.NotFoundHandler

	return httpBinding
}

func (binding *HttpBinding) Start() {
	go func() {
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
		return fmt.Errorf("Unexpected parameter type %T in http binding", data)
	}

	ok, path := binding.getServicePath(service)
	if !ok {
		return fmt.Errorf("No matching address (%v) found in service", binding.Address)
	}

	if handler, found := binding.handlers[path]; found {
		if service.Name != handler.WebService.Name {
			return fmt.Errorf("The service '%v' is already defined on path '%v'", handler.WebService.Name, path)
		} else {
			return nil
		}
	}

	log.Infof("Adding service %v at address %v on path %v", service.Name, binding.Address, path)

	handler := handlers.NewWebServiceHandler(service)
	binding.handlers[path] = handler

	return nil
}

func (binding *HttpBinding) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service, servicePath := binding.resolveHandler(r.URL)
	if service == nil {
		m := fmt.Sprintf("There was no service listening at %v", r.URL)
		http.Error(w, m, http.StatusInternalServerError)
		log.Error(m)
		//e.requestChannel <- &models.RequestMetric{Method: r.Method, Url: r.URL.String(), Error: m, HttpStatus: http.StatusInternalServerError}
		return
	}

	service.ServeHTTP(web.NewHttpContext(r, w, servicePath))
}

func (binding *HttpBinding) resolveHandler(u *url.URL) (*handlers.WebServiceHandler, string) {
	var matchedPath string
	var matchedHandler *handlers.WebServiceHandler
	for path, handler := range binding.handlers {
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
