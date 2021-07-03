package server

import (
	"crypto/tls"
	"crypto/x509"
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
	"mokapi/server/cert"
	"mokapi/server/kafka"
	ldapServer "mokapi/server/ldap"
	"mokapi/server/web"
	"path/filepath"
	"strings"
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
	store              *cert.Store

	scheduler *workflow.Scheduler
	Bindings  map[string]Binding
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
		scheduler:          workflow.NewScheduler(),
	}

	watcher.AddListener(func(o dynamic.Config) {
		server.updateConfigs(o)
	})

	if config.Api.Dashboard {
		addr := fmt.Sprintf(":%v", config.Api.Port)
		b := api.NewBinding(addr, server.runtime)
		server.Bindings[addr] = b
	}

	var err error
	server.store, err = cert.NewStore(config)
	if err != nil {
		log.Errorf("unable to create certificate store: %v", err)
	}

	return server
}

func (s *Server) Start() {
	for _, b := range s.Bindings {
		b.Start()
	}
	err := s.watcher.Start()
	if err != nil {
		log.Errorf("unable to start server: %v", err.Error())
	}
	s.startMetricUpdater()
	s.scheduler.Start()
}

func (s *Server) Wait() {
	<-s.stopChannel
}

func (s *Server) Stop() {
	s.watcher.Close()
	s.scheduler.Stop()
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
		s.config[c.ConfigPath] = c
		if len(c.Workflows) == 0 {
			log.Debugf("no workflows found in %v", c.ConfigPath)
		}
		for _, w := range c.Workflows {
			log.Debugf("adding workflow %q", w.Name)
		}
		err := s.scheduler.AddOrUpdate(c.ConfigPath, c.Workflows, workflow.WithWorkingDirectory(filepath.Dir(c.ConfigPath)))
		if err != nil {
			log.Errorf("unable to add scheduler for workflows %q", c.ConfigPath)
			return
		}
		for _, cer := range c.Certificates {
			err := s.appendCertificate(cer, filepath.Dir(c.ConfigPath))
			if err != nil {
				log.Errorf("unable to add certificate from %q: %v", c.ConfigPath, err)
			}
		}
		log.Infof("updated config %q", c.ConfigPath)
	case *openapi.Config:
		if _, ok := s.runtime.OpenApi[c.Info.Name]; !ok {
			s.runtime.OpenApi[c.Info.Name] = c
			s.runtime.Metrics.OpenApi[c.Info.Name] = &models.ServiceMetric{Name: c.Info.Name}
		}
		for _, server := range c.Servers {
			address := fmt.Sprintf(":%v", server.GetPort())
			binding, found := s.Bindings[address].(*web.Binding)
			if !found {
				if strings.HasPrefix(server.Url, "https://") {
					binding = web.NewBindingWithTls(address, s.runtime.Metrics.AddRequest, s.triggerHandler, s.store.GetCertificate)
				} else {
					binding = web.NewBinding(address, s.runtime.Metrics.AddRequest, s.triggerHandler)
				}
				s.Bindings[address] = binding
				binding.Start()
			}
			err := binding.Apply(c)
			if err != nil {
				log.Errorf("error on updating %q: %v", c.ConfigPath, err.Error())
				return
			}
		}
		log.Infof("updated config %q", c.ConfigPath)
	case *ldap.Config:
		if b, ok := s.Bindings[c.Address]; !ok {
			s.runtime.Ldap[c.Info.Name] = c

			lserver := ldapServer.NewServer(c)
			s.Bindings[c.Address] = lserver
			lserver.Start()
		} else {
			err := b.Apply(c)
			if err != nil {
				log.Errorf("error on updating %q: %v", c.ConfigPath, err.Error())
				return
			}
		}
		log.Infof("updated config %q", c.ConfigPath)
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
			log.Errorf("error on updating %q: %v", c.ConfigPath, err.Error())
			return
		}
		log.Infof("updated config %q", c.ConfigPath)
	}
}

func (s *Server) triggerHandler(event event.Handler, options ...workflow.Options) *runtime.Summary {
	summary := &runtime.Summary{}
	for _, c := range s.config {
		o := append(options, workflow.WithWorkingDirectory(filepath.Dir(c.ConfigPath)))
		s := workflow.Run(c.Workflows, event, o...)
		summary.Workflows = append(summary.Workflows, s.Workflows...)
	}

	return summary
}

func (s *Server) appendCertificate(c mokapi.Certificate, currentDir string) error {
	certContent, err := c.CertFile.Read(currentDir)
	if err != nil {
		return err
	}

	keyContent, err := c.KeyFile.Read(currentDir)
	if err != nil {
		return err
	}

	tlsCert, err := tls.X509KeyPair(certContent, keyContent)
	if err != nil {
		return err
	}

	cer, _ := x509.ParseCertificate(tlsCert.Certificate[0])
	s.store.AddCertificate(cer.Subject.CommonName, &tlsCert)
	for _, n := range cer.DNSNames {
		s.store.AddCertificate(n, &tlsCert)
	}
	return nil
}
