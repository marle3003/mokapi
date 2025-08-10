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

func NewSmtpServer(port string, handler smtp.Handler, store *cert.Store, tlsMode smtp.TlsMode) *SmtpServer {
	s := &SmtpServer{
		server: &smtp.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: handler,
			TlsMode: tlsMode,
			TLSConfig: &tls.Config{
				GetCertificate:     store.GetCertificate,
				InsecureSkipVerify: true,
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
		err := s.server.ListenAndServe()
		if !errors.Is(err, smtp.ErrServerClosed) {
			log.Error(err)
		}
	}()
}

func (s *SmtpServer) Stop() {
	s.server.Close()
}
