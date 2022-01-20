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

func (m *HttpManager) Update(file *common.File) {
	config, ok := file.Data.(*openapi.Config)
	if !ok {
		return
	}

	if err := config.Validate(); err != nil {
		log.Warnf("validation error %v: %v", file.Url, err)
		return
	}

	if len(config.Servers) == 0 {
		config.Servers = append(config.Servers, &openapi.Server{Url: "/"})
	}

	for _, s := range config.Servers {
		u, err := parseUrl(s.Url)
		if err != nil {
			log.Errorf("error %v: %v", file.Url, err.Error())
			continue
		}

		server, found := m.Servers[u.Port()]
		if !found {
			if u.Scheme == "https" {
				server = service.NewHttpServerTls(u.Port(), m.certStore)
			} else {
				server = service.NewHttpServer(u.Port())
			}

			m.Servers[u.Port()] = server
			server.Start()
			m.app.AddHttp(config)
		}
		err = server.AddOrUpdate(m.createService(u, config))
		if err != nil {
			log.Errorf("error on updating %v: %v", file.Url.String(), err.Error())
			return
		}
	}
	log.Infof("processed config %v", file.Url.String())
}

func (m *HttpManager) createService(u *url.URL, config *openapi.Config) *service.HttpService {
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
