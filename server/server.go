package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/models"
	"mokapi/server/ldap"
)

type Binding interface {
	Start()
	Stop()
	Apply(interface{}) error
}

type Server struct {
	watcher     *ConfigWatcher
	application *models.Application
	stopChannel chan bool

	Bindings map[string]Binding
}

func NewServer(config *static.Config) *Server {
	application := models.NewApplication()
	watcher := NewConfigWatcher(&config.Providers.File)

	server := &Server{
		watcher:     watcher,
		application: application,
		stopChannel: make(chan bool),
		Bindings:    make(map[string]Binding),
	}

	watcher.AddListener(func(config *dynamic.Configuration) {
		server.application.Apply(config)
		server.updateBindings()
	})

	return server
}

func (s *Server) Start() {
	s.watcher.Start()
}

func (s *Server) Wait() {
	<-s.stopChannel
}

func (s *Server) Stop() {
	s.watcher.Stop()
}

func (s *Server) updateBindings() {
	for _, webService := range s.application.WebServices {
		for _, server := range webService.Data.Servers {
			address := fmt.Sprintf(":%v", server.Port)
			binding, found := s.Bindings[address]
			if !found {
				binding = NewHttpBinding(address)
				s.Bindings[address] = binding
				binding.Start()
			}
			err := binding.Apply(webService.Data)
			if err != nil {
				log.Error(err.Error())
			}
		}

	}

	for _, config := range s.application.LdapServices {
		if b, ok := s.Bindings[config.Data.Address]; !ok {
			lserver := ldap.NewServer(config.Data)
			s.Bindings[config.Data.Address] = lserver
			lserver.Start()
		} else {
			b.Apply(config.Data)
		}

	}
}
