package service

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/imap"
)

type ImapServer struct {
	server *imap.Server
}

func NewImapServer(port string, handler imap.Handler) *ImapServer {
	s := &ImapServer{
		server: &imap.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: handler,
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
