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
	"mokapi/providers/workflow"
	"mokapi/providers/workflow/event"
	"mokapi/providers/workflow/runtime"
	"mokapi/server/api"
	"mokapi/server/kafka"
	ldapServer "mokapi/server/ldap"
	"mokapi/server/web"
	"path/filepath"
	"time"
)

type Binding interface {
	Start()
	Stop()
	Apply(interface{}) error
}

type Server struct {
	watcher            *dynamic.ConfigWatcher
	runtime            *models.Runtime
	stopChannel        chan bool
	stopMetricsUpdater chan bool
	config             map[string]*mokapi.Config

	Bindings map[string]Binding
}

func NewServer(config *static.Config) *Server {
	watcher := dynamic.NewConfigWatcher(config.Providers)

	server := &Server{
		watcher:            watcher,
		runtime:            models.NewRuntime(),
		stopChannel:        make(chan bool),
		stopMetricsUpdater: make(chan bool),
		Bindings:           make(map[string]Binding),
		config:             make(map[string]*mokapi.Config),
	}

	watcher.AddListener(func(o dynamic.Config) {
		server.updateConfigs(o)
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
	s.startMetricUpdater()
}

func (s *Server) Wait() {
	<-s.stopChannel
}

func (s *Server) Stop() {
	s.watcher.Close()
}

func (s *Server) startMetricUpdater() {
	go func() {
		ticker := time.NewTicker(time.Duration(5) * time.Second)

		for {
			select {
			case <-ticker.C:
				s.runtime.Metrics.Update()
				for _, e := range s.Bindings {
					switch b := e.(type) {
					case *kafka.Binding:
						b.UpdateMetrics(s.runtime.Metrics.Kafka)
					}
				}
			case <-s.stopMetricsUpdater:
				return
			}
		}
	}()
}

func (s *Server) updateConfigs(config dynamic.Config) {
	switch c := config.(type) {
	case *mokapi.Config:
		s.config[c.Name] = c
	case *openapi.Config:
		if _, ok := s.runtime.OpenApi[c.Info.Name]; !ok {
			s.runtime.OpenApi[c.Info.Name] = c
			s.runtime.Metrics.OpenApi[c.Info.Name] = &models.ServiceMetric{Name: c.Info.Name}
		}
		for _, server := range c.Servers {
			address := fmt.Sprintf(":%v", server.GetPort())
			binding, found := s.Bindings[address]
			if !found {
				binding = web.NewBinding(address, s.runtime.Metrics.AddRequest, s.triggerHandler)
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
			s.runtime.Ldap[c.Info.Name] = c

			lserver := ldapServer.NewServer(c)
			s.Bindings[c.Address] = lserver
			lserver.Start()
		} else {
			b.Apply(c)
		}
	case *asyncApi.Config:
		if _, ok := s.runtime.AsyncApi[c.Info.Name]; !ok {
			s.runtime.AsyncApi[c.Info.Name] = c
		}

		binding, found := s.Bindings[c.Info.Name]
		if !found {
			b := kafka.NewBinding(c, s.runtime.Metrics.AddMessage)
			binding = b
			s.Bindings[c.Info.Name] = b
			b.Start()
			b.UpdateMetrics(s.runtime.Metrics.Kafka)
		}
		err := binding.Apply(c)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func (s *Server) triggerHandler(event event.Handler, options ...runtime.WorkflowOptions) *runtime.Summary {
	for _, c := range s.config {
		o := append(options, runtime.WithWorkingDirectory(filepath.Dir(c.ConfigPath)))
		return workflow.Run(c.Workflows, event, o...)
	}

	return nil
}
