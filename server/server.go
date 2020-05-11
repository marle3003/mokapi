package server

import (
	"mokapi/models"
	"mokapi/server/http"
	"mokapi/server/ldap"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	httpServer *http.Server
	watcher    *ConfigWatcher

	stopChannel chan bool

	ldapServers map[string]*ldap.Server
}

func NewServer(watcher *ConfigWatcher) *Server {
	httpServer := http.NewServer()
	server := &Server{httpServer: httpServer, stopChannel: make(chan bool), watcher: watcher, ldapServers: make(map[string]*ldap.Server)}

	watcher.AddListener(func(s *models.Service) {
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
