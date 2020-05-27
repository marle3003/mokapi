package server

import (
	"mokapi/config/static"
	"mokapi/models"
	"mokapi/server/api"
	"mokapi/server/http"
	"mokapi/server/ldap"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	httpServer *http.Server
	watcher    *ConfigWatcher

	stopChannel chan bool

	ldapServers map[string]*ldap.Server

	application *models.Application
}

func NewServer(config *static.Config) *Server {
	application := models.NewApplication()
	apiHandler := api.New(application)
	httpServer := http.NewServer(apiHandler, config.Api)
	watcher := NewConfigWatcher(&config.Providers.File)

	server := &Server{
		httpServer:  httpServer,
		stopChannel: make(chan bool),
		watcher:     watcher,
		ldapServers: make(map[string]*ldap.Server),
		application: application,
	}

	watcher.AddListener(func(s *models.Service) {
		server.application.AddOrUpdateService(s)
		httpServer.AddOrUpdate(s)
	})

	watcher.AddLdapListener(func(key string, config *models.LdapServer) {
		if ldapServer, ok := server.ldapServers[key]; ok {
			ldapServer.UpdateConfig(config)
		} else {
			ldapServer := ldap.NewServer(config)
			server.ldapServers[key] = ldapServer
			go ldapServer.Start()
		}
	})

	return server
}

func (s *Server) Start() {
	s.watcher.Start()

	log.Error(":::TEST:::")
}

func (s *Server) Wait() {
	<-s.stopChannel
}

func (s *Server) Stop() {
	s.watcher.Stop()
	s.httpServer.Stop()

	for _, l := range s.ldapServers {
		l.Stop()
	}
}
