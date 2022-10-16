package server

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	config "mokapi/config/dynamic/mail"
	engine "mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/server/service"
)

type SmtpServers map[string]*service.MailServer

type SmtpManager struct {
	Servers SmtpServers

	eventEmitter engine.EventEmitter
	certStore    *cert.Store
	app          *runtime.App
}

func NewSmtpManager(servers SmtpServers, emitter engine.EventEmitter, store *cert.Store, app *runtime.App) *SmtpManager {
	return &SmtpManager{
		eventEmitter: emitter,
		certStore:    store,
		app:          app,
		Servers:      servers,
	}
}

func (s SmtpManager) UpdateConfig(c *common.Config) {
	cfg, ok := c.Data.(*config.Config)
	if !ok {
		return
	}

	h := runtime.NewSmtpHandler(s.app.Monitor.Smtp, config.NewHandler(cfg, s.eventEmitter))

	if server, ok := s.Servers[cfg.Info.Name]; !ok {
		server = service.NewMailServer(cfg.Server, h)
		server.Start()
	}

	s.app.AddSmtp(cfg)

	log.Debugf("processed %v", c.Url.String())
}

func (s SmtpServers) Stop() {
	for _, server := range s {
		server.Stop()
	}
}
