package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/models"
	event "mokapi/models/eventService"
	"mokapi/server/api"
	"mokapi/server/kafka"
	"mokapi/server/ldap"
	"mokapi/server/web"
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

	if config.Api.Dashboard {
		addr := fmt.Sprintf(":%v", config.Api.Port)
		b := api.NewBinding(addr, server.application)
		server.Bindings[addr] = b
	}

	return server
}

func (s *Server) Start() {
	for _, b := range s.Bindings {
		b.Start()
	}
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
				binding = web.NewBinding(address, s.application.Metrics.AddRequest)
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

	for _, e := range s.application.EventServices {
		for _, server := range e.Servers {
			switch server.Type {
			case event.Kafka:
				b, ok := s.Bindings["0.0.0.0:9092"]
				if !ok {
					b = kafka.NewServer("0.0.0.0:9092")
					b.Start()
					s.Bindings["0.0.0.0:9092"] = b
				}
				b.Apply(e)
			default:
				log.Errorf("server type %v not supported", server.Type)
			}
		}
	}
}
