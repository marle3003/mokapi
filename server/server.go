package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/ldap"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/static"
	"mokapi/models"
	"mokapi/server/api"
	"mokapi/server/kafka"
	ldapServer "mokapi/server/ldap"
	"mokapi/server/web"
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

	Bindings map[string]Binding
}

func NewServer(config *static.Config) *Server {
	runtime := models.NewRuntime()
	watcher := dynamic.NewConfigWatcher(config.Providers)

	server := &Server{
		watcher:            watcher,
		runtime:            runtime,
		stopChannel:        make(chan bool),
		stopMetricsUpdater: make(chan bool),
		Bindings:           make(map[string]Binding),
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
				for k, e := range s.Bindings {
					switch b := e.(type) {
					case *kafka.Binding:
						b.UpdateMetrics(s.runtime.Metrics.Kafka[k])
					}
				}
			case <-s.stopMetricsUpdater:
				return
			}
		}
	}()
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
		if _, ok := s.runtime.AsyncApi[c.Info.Name]; !ok {
			s.runtime.AsyncApi[c.Info.Name] = c
			s.runtime.Metrics.Kafka[c.Info.Name] = &models.KafkaMetrics{TopicSizes: make(map[string]int64)}
		}

		binding, found := s.Bindings[c.Info.Name]
		if !found {
			binding = kafka.NewBinding(c)
			s.Bindings[c.Info.Name] = binding
			binding.Start()
		}
		err := binding.Apply(c)
		if err != nil {
			log.Error(err.Error())
		}
	}
}
