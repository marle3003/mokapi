package server

import (
	"mokapi/engine"
	"mokapi/runtime"
	"mokapi/safe"

	log "github.com/sirupsen/logrus"
)

type Server struct {
	app       *runtime.App
	watcher   *ConfigWatcher
	kafka     *KafkaManager
	http      *HttpManager
	engine    *engine.Engine
	mail      *MailManager
	ldap      *LdapDirectoryManager
	mqtt      *MqttManager
	websocket *WebsocketManager

	pool     *safe.Pool
	stopChan chan bool
}

func NewServer(pool *safe.Pool, app *runtime.App, watcher *ConfigWatcher,
	kafka *KafkaManager, http *HttpManager, mail *MailManager, ldap *LdapDirectoryManager,
	mqtt *MqttManager, websocket *WebsocketManager, engine *engine.Engine) *Server {
	return &Server{
		app:       app,
		watcher:   watcher,
		kafka:     kafka,
		http:      http,
		mail:      mail,
		ldap:      ldap,
		mqtt:      mqtt,
		websocket: websocket,
		engine:    engine,
		pool:      pool,
		stopChan:  make(chan bool, 1),
	}
}

func (s *Server) Start() error {
	s.app.Start(s.pool)
	s.engine.Start()
	if err := s.watcher.Start(s.pool); err != nil {
		return err
	}
	s.app.Monitor.Start(s.pool)

	<-s.stopChan
	log.Debug("stopping server")
	s.app.Stop()
	s.kafka.Stop()
	s.http.Stop()
	s.mail.Stop()
	s.ldap.Stop()
	s.mqtt.Stop()
	s.websocket.Stop()
	s.engine.Close()
	s.pool.Stop()

	return nil
}

func (s *Server) Close() {
	close(s.stopChan)
}

// startAndLogError keeps a server registered when its port can't be bound, so a later update doesn't start a second one.
func startAndLogError(s interface{ Start() error }) {
	if err := s.Start(); err != nil {
		log.Error(err)
	}
}
