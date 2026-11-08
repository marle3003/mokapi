package service

import (
	"crypto/tls"
	"fmt"
	"mokapi/server/cert"
	"mokapi/smtp"
	"net"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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

func (s *SmtpServer) Start() error {
	ln, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}

	go func() {
		err := s.server.Serve(ln)
		if !errors.Is(err, smtp.ErrServerClosed) {
			log.Error(err)
		}
	}()
	return nil
}

func (s *SmtpServer) Stop() {
	s.server.Close()
}
