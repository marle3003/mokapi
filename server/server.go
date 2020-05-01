package server

import (
	"mokapi/config/dynamic"
	"mokapi/server/ldap"
	"mokapi/service"
)

type Server struct {
	apiServer *ApiServer
	watcher   *ConfigWatcher

	stopChannel chan bool

	ldapServers map[string]*ldap.Server
}

func NewServer(apiServer *ApiServer, watcher *ConfigWatcher) *Server {
	server := &Server{apiServer: apiServer, stopChannel: make(chan bool), watcher: watcher, ldapServers: make(map[string]*ldap.Server)}

	watcher.AddListener(func(s *service.Service) {
		apiServer.AddOrUpdate(s)
	})

	watcher.AddLdapListener(func(key string, config *dynamic.Ldap) {
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

func (server *Server) Start() {
	server.watcher.Start()
}

func (server *Server) Wait() {
	<-server.stopChannel
}
