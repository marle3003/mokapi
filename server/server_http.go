package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	config "mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/server/service"
	"net/http"
	"net/url"
)

type HttpServers map[string]*service.HttpServer

type HttpManager struct {
	Servers HttpServers

	eventEmitter common.EventEmitter
	certStore    *cert.Store
	app          *runtime.App
}

func NewHttpManager(servers HttpServers, emitter common.EventEmitter, store *cert.Store, app *runtime.App) *HttpManager {
	return &HttpManager{
		eventEmitter: emitter,
		certStore:    store,
		app:          app,
		Servers:      servers,
	}
}

func (m *HttpManager) AddService(name string, u *url.URL, handler http.Handler, isInternal bool) error {
	server, found := m.Servers[u.Port()]
	if !found {
		if u.Scheme == "https" {
			server = service.NewHttpServerTls(u.Port(), m.certStore)
		} else {
			server = service.NewHttpServer(u.Port())
		}

		m.Servers[u.Port()] = server
		server.Start()
	}

	h := handler
	if !isInternal {
		h = runtime.NewHttpHandler(m.app.Monitor.Http, h)
	}

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

func (m *HttpManager) Update(c *config.Config) {
	config, ok := c.Data.(*openapi.Config)
	if !ok {
		return
	}

	m.app.AddHttp(config)

	if len(config.Servers) == 0 {
		config.Servers = append(config.Servers, &openapi.Server{Url: "/"})
	}

	for _, s := range config.Servers {
		u, err := parseUrl(s.Url)
		if err != nil {
			log.Errorf("error %v: %v", c.Url, err.Error())
			continue
		}

		err = m.AddService(config.Info.Name, u, openapi.NewHandler(config, m.eventEmitter), false)
		if err != nil {
			log.Errorf("error on updating %v: %v", c.Url.String(), err.Error())
			return
		}
	}
	log.Debugf("processed %v", c.Url.String())
}

func (servers HttpServers) Stop() {
	for _, server := range servers {
		server.Stop()
	}
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
