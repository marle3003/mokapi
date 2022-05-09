package server

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	config "mokapi/config/dynamic/smtp"
	engine "mokapi/engine/common"
	"mokapi/server/cert"
	"mokapi/server/smtp"
)

type SmtpServers map[string]*smtp.Server

func (s SmtpServers) UpdateConfig(c *common.Config, store *cert.Store, emitter engine.EventEmitter) {
	cfg, ok := c.Data.(*config.Config)
	if !ok {
		return
	}

	if server, ok := s[cfg.Name]; !ok {
		server, err := smtp.New(cfg, store, emitter)
		if err != nil {
			log.Errorf("unable to start smtp server: %v", err)
			return
		}
		err = server.Start()
		if err != nil {
			log.Errorf("unable to start smtp server: %v", err)
			return
		}
	} else {
		if err := server.Update(cfg); err != nil {
			log.Error(err)
		}
	}
}

func (s SmtpServers) Stop() {
	for _, server := range s {
		server.Stop()
	}
}
