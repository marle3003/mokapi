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
	"unicode"
)

var (
	noServiceFound = fmt.Errorf("there was no service listening at")
	tooManyMatches = fmt.Errorf("please use a specific domain: request could not be uniquely assigned to an API")
)

type HttpServer struct {
	server   *http.Server
	handlers map[string]map[string]*HttpService // map[host][path]Handler
	m        sync.RWMutex
	isTls    bool
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
		isTls:    false,
	}
	s.server.Handler = s
	return s
}

func NewHttpServerTls(port string, store *cert.Store) *HttpServer {
	s := NewHttpServer(port)
	s.server.TLSConfig = &tls.Config{
		GetCertificate: store.GetCertificate,
	}
	s.isTls = true
	return s
}

func (s *HttpServer) AddOrUpdate(service *HttpService) error {
	s.m.Lock()
	defer s.m.Unlock()

	hostname := service.Url.Hostname()
	paths, ok := s.handlers[hostname]
	if !ok {
		log.Infof("adding new HTTP host '%v' on binding %v", hostname, s.server.Addr)
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
		log.Infof("adding service '%v' on binding %v on path %v", service.Name, s.server.Addr, path)
		paths[service.Url.Path] = service
	}

	return nil
}

func (s *HttpServer) RemoveUrl(u *url.URL) {
	hostname := u.Hostname()
	if paths, ok := s.handlers[hostname]; ok {
		delete(paths, u.Path)
		if len(paths) == 0 {
			delete(s.handlers, hostname)
		}
	}
}

func (s *HttpServer) Remove(name string) {
	s.m.Lock()
	defer s.m.Unlock()

	for hostname, paths := range s.handlers {
		for path, service := range paths {
			if service.Name == name {
				log.Infof("removing service '%v' on binding %v on path %v", name, s.server.Addr, path)
				delete(paths, service.Url.Path)
			}
		}
		if len(paths) == 0 {
			delete(s.handlers, hostname)
		}
	}
}

func (s *HttpServer) Start() {
	go func() {
		var err error
		switch {
		case s.isTls:
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

func (s *HttpServer) CanClose() bool {
	return len(s.handlers) == 0
}

func (s *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(context.WithValue(r.Context(), "time", time.Now()))

	service, servicePath, err := s.resolveService(r)

	if service != nil {
		if !service.IsInternal {
			u := r.URL.String()
			if r.URL.Host == "" {
				u = r.Host + u
				if s.isTls {
					u = "https://" + u
				} else {
					u = "http://" + u
				}
			}
			log.Infof("processing http request %v %s", r.Method, u)
		}

		if service.Handler == nil {
			http.Error(w, "handler is nil", 500)
		} else {
			r = r.WithContext(context.WithValue(r.Context(), "servicePath", servicePath))
			service.Handler.ServeHTTP(w, r)
		}
	} else {
		if err == nil {
			err = noServiceFound
		}
		serveError(w, r, err)
	}
}

func (s *HttpServer) resolveService(r *http.Request) (*HttpService, string, error) {
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}

	if paths, ok := s.handlers[host]; ok {
		if matchedService, matchedPath := matchPath(paths, r); matchedService != nil {
			return matchedService, matchedPath, nil
		}
	}

	// any host
	if paths, ok := s.handlers[""]; ok {
		m, p := matchPath(paths, r)
		return m, p, nil
	}

	// try every host and check whether only one matches
	matches := map[string]*HttpService{}
	for _, paths := range s.handlers {
		if matchedService, matchedPath := matchPath(paths, r); matchedService != nil {
			matches[matchedPath] = matchedService
		}
	}
	if len(matches) == 1 {
		for k, v := range matches {
			return v, k, nil
		}
	} else if len(matches) > 1 {
		return nil, "", tooManyMatches
	}

	return nil, "", noServiceFound
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

func serveError(w http.ResponseWriter, r *http.Request, err error) {
	msg := fmt.Sprintf("%s %v", err.Error(), lib.GetUrl(r))
	entry := log.WithFields(log.Fields{"url": r.URL, "method": r.Method, "status": http.StatusNotFound})
	entry.Info(msg)
	http.Error(w, formatMessageForResponse(msg), 404)

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

	err = events.Push(l, events.NewTraits().WithNamespace("http"))
	if err != nil {
		log.Errorf("unable to log event: %v", err)
	}

	if m, ok := monitor.HttpFromContext(r.Context()); ok {
		m.RequestErrorCounter.WithLabel("").Add(1)
	}
}

func formatMessageForResponse(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
