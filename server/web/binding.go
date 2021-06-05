package web

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"mokapi/config/dynamic/openapi"
	"mokapi/models"
	"mokapi/providers/workflow"
	"mokapi/providers/workflow/event"
	"mokapi/providers/workflow/runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"net/http"
)

type AddRequestMetric func(metric *models.RequestMetric)

type EventHandler func(events event.Handler, options ...workflow.WorkflowOptions) *runtime.Summary

type Binding struct {
	Addr             string
	server           *http.Server
	handlers         map[string]map[string]*ServiceHandler
	addRequestMetric AddRequestMetric
	workflowHandler  EventHandler
}

func NewBinding(addr string, mh AddRequestMetric, wh EventHandler) *Binding {
	b := &Binding{
		Addr:             addr,
		handlers:         make(map[string]map[string]*ServiceHandler),
		addRequestMetric: mh,
		workflowHandler:  wh,
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
	service, ok := data.(*openapi.Config)
	if !ok {
		return errors.Errorf("unexpected parameter type %T in http binding", data)
	}

	for _, server := range service.Servers {
		hostName, port, path := server.GetHost(), server.GetPort(), server.GetPath()

		address := fmt.Sprintf(":%v", port)
		if binding.Addr != address {
			continue
		}

		host, found := binding.handlers[hostName]
		if !found {
			log.Infof("Adding new host '%v' on binding %v", hostName, binding.Addr)
			host = make(map[string]*ServiceHandler)
			binding.handlers[hostName] = host
		}

		if handler, found := host[path]; found {
			if service.Info.Name != handler.config.Info.Name {
				return errors.Errorf("service '%v' is already defined on path '%v'", handler.config.Info.Name, path)
			}
		} else {
			log.Infof("Adding service %v on binding %v on path %v", service.Info.Name, binding.Addr, path)
			handler = NewWebServiceHandler(service)
			host[path] = handler
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

	ctx := NewHttpContext(r, w, servicePath, binding.workflowHandler)
	ctx.metric = models.NewRequestMetric(r.Method, fmt.Sprintf("%s%s", r.Host, r.URL.String()), service.config)

	service.ServeHTTP(ctx)

	binding.addRequestMetric(ctx.metric)
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
