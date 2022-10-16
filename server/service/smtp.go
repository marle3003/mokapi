package service

import (
	log "github.com/sirupsen/logrus"
	"mokapi/smtp"
)

type MailServer struct {
	server *smtp.Server
}

func NewMailServer(addr string, handler smtp.Handler) *MailServer {
	return &MailServer{
		server: &smtp.Server{Addr: addr, Handler: handler}}
}

func (m *MailServer) Start() {
	log.Infof("adding new smtp server on %v", m.server.Addr)
	go func() {
		err := m.server.ListenAndServe()
		if err != smtp.ErrServerClosed {
			log.Errorf("unable to start smtp server on %v: %v", m.server.Addr, err)
		}
	}()
}

func (m *MailServer) Stop() {
	m.server.Close()
}
