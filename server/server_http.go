package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/openapi"
	"mokapi/engine"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/server/service"
	"net/http"
	"net/url"
)

type HttpServers map[string]*service.HttpServer

type HttpManager struct {
	Servers HttpServers

	eventEmitter engine.EventEmitter
	certStore    *cert.Store
	app          *runtime.App
}

func NewHttpManager(servers HttpServers, emitter engine.EventEmitter, store *cert.Store, app *runtime.App) *HttpManager {
	return &HttpManager{
		eventEmitter: emitter,
		certStore:    store,
		app:          app,
		Servers:      servers,
	}
}

func (m *HttpManager) AddService(name string, u *url.URL, h http.Handler) error {
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
	err := server.AddOrUpdate(&service.HttpService{
		Url:     u,
		Handler: runtime.NewHttpHandler(m.app.Monitor.Http, h),
		Name:    name,
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *HttpManager) Update(c *common.Config) {
	config, ok := c.Data.(*openapi.Config)
	if !ok {
		return
	}

	if err := config.Validate(); err != nil {
		log.Warnf("validation error %v: %v", c.Url, err)
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

		err = m.AddService(config.Info.Name, u, openapi.NewHandler(config, m.eventEmitter))
		if err != nil {
			log.Errorf("error on updating %v: %v", c.Url.String(), err.Error())
			return
		}
	}
	log.Infof("processed file %v", c.Url.String())
}

func (m *HttpManager) createOpenApiService(u *url.URL, config *openapi.Config) *service.HttpService {
	return &service.HttpService{
		Url:     u,
		Handler: runtime.NewHttpHandler(m.app.Monitor.Http, openapi.NewHandler(config, m.eventEmitter)),
		Name:    config.Info.Name,
	}
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
