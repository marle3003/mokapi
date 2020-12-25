package server

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/models"
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
		//httpServer.AddOrUpdate(s)
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
			address := fmt.Sprintf("%v:%v", server.Host, server.Port)
			binding, found := s.Bindings[address]
			if !found {
				binding = NewHttpBinding(address)
				s.Bindings[address] = binding
				binding.Start()
			}
			binding.Apply(webService.Data)
		}

	}
}
