package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/ldap"
	"mokapi/config/dynamic/mokapi"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/static"
	"mokapi/models"
	"mokapi/providers/pipeline"
	"mokapi/server/api"
	"mokapi/server/kafka"
	ldapServer "mokapi/server/ldap"
	"mokapi/server/web"
)

type Binding interface {
	Start()
	Stop()
	Apply(interface{}) error
}

type Server struct {
	watcher     *dynamic.ConfigWatcher
	runtime     *models.Runtime
	stopChannel chan bool

	Bindings   map[string]Binding
	Schedulers map[string]*pipeline.Scheduler
}

func NewServer(config *static.Config) *Server {
	runtime := models.NewRuntime()
	watcher := dynamic.NewConfigWatcher(config.Providers)

	server := &Server{
		watcher:     watcher,
		runtime:     runtime,
		stopChannel: make(chan bool),
		Bindings:    make(map[string]Binding),
		Schedulers:  make(map[string]*pipeline.Scheduler),
	}

	watcher.AddListener(func(o dynamic.Config) {
		server.updateBindings(o)
	})

	if config.Api.Dashboard {
		addr := fmt.Sprintf(":%v", config.Api.Port)
		b := api.NewBinding(addr, server.runtime)
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
	s.watcher.Close()
}

func (s *Server) updateBindings(config dynamic.Config) {
	switch c := config.(type) {
	case *openapi.Config:
		if _, ok := s.runtime.OpenApi[c.Info.Name]; !ok {
			s.runtime.OpenApi[c.Info.Name] = c
		}
		for _, server := range c.Servers {
			address := fmt.Sprintf(":%v", server.GetPort())
			binding, found := s.Bindings[address]
			if !found {
				binding = web.NewBinding(address, s.runtime.Metrics.AddRequest)
				s.Bindings[address] = binding
				binding.Start()
			}
			err := binding.Apply(c)
			if err != nil {
				log.Error(err.Error())
			}
		}
	case *ldap.Config:
		if b, ok := s.Bindings[c.Address]; !ok {
			lserver := ldapServer.NewServer(c)
			s.Bindings[c.Address] = lserver
			lserver.Start()
		} else {
			b.Apply(c)
		}
	case *asyncApi.Config:
		for _, server := range c.Servers {
			address := fmt.Sprintf(":%v", server.GetPort())
			binding, found := s.Bindings[address]
			if !found {
				binding = kafka.NewBinding(address, server.Bindings.Kafka, c)
				s.Bindings[address] = binding
				binding.Start()
			}
			err := binding.Apply(c)
			if err != nil {
				log.Error(err.Error())
			}
		}
	case *mokapi.Config:
		if scheduler, ok := s.Schedulers[c.ConfigPath]; ok {
			scheduler.Stop()
			if err := scheduler.Start(); err != nil {
				log.Errorf("unable to start scheduler with config %q: %v", c.ConfigPath, err)
			}
		} else {
			scheduler = pipeline.NewScheduler(c)
			s.Schedulers[c.ConfigPath] = scheduler
			if err := scheduler.Start(); err != nil {
				log.Errorf("unable to start scheduler with config %q: %v", c.ConfigPath, err)
			}
		}
	}
}
