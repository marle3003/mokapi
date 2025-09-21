package server

import (
	log "github.com/sirupsen/logrus"
	"mokapi/engine"
	"mokapi/runtime"
	"mokapi/safe"
)

type Server struct {
	app     *runtime.App
	watcher *ConfigWatcher
	kafka   *KafkaManager
	http    *HttpManager
	engine  *engine.Engine
	mail    *MailManager
	ldap    *LdapDirectoryManager

	pool     *safe.Pool
	stopChan chan bool
}

func NewServer(pool *safe.Pool, app *runtime.App, watcher *ConfigWatcher,
	kafka *KafkaManager, http *HttpManager, mail *MailManager, ldap *LdapDirectoryManager, engine *engine.Engine) *Server {
	return &Server{
		app:      app,
		watcher:  watcher,
		kafka:    kafka,
		http:     http,
		mail:     mail,
		ldap:     ldap,
		engine:   engine,
		pool:     pool,
		stopChan: make(chan bool, 1),
	}
}

func (s *Server) Start() error {
	s.engine.Start()
	if err := s.watcher.Start(s.pool); err != nil {
		return err
	}
	s.app.Monitor.Start(s.pool)

	<-s.stopChan
	log.Debug("stopping server")
	s.pool.Stop()
	s.kafka.Stop()
	s.http.Stop()
	s.mail.Stop()
	s.ldap.Stop()
	s.engine.Close()

	return nil
}

func (s *Server) Close() {
	close(s.stopChan)
}
