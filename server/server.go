package server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/engine"
	"mokapi/runtime"
	"mokapi/safe"
	"os"
	"runtime/pprof"
)

type Server struct {
	app     *runtime.App
	watcher *dynamic.ConfigWatcher
	kafka   KafkaClusters
	http    HttpServers
	engine  *engine.Engine
	mail    SmtpServers
	ldap    LdapDirectories

	pool     *safe.Pool
	stopChan chan bool
}

func NewServer(pool *safe.Pool, app *runtime.App, watcher *dynamic.ConfigWatcher,
	kafka KafkaClusters, http HttpServers, mail SmtpServers, ldap LdapDirectories, engine *engine.Engine) *Server {
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

func (s *Server) StartAsync(ctx context.Context) {
	go func() {
		s.Start(ctx)
	}()
}

func (s *Server) Start(ctx context.Context) error {
	s.engine.Start()
	if err := s.watcher.Start(s.pool); err != nil {
		return err
	}
	s.app.Monitor.Start(s.pool)

	<-ctx.Done()
	log.Debug("stopping server")
	s.pool.Stop()
	s.kafka.Stop()
	s.http.Stop()
	s.mail.Stop()
	s.ldap.Stop()
	s.engine.Close()
	s.stopChan <- true

	pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)

	return nil
}

func (s *Server) Close() {
	<-s.stopChan
}
