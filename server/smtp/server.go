package smtp

import (
	"crypto/tls"
	"fmt"
	"github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"
	config "mokapi/config/dynamic/smtp"
	"mokapi/engine/common"
	"mokapi/models"
	"mokapi/safe"
	"mokapi/server/cert"
	"net"
	"net/url"
)

type ReceivedMailHandler func(mail *models.MailMetric)

//type EventHandler func(events event.Handler, options ...workflow.Options) (*runtime.Summary, error)

type Server struct {
	server *smtp.Server
	config *config.Config

	close    chan bool
	received chan *models.MailMetric

	mh      ReceivedMailHandler
	emitter common.EventEmitter
	//wh EventHandler
	inShutdown safe.AtomicBool
}

type backend struct {
	received chan *models.MailMetric
	//wh       EventHandler
}

func New(c *config.Config, store *cert.Store, emitter common.EventEmitter) (*Server, error) {
	received := make(chan *models.MailMetric)
	b := &backend{received: received /*wh: wh*/}
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
		config:   c,
		server:   s,
		received: received,
		close:    make(chan bool),
		emitter:  emitter,
		//mh:       mh,
		//wh:       wh,
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
		log.Infof("start serving smtp on %v", l.Addr())
		err := s.server.Serve(l)
		if err != nil {
			log.Errorf("unable to start smtp server: %v", err)
			return
		}
	}()

	go func() {
		defer func() {
			s.Stop()
		}()

		for {
			select {
			case mail := <-s.received:
				if s.mh != nil {
					s.mh(mail)
				}
			case <-s.close:
				return
			}
		}
	}()
}

func (s *Server) Stop() {
	if s.inShutdown.IsSet() {
		return
	}
	s.inShutdown.SetTrue()

	err := s.server.Close()
	if err != nil {
		log.Errorf("unable to close smtp server: %v", err)
	}
	s.close <- true
}

func (s *Server) Update(config *config.Config) error {
	return nil
}

func (b backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	log.Debugf("smtp login with username %q", username)
	//summary, err := b.wh(event.WithSmtpEvent(event.SmtpEvent{Login: true, Address: state.LocalAddr.String()}), workflow.WithContext("auth", &Login{Username: username, Password: password}))
	//if err != nil {
	//	log.Errorf("error on smtp login: %v", err)
	//}
	//if summary == nil {
	//	log.Debugf("no actions found")
	//} else {
	//	log.WithField("action summary", summary).Debugf("executed actions")
	//}

	return newSession(b.received /*, b.wh*/, state), nil
}

func (b backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	log.Debug("smtp anonymous login")
	//summary, err := b.wh(event.WithSmtpEvent(event.SmtpEvent{Login: true, Address: state.LocalAddr.String()}), workflow.WithContext("auth", &Login{Anonymous: true}))
	//if err != nil {
	//	log.Errorf("error on smtp login: %v", err)
	//}
	//
	//if summary == nil {
	//	log.Debugf("no actions found")
	//} else {
	//	log.WithField("action summary", summary).Debugf("executed actions")
	//}

	return newSession(b.received /*, b.wh*/, state), nil
}
