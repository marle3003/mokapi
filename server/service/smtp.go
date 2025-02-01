package service

import (
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/server/cert"
	"mokapi/smtp"
)

type SmtpServer struct {
	server *smtp.Server
}

func NewSmtpServer(port string, handler smtp.Handler) *SmtpServer {
	s := &SmtpServer{
		server: &smtp.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: handler,
		},
	}
	return s
}

func NewSmtpServerTls(port string, handler smtp.Handler, store *cert.Store) *SmtpServer {
	s := &SmtpServer{
		server: &smtp.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: handler,
			TLSConfig: &tls.Config{
				GetCertificate: store.GetCertificate,
			},
		},
	}
	return s
}

func (s *SmtpServer) Addr() string {
	return s.server.Addr
}

func (s *SmtpServer) Start() {
	go func() {
		var err error
		if s.server.TLSConfig == nil {
			err = s.server.ListenAndServe()
		} else {
			err = s.server.ListenAndServeTLS()
		}
		if !errors.Is(err, smtp.ErrServerClosed) {
			log.Error(err)
		}
	}()
}

func (s *SmtpServer) Stop() {
	s.server.Close()
}
