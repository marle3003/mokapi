package service

import (
	"cmp"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"maps"
	"mokapi/lib"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/parameter"
	"mokapi/runtime/events"
	"mokapi/runtime/monitor"
	"mokapi/server/cert"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type HttpServer struct {
	server   *http.Server
	handlers map[string]*HttpHost
	eh       events.Handler
	m        sync.RWMutex
	isTls    bool
}

type HttpService struct {
	Url        *url.URL
	Handler    openapi.Handler
	Name       string
	IsInternal bool
}

type HttpHost struct {
	Paths map[string]HttpServices
}

type HttpServices map[string]*HttpService

func NewHttpServer(port string, eh events.Handler) *HttpServer {
	s := &HttpServer{
		server:   &http.Server{Addr: fmt.Sprintf(":%v", port)},
		handlers: map[string]*HttpHost{},
		eh:       eh,
		isTls:    false,
	}
	s.server.Handler = s
	return s
}

func NewHttpServerTls(port string, store *cert.Store, eh events.Handler) *HttpServer {
	s := NewHttpServer(port, eh)
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
	host, ok := s.handlers[hostname]
	if !ok {

		log.Infof("adding new %s host '%s' on binding %s", s.getProto(), hostname, s.server.Addr)
		host = &HttpHost{Paths: map[string]HttpServices{}}
		s.handlers[hostname] = host
	}

	path := service.Url.Path
	if len(path) == 0 {
		path = "/"
	}

	services, ok := host.Paths[path]
	if !ok {
		services = HttpServices{}
		host.Paths[path] = services
	}

	return services.Add(service, s.server.Addr, path)
}

func (s *HttpServer) RemoveUrl(u *url.URL) {
	hostname := u.Hostname()
	if host, ok := s.handlers[hostname]; ok {
		path := u.Path
		if len(path) == 0 {
			path = "/"
		}
		delete(host.Paths, path)
		if len(host.Paths) == 0 {
			delete(s.handlers, hostname)
		}
	}
}

func (s *HttpServer) Remove(name string) {
	s.m.Lock()
	defer s.m.Unlock()

	for hostname, host := range s.handlers {
		for path, services := range host.Paths {
			_, ok := services[name]
			if ok {
				log.Infof("removing service '%v' on binding %v on path %v", name, s.server.Addr, path)
				delete(services, name)
			}
			if len(services) == 0 {
				delete(host.Paths, path)
			}
		}
		if len(host.Paths) == 0 {
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
			log.Errorf("failed to start %s server %s: %s", s.getProto(), s.server.Addr, err)
		}
	}()
}

func (s *HttpServer) Stop() {
	err := s.server.Close()
	if err != nil {
		log.Errorf("failed to stop %s server %s: %s", s.getProto(), s.server.Addr, err)
	}
}

func (s *HttpServer) CanClose() bool {
	return len(s.handlers) == 0
}

func (s *HttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(context.WithValue(r.Context(), "time", time.Now()))

	httpError := s.dispatchRequest(w, r)
	if httpError != nil {
		serveError(w, r, httpError, s.eh)
	}
}

type logHttpRequestContext struct {
	logged bool
}

func (s *HttpServer) dispatchRequest(rw http.ResponseWriter, r *http.Request) *openapi.HttpError {
	hostname, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		hostname = r.Host
	}

	r = r.WithContext(context.WithValue(r.Context(), "logHttpRequestContext", &logHttpRequestContext{logged: false}))

	var httpError *openapi.HttpError
	if host, ok := s.handlers[hostname]; ok {
		services := matchPath(host.Paths, r)
		httpError = s.serve(services, rw, r)
		if httpError == nil || httpError.StatusCode != http.StatusNotFound {
			return httpError
		}
	}

	// any host
	if host, ok := s.handlers[""]; ok {
		services := matchPath(host.Paths, r)
		httpError = s.serve(services, rw, r)
		if httpError == nil || httpError.StatusCode != http.StatusNotFound {
			return httpError
		}
	}

	// try every host
	for _, host := range s.handlers {
		services := matchPath(host.Paths, r)
		httpError = s.serve(services, rw, r)
		if httpError == nil || httpError.StatusCode != http.StatusNotFound {
			return httpError
		}
	}

	if httpError != nil {
		return httpError
	}

	return &openapi.HttpError{
		StatusCode: http.StatusNotFound,
		Message:    fmt.Sprintf("There was no service listening at %s", lib.GetUrl(r)),
	}
}

