package smtp

import (
	"crypto/tls"
	"fmt"
	"github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"
	config "mokapi/config/dynamic/mail"
	"mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/runtime/events"
	"mokapi/server/cert"
	"net"
	"net/url"
)

type Server struct {
	app    *runtime.App
	server *smtp.Server
	config *config.Config

	close    chan bool
	received chan *Mail

	emitter common.EventEmitter
}

type backend struct {
	received chan *Mail
}

func New(c *config.Config, store *cert.Store, emitter common.EventEmitter, app *runtime.App) (*Server, error) {
	received := make(chan *Mail)
	b := &backend{received: received}
	s := smtp.NewServer(b)

	u, err := url.Parse(c.Server)
	if err != nil {
		return nil, err
	}

	s.Addr = u.Hostname()
	if len(u.Port()) == 0 {
		switch u.Scheme {
		case "smtp", "smtps":
			s.Addr += u.Scheme
		case "":
			s.Addr += ":smtp"
		default:
			return nil, fmt.Errorf("unsupported scheme: %v", u.Scheme)
		}
	} else {
		s.Addr += ":" + u.Port()
	}

	if u.Scheme == "smtps" || u.Port() == "587" {
		s.TLSConfig = &tls.Config{
			GetCertificate: store.GetCertificate,
		}
	}

	return &Server{
		app:      app,
		config:   c,
		server:   s,
		received: received,
		close:    make(chan bool),
		emitter:  emitter,
	}, nil
}

func (s *Server) Start() error {
	var l net.Listener
	var err error
	if s.server.TLSConfig != nil {
		l, err = tls.Listen("tcp", s.server.Addr, s.server.TLSConfig)
	} else {
		l, err = net.Listen("tcp", s.server.Addr)
	}
	if err != nil {
		return err
	}

	s.StartWith(l)
	return nil

}

func (s *Server) StartWith(l net.Listener) {
	go func() {
		log.Infof("starting smtp binding %v", s.config.Server)
		err := s.server.Serve(l)
		if err != nil {
			log.Errorf("unable to start smtp server: %v", err)
			return
		}
	}()

	go func() {
		defer func() {
			err := s.server.Close()
			if err != nil {
				log.Errorf("unable to close smtp server: %v", err)
			}
		}()

		for {
			select {
			case mail := <-s.received:
				log.Infof("recevied new mail on %v from %v with subject %v", l.Addr(), mail.From, mail.Subject)
				s.app.Monitor.Smtp.Mails.WithLabel(s.config.Info.Name).Add(1)
				events.Push(mail, events.NewTraits().WithNamespace("smtp").WithName(s.config.Info.Name))
				s.emitter.Emit("smtp", mail)
			case <-s.close:
				return
			}
		}
	}()
}

func (s *Server) Stop() {
	s.server.Close()
	s.close <- true
}

func (s *Server) Update(config *config.Config) error {
	return nil
}

func (b backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	return newSession(b.received, state), nil
}

func (b backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return newSession(b.received, state), nil
}
