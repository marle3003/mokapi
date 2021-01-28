package web

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"mokapi/models"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"net/http"
)

type AddRequestMetric func(metric *models.RequestMetric)

type Binding struct {
	Addr     string
	server   *http.Server
	handlers map[string]map[string]*ServiceHandler
	mh       AddRequestMetric
}

func NewBinding(addr string, mh AddRequestMetric) *Binding {
	b := &Binding{
		Addr:     addr,
		handlers: make(map[string]map[string]*ServiceHandler),
		mh:       mh,
	}
	b.server = &http.Server{Addr: addr, Handler: b}

	return b
}

func (binding *Binding) Start() {
	go func() {
		log.Infof("Starting web binding %v", binding.Addr)
		binding.server.ListenAndServe()
	}()
}

func (binding *Binding) Stop() {
	go func() {
		log.Infof("Stopping server on %v", binding.Addr)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		binding.server.SetKeepAlivesEnabled(false)
		if error := binding.server.Shutdown(ctx); error != nil {
			log.Errorf("Could not gracefully shutdown server %v", binding.Addr)
		}
	}()
}

func (binding *Binding) Apply(data interface{}) error {
	service, ok := data.(*models.WebService)
	if !ok {
		return errors.Errorf("unexpected parameter type %T in http binding", data)
	}

	for _, server := range service.Servers {
		address := fmt.Sprintf(":%v", server.Port)
		if binding.Addr != address {
			continue
		}

		host, found := binding.handlers[server.Host]
		if !found {
			log.Infof("Adding new host '%v' on binding %v", server.Host, binding.Addr)
			host = make(map[string]*ServiceHandler)
			binding.handlers[server.Host] = host
		}

		if handler, found := host[server.Path]; found {
			if service.Name != handler.WebService.Name {
				return errors.Errorf("service '%v' is already defined on path '%v'", handler.WebService.Name, server.Path)
			}
		} else {
			log.Infof("Adding service %v on binding %v on path %v", service.Name, binding.Addr, server.Path)
			handler = NewWebServiceHandler(service)
			host[server.Path] = handler
		}
	}

	return nil
}

func (binding *Binding) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service, servicePath := binding.resolveHandler(r)
	if service == nil {
		m := fmt.Sprintf("There was no service listening at %v", r.URL)
		http.Error(w, m, http.StatusInternalServerError)
		log.Error(m)
		//e.requestChannel <- &models.RequestMetric{Method: r.Method, Url: r.URL.String(), Error: m, HttpStatus: http.StatusInternalServerError}
		return
	}

	ctx := NewHttpContext(r, w, servicePath)
	ctx.metric = models.NewRequestMetric(r.Method, r.URL.String())

	service.ServeHTTP(ctx)

	binding.mh(ctx.metric)
}

func (binding *Binding) resolveHandler(r *http.Request) (*ServiceHandler, string) {
	var matchedPath string
	var matchedHandler *ServiceHandler
	if host, ok := binding.handlers[r.Host]; ok {
		for path, handler := range host {
			if strings.HasPrefix(strings.ToLower(r.URL.Path), strings.ToLower(path)) {
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

func (binding *Binding) getServicePath(service *models.WebService) (bool, string) {
	for _, server := range service.Servers {
		if fmt.Sprintf("%v:%v", server.Host, server.Port) == binding.Addr {
			return true, server.Path
		}
	}
	return false, ""
}
