package smtp

import (
	"crypto/tls"
	"github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"
	config "mokapi/config/dynamic/smtp"
	"mokapi/models"
	"mokapi/providers/workflow"
	"mokapi/providers/workflow/event"
	"mokapi/providers/workflow/runtime"
)

type ReceivedMailHandler func(mail *models.MailMetric)

type EventHandler func(events event.Handler, options ...workflow.Options) (*runtime.Summary, error)

type Binding struct {
	server *smtp.Server
	config *config.Config

	close    chan bool
	received chan *models.MailMetric

	mh ReceivedMailHandler
	wh EventHandler
}

type backend struct {
	received chan *models.MailMetric
	wh       EventHandler
}

func NewBinding(c *config.Config, mh ReceivedMailHandler, getCertificate func(info *tls.ClientHelloInfo) (*tls.Certificate, error), wh EventHandler) *Binding {
	received := make(chan *models.MailMetric)
	b := &backend{received: received, wh: wh}
	s := smtp.NewServer(b)
	s.Addr = c.Address

	if c.Tls != nil {
		s.TLSConfig = &tls.Config{
			GetCertificate: getCertificate,
		}
	}

	return &Binding{
		config:   c,
		server:   s,
		received: received,
		close:    make(chan bool),
		mh:       mh,
		wh:       wh,
	}
}

func (b *Binding) Start() {
	go func() {
		log.Infof("starting smtp binding %v", b.config.Address)
		var err error
		if b.config.Tls != nil {
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
				b.mh(mail)
			case <-b.close:
				return
			}
		}
	}()
}

func (b *Binding) Stop() {
	b.close <- true
}

func (b *Binding) Apply(i interface{}) error {
	return nil
}

func (b backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	log.Debugf("smtp login with username %q", username)
	summary, err := b.wh(event.WithSmtpEvent(event.SmtpEvent{Login: true, Address: state.LocalAddr.String()}), workflow.WithContext("auth", &Login{Username: username, Password: password}))
	if err != nil {
		log.Errorf("error on smtp login: %v", err)
	}
	if summary == nil {
		log.Debugf("no actions found")
	} else {
		log.WithField("action summary", summary).Debugf("executed actions")
	}

	return newSession(b.received, b.wh, state), nil
}

func (b backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	log.Debug("smtp anonymous login")
	summary, err := b.wh(event.WithSmtpEvent(event.SmtpEvent{Login: true, Address: state.LocalAddr.String()}), workflow.WithContext("auth", &Login{Anonymous: true}))
	if err != nil {
		log.Errorf("error on smtp login: %v", err)
	}

	if summary == nil {
		log.Debugf("no actions found")
	} else {
		log.WithField("action summary", summary).Debugf("executed actions")
	}

	return newSession(b.received, b.wh, state), nil
}
