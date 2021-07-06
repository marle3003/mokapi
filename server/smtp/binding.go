package smtp

import (
	"crypto/tls"
	"github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"
	config "mokapi/config/dynamic/smtp"
	"mokapi/models"
)

type ReceivedMailHandler func(mail *models.Mail)

type Binding struct {
	server *smtp.Server
	config *config.Config

	close    chan bool
	received chan *models.Mail

	mh ReceivedMailHandler
}

type backend struct {
	received chan *models.Mail
}

func NewBinding(c *config.Config, mh ReceivedMailHandler, getCertificate func(info *tls.ClientHelloInfo) (*tls.Certificate, error)) *Binding {
	received := make(chan *models.Mail)
	b := &backend{received: received}
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
	log.Infof("smtp login with username %q", username)

	return newSession(b.received), nil
}

func (b backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	log.Info("smtp anonymous login")

	return newSession(b.received), nil
}
