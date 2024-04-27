package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/server/service"
	"net/http"
	"net/url"
	"sync"
)

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

func (m *HttpManager) Update(c *dynamic.Config) {
	if !runtime.IsHttpConfig(c) {
		return
	}
	cfg := m.app.AddHttp(c)

	for _, s := range cfg.Servers {
		u, err := parseUrl(s.Url)
		if err != nil {
			log.Errorf("url syntax error %v: %v", c.Info.Url, err.Error())
			continue
		}

		err = m.AddService(cfg.Info.Name, u, cfg.Handler(m.app.Monitor.Http, m.eventEmitter), false)
		if err != nil {
			log.Warnf("unable to add '%v' on %v: %v", cfg.Info.Name, s.Url, err.Error())
			continue
		}
	}
	log.Debugf("processed %v", c.Info.Path())
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
			port = "443"
		default:
			port = "80"
		}
		u.Host = fmt.Sprintf("%v:%v", u.Hostname(), port)
	}

	return
}