func (s *HttpServer) serve(services map[string][]*HttpService, rw http.ResponseWriter, r *http.Request) *openapi.HttpError {
	keys := slices.Collect(maps.Keys(services))
	slices.SortFunc(keys, func(a, b string) int {
		return cmp.Compare(len(a), len(b))
	})

	var httpError *openapi.HttpError
	for _, key := range keys {
		r = r.WithContext(context.WithValue(r.Context(), "servicePath", key))
		for _, service := range services[key] {
			if !service.IsInternal {
				ctx := r.Context().Value("logHttpRequestContext").(*logHttpRequestContext)
				if !ctx.logged {
					ctx.logged = true
					log.Infof("processing %s request %s %s", s.getProto(), r.Method, lib.GetUrl(r))
				}
			}

			if service.Handler == nil {
				return &openapi.HttpError{StatusCode: http.StatusInternalServerError, Message: fmt.Sprintf("Handler is nil for %s", lib.GetUrl(r))}
			}

			httpError = service.Handler.ServeHTTP(rw, r)
			if httpError == nil || httpError.StatusCode != http.StatusNotFound {
				return httpError
			}
		}
	}

	if httpError != nil {
		return httpError
	}

	return &openapi.HttpError{
		StatusCode: http.StatusNotFound,
		Message:    fmt.Sprintf("There was no service listening at %s", lib.GetUrl(r)),
	}
}

func (s *HttpServer) getProto() string {
	if s.isTls {
		return "HTTPS"
	}
	return "HTTP"
}

func matchPath(paths map[string]HttpServices, r *http.Request) map[string][]*HttpService {
	results := map[string][]*HttpService{}
	for path, services := range paths {
		if strings.HasPrefix(strings.ToLower(r.URL.Path), strings.ToLower(path)) {
			results[path] = append(results[path], slices.Collect(maps.Values(services))...)
		}
	}
	return results
}

func serveError(w http.ResponseWriter, r *http.Request, err error, eh events.Handler) {
	status := http.StatusInternalServerError
	var msg string

	var hErr *openapi.HttpError
	if errors.As(err, &hErr) {
		status = hErr.StatusCode
		for k, values := range hErr.Header {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
		msg = hErr.Message
	} else {
		msg = fmt.Sprintf("%s %v", err.Error(), lib.GetUrl(r))
	}

	entry := log.WithFields(log.Fields{"url": r.URL, "method": r.Method, "status": status})
	entry.Info(msg)
	http.Error(w, formatMessageForResponse(msg), status)

	body, _ := io.ReadAll(r.Body)
	l := &openapi.HttpLog{
		Request: &openapi.HttpRequestLog{
			Method:      r.Method,
			Url:         lib.GetUrl(r),
			ContentType: r.Header.Get("Content-Type"),
			Body:        string(body),
		},
		Response: &openapi.HttpResponseLog{
			Headers:    map[string]string{"Content-Type": w.Header().Get("Content-Type")},
			StatusCode: status,
			Body:       msg,
		},
	}

	for k, v := range r.Header {
		raw := strings.Join(v, ",")
		p := openapi.HttpParameter{
			Name: k,
			Type: string(parameter.Header),
			Raw:  &raw,
		}
		l.Request.Parameters = append(l.Request.Parameters, p)
	}

	err = eh.Push(l, events.NewTraits().WithNamespace("http"))
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

func (m HttpServices) Add(service *HttpService, addr, path string) error {
	for _, s := range m {
		if s.IsInternal {
			return fmt.Errorf("internal service '%v' is already defined on path '%v'", s.Name, s.Url.Path)
		}
	}

	if _, ok := m[service.Name]; !ok {
		log.Infof("adding service '%v' on binding %v on path %v", service.Name, addr, path)
	}

	m[service.Name] = service
	return nil
}

type StdHandlerAdapter struct {
	H http.Handler
}

func (a *StdHandlerAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) *openapi.HttpError {
	a.H.ServeHTTP(w, r)
	return nil
}
