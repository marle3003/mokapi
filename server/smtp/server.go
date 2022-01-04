package smtp

import (
	"crypto/tls"
	"fmt"
	"github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"
	config "mokapi/config/dynamic/smtp"
	"mokapi/models"
	"mokapi/server/cert"
	"net/url"
)

type ReceivedMailHandler func(mail *models.MailMetric)

//type EventHandler func(events event.Handler, options ...workflow.Options) (*runtime.Summary, error)

type Server struct {
	server *smtp.Server
	config *config.Config

	close    chan bool
	received chan *models.MailMetric

	mh ReceivedMailHandler
	//wh EventHandler
}

type backend struct {
	received chan *models.MailMetric
	//wh       EventHandler
}

func New(c *config.Config, store *cert.Store) (*Server, error) {
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
		//mh:       mh,
		//wh:       wh,
	}, nil
}

func (b *Server) Start() {
	go func() {
		log.Infof("starting smtp binding %v", b.config.Server)
		var err error
		if b.server.TLSConfig != nil {
			err = b.server.ListenAndServeTLS()
		} else {
			err = b.server.ListenAndServe()
		}
		if err != nil {
			log.Errorf("unable to start smtp server: %v", err)
			return
		}
	}()

	go func() {
		defer func() {
			err := b.server.Close()
			if err != nil {
				log.Errorf("unable to close smtp server: %v", err)
			}
		}()

		for {
			select {
			case mail := <-b.received:
				if b.mh != nil {
					b.mh(mail)
				}
			case <-b.close:
				return
			}
		}
	}()
}

func (b *Server) Stop() {
	b.server.Close()
	b.close <- true
}

func (b *Server) Update(config *config.Config) error {
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
