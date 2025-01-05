package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"mokapi/lib"
	"mokapi/providers/openapi"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/server/cert"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type HttpServer struct {
	server   *http.Server
	handlers map[string]map[string]*HttpService // map[host][path]Handler
	m        sync.RWMutex
}

type HttpService struct {
	Url        *url.URL
	Handler    http.Handler
	Name       string
	IsInternal bool
}

func NewHttpServer(port string) *HttpServer {
	s := &HttpServer{
		server:   &http.Server{Addr: fmt.Sprintf(":%v", port)},
		handlers: make(map[string]map[string]*HttpService),
	}
	s.server.Handler = s
	return s
}

func NewHttpServerTls(port string, store *cert.Store) *HttpServer {
	s := NewHttpServer(port)
	s.server.TLSConfig = &tls.Config{
		GetCertificate: store.GetCertificate,
	}
	return s
}

func (s *HttpServer) IsTls() bool {
	return s.server.TLSConfig != nil
}

func (s *HttpServer) AddOrUpdate(service *HttpService) error {
	s.m.Lock()
	defer s.m.Unlock()

	hostname := service.Url.Hostname()
	paths, ok := s.handlers[hostname]
	if !ok {
		log.Infof("adding new host '%v' on binding %v", hostname, s.server.Addr)
		paths = make(map[string]*HttpService)
		s.handlers[hostname] = paths
	}

	if serviceReg, found := paths[service.Url.Path]; found {
		if service.Name != serviceReg.Name {
			return fmt.Errorf("service '%v' is already defined on path '%v'", serviceReg.Name, service.Url.Path)
		} else {
			paths[service.Url.Path] = service
		}
	} else {
		path := service.Url.Path
		if len(path) == 0 {
			path = "/"
		}
		log.Infof("adding service %v on binding %v on path %v", service.Name, s.server.Addr, path)
		paths[service.Url.Path] = service
	}

	return nil
}

func (s *HttpServer) Start() {
	go func() {
		var err error
		switch {
		case s.IsTls():
			err = s.server.ListenAndServeTLS("", "")
		default:
			err = s.server.ListenAndServe()
		}
		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("unable to start http server %v: %v", s.server.Addr, err)
		}
	}()
}

func (s *HttpServer) Stop() {
	err := s.server.Close()
	if err != nil {
		log.Errorf("unable to stop http server %v: %v", s.server.Addr, err)
	}
}

func (s *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(context.WithValue(r.Context(), "time", time.Now()))

	service, servicePath := s.resolveService(r)

	if service != nil {
		if !service.IsInternal {
			log.WithFields(log.Fields{
				"url":    r.URL.String(),
				"host":   r.Host,
				"method": r.Method,
			}).Info("processing http request")
		}

		if service.Handler == nil {
			http.Error(w, "handler is nil", 500)
		} else {
			r = r.WithContext(context.WithValue(r.Context(), "servicePath", servicePath))
			service.Handler.ServeHTTP(w, r)
		}
	} else {
		serveNoServiceFound(w, r)
	}
}

func (s *HttpServer) resolveService(r *http.Request) (*HttpService, string) {
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}

	if paths, ok := s.handlers[host]; ok {
		if matchedService, matchedPath := matchPath(paths, r); matchedService != nil {
			return matchedService, matchedPath
		}
	}

	// any host
	if paths, ok := s.handlers[""]; ok {
		return matchPath(paths, r)
	}

	return nil, ""
}

func matchPath(paths map[string]*HttpService, r *http.Request) (matchedService *HttpService, matchedPath string) {
	for path, handler := range paths {
		if strings.HasPrefix(strings.ToLower(r.URL.Path), strings.ToLower(path)) {
			if matchedPath == "" || len(matchedPath) < len(path) {
				matchedPath = path
				matchedService = handler
			}
		}
	}
	return
}

func serveNoServiceFound(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("There was no service listening at %v", lib.GetUrl(r))
	entry := log.WithFields(log.Fields{"url": r.URL, "method": r.Method, "status": http.StatusNotFound})
	entry.Info(msg)
	http.Error(w, msg, 404)

	body, _ := io.ReadAll(r.Body)
	l := &openapi.HttpLog{
		Request: &openapi.HttpRequestLog{
			Method:      r.Method,
			Url:         lib.GetUrl(r),
			ContentType: r.Header.Get("content-type"),
			Body:        string(body),
		},
		Response: &openapi.HttpResponseLog{
			Headers:    map[string]string{"Content-Type": w.Header().Get("Content-Type")},
			StatusCode: 404,
			Body:       msg,
		},
	}

	err := events.Push(l, events.NewTraits().WithNamespace("http"))
	if err != nil {
		log.Errorf("unable to log event: %v", err)
	}

	if m, ok := monitor.HttpFromContext(r.Context()); ok {
		m.RequestErrorCounter.WithLabel("").Add(1)
	}
}
