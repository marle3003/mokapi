package service

import (
	"crypto/tls"
	"fmt"
	"mokapi/imap"
	"mokapi/server/cert"
	"net"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type ImapServer struct {
	server *imap.Server
	ln     net.Listener
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

func (s *ImapServer) Start() error {
	ln, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}
	s.ln = ln

	go func() {
		err := s.server.Serve(ln)
		if !errors.Is(err, imap.ErrServerClosed) {
			log.Error(err)
		}
	}()
	return nil
}

func (s *ImapServer) Stop() {
	s.server.Close()
	// Serve only learns about the listener once its goroutine runs, so Close alone can leave the port open.
	if s.ln != nil {
		_ = s.ln.Close()
	}
}
