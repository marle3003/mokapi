package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/ldap"
	"mokapi/config/dynamic/mokapi"
	"mokapi/config/dynamic/openapi"
	"mokapi/config/dynamic/script"
	"mokapi/config/dynamic/smtp"
	"mokapi/config/static"
	"mokapi/engine"
	"mokapi/kafka"
	"mokapi/models"
	"mokapi/safe"
	"mokapi/server/api"
	"mokapi/server/cert"
	ldapServer "mokapi/server/ldap"
	smtpServer "mokapi/server/smtp"
	"mokapi/server/web"
	"path/filepath"
	"strings"
	"sync"
)

type ConfigWatcher interface {
	AddListener(listener func(c *common.File))
}

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

	engine *engine.Engine

	kafkaClusters map[string]*kafka.Cluster

	Bindings map[string]Binding
	mutex    sync.RWMutex
	pool     *safe.Pool
}

func NewServer(config *static.Config) *Server {
	watcher := dynamic.NewConfigWatcher(config)

	server := &Server{
		watcher:            watcher,
		runtime:            models.NewRuntime(),
		stopChannel:        make(chan bool),
		stopMetricsUpdater: make(chan bool),
		Bindings:           make(map[string]Binding),
		config:             make(map[string]*mokapi.Config),
		engine:             engine.New(watcher),
		kafkaClusters:      make(map[string]*kafka.Cluster),
		pool:               safe.NewPool(context.Background()),
	}

	watcher.AddListener(func(o *common.File) {
		server.updateConfigs(o)
	})

	if config.Api.Dashboard {
		addr := fmt.Sprintf(":%v", config.Api.Port)
		b := api.NewBinding(addr, server.runtime, config.Api.Path)
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
	err := s.watcher.Start(s.pool)
	if err != nil {
		log.Errorf("unable to start server: %v", err.Error())
	}
	s.startMetricUpdater()
	s.engine.Start()
}

func (s *Server) Wait() {
	<-s.stopChannel
}

func (s *Server) Stop() {
	s.engine.Close()
}

func (s *Server) startMetricUpdater() {
	go func() {
		//ticker := time.NewTicker(time.Duration(5) * time.Second)

		//for {
		//	select {
		//	case <-ticker.C:
		//		s.mutex.Lock()
		//		s.runtime.Metrics.Update()
		//		for _, e := range s.Bindings {
		//			switch b := e.(type) {
		//			case *kafka.Binding:
		//				b.UpdateMetrics(s.runtime.Metrics.Kafka)
		//			}
		//		}
		//		s.mutex.Unlock()
		//	case <-s.stopMetricsUpdater:
		//		return
		//	}
		//}
	}()
}

func (s *Server) updateConfigs(config *common.File) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch c := config.Data.(type) {
	case *script.Script:
		err := s.engine.AddScript(c.Filename, c.Code)
		if err != nil {
			log.Errorf("unable to execute script %v: %v", config.Url.String(), err)
		}
	case *mokapi.Config:
		s.config[c.ConfigPath] = c
		for _, cer := range c.Certificates {
			err := s.appendCertificate(cer, filepath.Dir(c.ConfigPath))
			if err != nil {
				log.Errorf("unable to add certificate from %v: %v", config.Url.String(), err)
			}
		}
		log.Infof("processed config %v", config.Url.String())
	case *openapi.Config:
		if err := c.Validate(); err != nil {
			log.Warnf("validation error %v: %v", config.Url, err)
			return
		}

		if _, ok := s.runtime.OpenApi[c.Info.Name]; !ok {
			s.runtime.OpenApi[c.Info.Name] = c
			s.runtime.Metrics.OpenApi[c.Info.Name] = &models.ServiceMetric{Name: c.Info.Name}
		}
		for _, server := range c.Servers {
			_, port, _, err := web.ParseAddress(server.Url)
			if err != nil {
				log.Errorf("%v: %v", err, config.Url)
				continue
			}
			address := fmt.Sprintf(":%v", port)
			binding, found := s.Bindings[address].(*web.Binding)
			if !found {
				if strings.HasPrefix(strings.ToLower(server.Url), "https://") {
					binding = web.NewBindingWithTls(address, s.runtime.Metrics.AddRequest, s.Run, s.store.GetCertificate)
				} else {
					binding = web.NewBinding(address, s.runtime.Metrics.AddRequest, s.Run)
				}
				s.Bindings[address] = binding
				binding.Start()
			}
			err = binding.Apply(c)
			if err != nil {
				log.Errorf("error on updating %v: %v", config.Url.String(), err.Error())
				return
			}
		}
		log.Infof("processed config %v", config.Url.String())
	case *ldap.Config:
		if b, ok := s.Bindings[c.Address]; !ok {
			s.runtime.Ldap[c.Info.Name] = c

			lserver := ldapServer.NewServer(c)
			s.Bindings[c.Address] = lserver
			lserver.Start()
		} else {
			err := b.Apply(c)
			if err != nil {
				log.Errorf("error on updating %v: %v", config.Url.String(), err.Error())
				return
			}
		}
		log.Infof("processed config %v", config.Url.String())
	case *asyncApi.Config:
		s.updateAsyncConfig(c)
		//if err := c.Validate(); err != nil {
		//	log.Warnf("validation error %v: %v", config.Url, err)
		//	return
		//}
		//
		//if _, ok := s.runtime.AsyncApi[c.Info.Name]; !ok {
		//	s.runtime.AsyncApi[c.Info.Name] = c
		//}
		//
		//binding, found := s.Bindings[c.Info.Name]
		//if !found {
		//	b := kafka.NewBinding(s.runtime.Metrics.AddMessage)
		//	binding = b
		//	s.Bindings[c.Info.Name] = b
		//	b.Start()
		//	b.UpdateMetrics(s.runtime.Metrics.Kafka)
		//}
		//err := binding.Apply(c)
		//if err != nil {
		//	log.Errorf("error on updating %v: %v", config.Url.String(), err.Error())
		//	return
		//}
		//log.Infof("processed config %v", config.Url.String())
	case *smtp.Config:
		if _, ok := s.runtime.Smtp[c.Name]; !ok {
			s.runtime.Smtp[c.Name] = c
		}

		_, found := s.Bindings[c.Address]
		if !found {
			b := smtpServer.NewBinding(c, s.runtime.Metrics.AddMail, s.store.GetCertificate /*, s.triggerHandler*/)
			b.Start()
			s.Bindings[c.Address] = b
		}
	}
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

func (s *Server) writeKafkaMessage(broker, topic string, partition int, key, message, header interface{}) (interface{}, interface{}, error) {
	//s.mutex.RLock()
	//defer s.mutex.RUnlock()
	//
	//for _, c := range s.Bindings {
	//	if b, ok := c.(*kafka.Binding); ok {
	//		// if empty broker, try first binding
	//		if len(broker) == 0 {
	//			return b.AddMessage(topic, partition, key, message, header)
	//		}
	//
	//		if b.HasBroker(broker) {
	//			return b.AddMessage(topic, partition, key, message, header)
	//		}
	//	}
	//}
	//
	//return nil, nil, fmt.Errorf("no broker found at %v", broker)
	return nil, nil, fmt.Errorf("not implemented")
}
