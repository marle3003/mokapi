package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/providers/openapi"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/server/service"
	"net/http"
	"net/url"
	"slices"
	"sync"
)

var DefaultHttpPort = 80
var DefaultHttpsPort = 443

type HttpManager struct {
	servers map[string]*service.HttpServer

	eventEmitter common.EventEmitter
	certStore    *cert.Store
	app          *runtime.App
	services     static.Services
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

func (m *HttpManager) AddService(name string, u *url.URL, h http.Handler, isInternal bool) error {
	server := m.getOrCreateServer(u.Scheme, u.Port())
	err := server.AddOrUpdate(&service.HttpService{
		Url:        u,
		Handler:    h,
		Name:       name,
		IsInternal: isInternal,
	})
	if err != nil {
		return err
	}
	return nil
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
	if e.Event == dynamic.Delete {
		m.app.Http.Remove(e.Config)
		if info.Config == nil {
			m.removeService(cfg.Info.Name)
			m.stopEmptyServers()
			return
		}
	} else if info == nil {
		info = m.app.Http.Add(e.Config)
	} else {
		oldServers := info.Servers
		info.AddConfig(e.Config)
		m.cleanupRemovedServers(info, oldServers)
	}

	for _, s := range info.Servers {
		u, err := parseUrl(s.Url)
		if err != nil {
			log.Errorf("url syntax error %v: %v", e.Config.Info.Url, err.Error())
			continue
		}

		err = m.AddService(cfg.Info.Name, u, info.Handler(m.app.Monitor.Http, m.eventEmitter), false)
		if err != nil {
			log.Warnf("unable to add '%v' on %v: %v", cfg.Info.Name, s.Url, err.Error())
			continue
		}
	}

	m.stopEmptyServers()
	log.Debugf("processed %v", e.Config.Info.Path())
}

func (m *HttpManager) Stop() {
	for _, server := range m.servers {
		server.Stop()
	}
}

func (m *HttpManager) getOrCreateServer(scheme string, port string) *service.HttpServer {
	m.m.RLock()
	server, found := m.servers[port]
	if found {
		m.m.RUnlock()
		return server
	}

	m.m.RUnlock()
	m.m.Lock()
	defer m.m.Unlock()

	if scheme == "https" {
		server = service.NewHttpServerTls(port, m.certStore)
	} else {
		server = service.NewHttpServer(port)
	}

	m.servers[port] = server
	server.Start()
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

func (m *HttpManager) cleanupRemovedServers(cfg *runtime.HttpInfo, old []*openapi.Server) {
	for _, server := range old {
		if !slices.ContainsFunc(cfg.Servers, func(s *openapi.Server) bool {
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
			log.Infof("removing '%v' on binding %v on path %v", cfg.Info.Name, u.Port(), path)
			s.RemoveUrl(u)
		}
	}
}

func (m *HttpManager) stopEmptyServers() {
	for port, server := range m.servers {
		if server.CanClose() {
			log.Infof("stopping HTTP server on binding :%v", port)
			server.Stop()
			delete(m.servers, port)
		}
	}
}
