package server

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	config "mokapi/config/dynamic/mail"
	engine "mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/smtp"
)

type SmtpServers map[string]*smtp.Server

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
		server = &smtp.Server{
			Addr:    "127.0.0.1:25",
			Handler: h,
		}
		err := server.ListenAndServe()
		if err != nil {
			log.Errorf("unable to start smtp server: %v", err)
			return
		}

	} else {
		server.Handler = h
	}

	s.app.AddSmtp(cfg)
}

func (s SmtpServers) Stop() {
	if len(s) > 0 {
		log.Debug("stopping smtp servers")
	}
	for _, server := range s {
		server.Close()
	}
}
