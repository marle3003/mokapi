package service

import (
	"crypto/tls"
	"fmt"
	"mokapi/imap"
	"mokapi/server/cert"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type ImapServer struct {
	server *imap.Server
}

func NewImapServer(port string, handler imap.Handler, store *cert.Store, tlsMode imap.TlsMode) *ImapServer {
	s := &ImapServer{
		server: &imap.Server{
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

func (s *ImapServer) Addr() string {
	return s.server.Addr
}

func (s *ImapServer) Start() {
	go func() {
		err := s.server.ListenAndServe()
		if !errors.Is(err, imap.ErrServerClosed) {
			log.Error(err)
		}
	}()
}

func (s *ImapServer) Stop() {
	s.server.Close()
}
