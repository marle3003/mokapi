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

type EventHandler func(events event.Handler, options ...workflow.Options) *runtime.Summary

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
		log.Infof("starting web binding %v", binding.Addr)
		err := binding.server.ListenAndServe()
		if err != nil {
			log.Errorf("unable to start web binding %v: %v ", binding.Addr, err.Error())
		}
	}()
}

func (binding *Binding) Stop() {
	go func() {
		log.Infof("stopping server on %v", binding.Addr)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		binding.server.SetKeepAlivesEnabled(false)
		if err := binding.server.Shutdown(ctx); err != nil {
			log.Errorf("could not gracefully shutdown server %v: %v", binding.Addr, err.Error())
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
	ctx := NewHttpContext(r, w, binding.workflowHandler)

	defer binding.addRequestMetric(ctx.metric)

	var service *ServiceHandler
	service, ctx.ServicPath = binding.resolveHandler(r)

	if service != nil {
		service.ServeHTTP(ctx)
	} else {
		m := fmt.Sprintf("There was no service listening at %v", r.URL)
		writeError(m, http.StatusInternalServerError, ctx)
	}
}

func (binding *Binding) resolveHandler(r *http.Request) (*ServiceHandler, string) {
	var matchedPath string
	var matchedHandler *ServiceHandler
	rHost := strings.Split(r.Host, ":")[0]
	if host, ok := binding.handlers[rHost]; ok {
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

func writeError(message string, status int, ctx *HttpContext) {
	ctx.updateMetricWithError(status, message)
	log.WithFields(log.Fields{"url": ctx.metric.Url, "method": ctx.metric.Method, "status": status}).Error(message)
	http.Error(ctx.Response, message, status)
}
