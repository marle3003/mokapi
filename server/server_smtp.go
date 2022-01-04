package server

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	config "mokapi/config/dynamic/smtp"
	"mokapi/server/cert"
	"mokapi/server/smtp"
)

type SmtpServers map[string]*smtp.Server

func (s SmtpServers) UpdateConfig(file *common.File, store *cert.Store) {
	cfg, ok := file.Data.(*config.Config)
	if !ok {
		return
	}

	if server, ok := s[cfg.Name]; !ok {
		server, err := smtp.New(cfg, store)
		if err != nil {
			log.Errorf("unable to start smtp server: %v", err)
			return
		}
		server.Start()
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
