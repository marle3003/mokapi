package server

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/engine/common"
	"mokapi/providers/openapi"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/server/service"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

var DefaultHttpPort = 80
var DefaultHttpsPort = 443

type HttpManager struct {
	servers map[string]*service.HttpServer

	eventEmitter common.EventEmitter
	certStore    *cert.Store
	app          *runtime.App
	m            sync.RWMutex
}

func NewHttpManager(emitter common.EventEmitter, store *cert.Store, app *runtime.App) *HttpManager {
	return &HttpManager{
		eventEmitter: emitter,
		certStore:    store,
		app:          app,
		servers:      make(map[string]*service.HttpServer),
	}
}

func (m *HttpManager) AddService(name string, u *url.URL, h openapi.Handler) error {
	return m.addService(&service.HttpService{
		Url:     u,
		Handler: h,
		Name:    name,
	})
}

func (m *HttpManager) AddInternalService(name string, u *url.URL, h http.Handler) error {
	return m.addService(&service.HttpService{
		Url:        u,
		Handler:    &service.StdHandlerAdapter{H: h},
		Name:       name,
		IsInternal: true,
	})
}

// addService creates the server if needed and adds the service in one critical section. A server
// is therefore never visible in m.servers without a handler, so stopEmptyServers of a parallel
// update cannot stop a server that is just about to receive its first service.
func (m *HttpManager) addService(svc *service.HttpService) error {
	m.m.Lock()
	defer m.m.Unlock()

	server := m.getOrCreateServer(svc.Url.Scheme, svc.Url.Port())
	return server.AddOrUpdate(svc)
}

func (m *HttpManager) removeService(name string) {
	m.m.Lock()
	defer m.m.Unlock()

	for _, server := range m.servers {
		server.Remove(name)
	}
}

func (m *HttpManager) Update(e dynamic.ConfigEvent) {
	cfg, ok := runtime.IsHttpConfig(e.Config)
	if !ok {
		return
	}

	info := m.app.Http.Get(cfg.Info.Name)
	var servers []*openapi.Server
	if info != nil {
		if old := info.Config(); old != nil {
			servers = old.Servers
		}
	}

	if e.Event == dynamic.Delete {
		if info == nil {
			return
		}
		m.app.Http.Remove(e.Config)
		if info.Config() == nil {
			m.removeService(cfg.Info.Name)
			m.stopEmptyServers()
			return
		}
	} else {
		info = m.app.Http.Add(e.Config)
	}

	m.addServices(info, servers, e.Config.Info.Url)

	m.stopEmptyServers()
	log.Debugf("processed %v", e.Config.Info.Path())
}

// addServices registers the handler of the (merged) config on every server URL of the config
// and removes the URLs that are no longer part of it. The handlers of one update are all created
// from one snapshot. file is the config file that triggered the update and only used for log messages.
func (m *HttpManager) addServices(info *runtime.HttpInfo, old []*openapi.Server, file *url.URL) {
	cfg := info.Config()
	if cfg == nil {
		return
	}

	m.cleanupRemovedServers(cfg.Info.Name, old, cfg.Servers)

	for _, s := range cfg.Servers {
		u, err := parseUrl(s.Url)
		if err != nil {
			log.Errorf("url syntax error %v: %v", file, err.Error())
			continue
		}

		err = m.AddService(cfg.Info.Name, u, info.Handler(cfg, m.app.Monitor.Http, m.eventEmitter, m.app.Events))
		if err != nil {
			log.Warnf("unable to add '%v' on %v: %v", cfg.Info.Name, s.Url, err.Error())
			continue
		}
	}
}

func (m *HttpManager) Stop() {
	m.m.Lock()
	defer m.m.Unlock()

	for _, server := range m.servers {
		server.Stop()
	}
}

// getOrCreateServer expects the caller to hold m.m.
func (m *HttpManager) getOrCreateServer(scheme string, port string) *service.HttpServer {
	if server, found := m.servers[port]; found {
		return server
	}

	var server *service.HttpServer
	if scheme == "https" {
		server = service.NewHttpServerTls(port, m.certStore, m.app.Events)
	} else {
		server = service.NewHttpServer(port, m.app.Events)
	}

	m.servers[port] = server
	if err := server.Start(); err != nil {
		log.Errorf("failed to start %s server :%s: %s", strings.ToUpper(scheme), port, err)
	}
	return server
}

func parseUrl(s string) (u *url.URL, err error) {
	u, err = url.Parse(s)
	if err != nil {
		return
	}

	port := u.Port()
	if len(port) == 0 {
		switch u.Scheme {
		case "https":
			port = fmt.Sprintf("%d", DefaultHttpsPort)
		default:
			port = fmt.Sprintf("%d", DefaultHttpPort)
		}
		u.Host = fmt.Sprintf("%v:%v", u.Hostname(), port)
	}

	return
}

func (m *HttpManager) cleanupRemovedServers(name string, old, current []*openapi.Server) {
	m.m.Lock()
	defer m.m.Unlock()

	for _, server := range old {
		if !slices.ContainsFunc(current, func(s *openapi.Server) bool {
			return s.Url == server.Url
		}) {
			u, err := parseUrl(server.Url)
			if err != nil {
				continue
			}
			s, ok := m.servers[u.Port()]
			if !ok {
				continue
			}
			path := u.Path
			if path == "" {
				path = "/"
			}
			log.Infof("removing '%v' on binding %v on path %v", name, u.Port(), path)
			s.RemoveUrl(u)
		}
	}
}

func (m *HttpManager) stopEmptyServers() {
	m.m.Lock()
	defer m.m.Unlock()

	for port, server := range m.servers {
		if server.CanClose() {
			log.Infof("stopping HTTP server on binding :%v", port)
			server.Stop()
			delete(m.servers, port)
		}
	}
}
