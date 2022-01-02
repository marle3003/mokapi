package web

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"mokapi/config/dynamic/openapi"
	"mokapi/engine"
	"mokapi/models"
	"mokapi/server/cert"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"net/http"
)

type AddRequestMetric func(metric *models.RequestMetric)

type eventHandler func(request *Request, response *Response) []*engine.Summary

type Binding struct {
	Addr             string
	server           *http.Server
	handlers         map[string]map[string]*serviceHandler // map[host][path]Handler
	addRequestMetric AddRequestMetric
	eventHandler     eventHandler
	IsTls            bool
	certificates     map[string]*tls.Certificate
	Engine           *engine.Engine
}

func NewBinding(addr string) *Binding {
	b := &Binding{
		Addr:     addr,
		handlers: make(map[string]map[string]*serviceHandler),
	}
	b.server = &http.Server{Addr: addr, Handler: b}

	return b
}

func NewBindingWithTls(addr string, store *cert.Store) *Binding {
	b := &Binding{
		Addr:         addr,
		handlers:     make(map[string]map[string]*serviceHandler),
		IsTls:        true,
		certificates: make(map[string]*tls.Certificate),
	}

	b.server = &http.Server{
		Addr:    addr,
		Handler: b,
		TLSConfig: &tls.Config{
			GetCertificate: store.GetCertificate,
		},
	}

	return b
}

func (binding *Binding) Start() {
	go func() {
		if binding.IsTls {
			log.Infof("starting https binding %v", binding.Addr)
			err := binding.server.ListenAndServeTLS("", "")
			if err != nil {
				log.Errorf("unable to start https binding %v: %v ", binding.Addr, err.Error())
			}
		} else {
			log.Infof("starting http binding %v", binding.Addr)
			err := binding.server.ListenAndServe()
			if err != nil {
				log.Errorf("unable to start http binding %v: %v ", binding.Addr, err.Error())
			}
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
		if len(strings.TrimSpace(server.Url)) == 0 {
			continue
		}

		hostName, port, path, err := ParseAddress(server.Url)
		if err != nil {
			log.Errorf("API %v: %v", service.Info.Name, err)
			continue
		}

		address := fmt.Sprintf(":%v", port)
		if binding.Addr != address {
			continue
		}

		host, found := binding.handlers[hostName]
		if !found {
			log.Infof("Adding new host '%v' on binding %v", hostName, binding.Addr)
			host = make(map[string]*serviceHandler)
			binding.handlers[hostName] = host
		}

		if handler, found := host[path]; found {
			if service.Info.Name != handler.config.Info.Name {
				return errors.Errorf("service '%v' is already defined on path '%v'", handler.config.Info.Name, path)
			}
		} else {
			log.Infof("Adding service %v on binding %v on path %v", service.Info.Name, binding.Addr, path)
			handler = newServiceHandler(service)
			host[path] = handler
		}
	}

	return nil
}

func (binding *Binding) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewHttpContext(r, w)
	ctx.engine = binding.Engine

	defer binding.setMetric(ctx.metric)

	var service *serviceHandler
	service, ctx.ServicePath = binding.resolveHandler(r)

	if service != nil {
		service.ServeHTTP(ctx)
	} else {
		m := fmt.Sprintf("There was no service listening at %v", r.URL)
		writeError(m, http.StatusNotFound, ctx)
	}
}

func (binding *Binding) resolveHandler(r *http.Request) (*serviceHandler, string) {
	var matchedPath string
	var matchedHandler *serviceHandler
	rHost := strings.Split(r.Host, ":")[0]
	if host, ok := binding.handlers[rHost]; ok {
		matchedHandler, matchedPath = matchPath(host, r)
	}

	if matchedHandler != nil {
		return matchedHandler, matchedPath
	}

	if host, ok := binding.handlers[""]; ok {
		return matchPath(host, r)
	}

	return nil, ""
}

func (binding *Binding) setMetric(metric *models.RequestMetric) {
	if binding.addRequestMetric != nil {
		binding.addRequestMetric(metric)
	}
}

func matchPath(host map[string]*serviceHandler, r *http.Request) (matchedHandler *serviceHandler, matchedPath string) {
	for path, handler := range host {
		if strings.HasPrefix(strings.ToLower(r.URL.Path), strings.ToLower(path)) {
			if matchedPath == "" || len(matchedPath) < len(path) {
				matchedPath = path
				matchedHandler = handler
			}
		}
	}
	return
}

func writeError(message string, status int, ctx *HttpContext) {
	ctx.updateMetricWithError(status, message)
	entry := log.WithFields(log.Fields{"url": ctx.metric.Url, "method": ctx.metric.Method, "status": status})
	if status == http.StatusInternalServerError {
		entry.Error(message)
	} else {
		entry.Info(message)
	}
	http.Error(ctx.ResponseWriter, message, status)
}
