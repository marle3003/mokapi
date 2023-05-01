package server

import (
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/mail"
	engine "mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/smtp"
	"net/url"
)

type SmtpManager struct {
	servers      map[string]*smtp.Server
	app          *runtime.App
	eventEmitter engine.EventEmitter
	certStore    *cert.Store
}

func NewSmtpManager(app *runtime.App, eventEmitter engine.EventEmitter, store *cert.Store) *SmtpManager {
	return &SmtpManager{
		servers:      make(map[string]*smtp.Server),
		app:          app,
		eventEmitter: eventEmitter,
		certStore:    store,
	}
}

func (m *SmtpManager) UpdateConfig(c *common.Config) {
	cfg, ok := c.Data.(*mail.Config)
	if !ok {
		return
	}

	if server, ok := m.servers[cfg.Info.Name]; !ok {
		h := runtime.NewSmtpHandler(m.app.Monitor.Smtp, mail.NewHandler(cfg, m.eventEmitter))
		u, err := parseSmtpUrl(cfg.Server)
		if err != nil {
			log.Errorf("url syntax error %v: %v", c.Url, err.Error())
			return
		}
		log.Infof("adding new smtp host on %v", u)
		server = &smtp.Server{
			Addr: fmt.Sprintf(":%v", u.Port()),
			TLSConfig: &tls.Config{
				GetCertificate: m.certStore.GetCertificate,
			},
			Handler: h}

		m.servers[cfg.Info.Name] = server
		if u.Scheme == "smtps" {
			startServer(server.ListenAndServeTLS)
		} else {
			startServer(server.ListenAndServe)
		}
	} else {
	}
}

func startServer(f func() error) {
	go func() {
		err := f()
		if err != nil {
			log.Errorf("unable to start smtp server: %v", err)
			return
		}
	}()
}

func (m *SmtpManager) Stop() {
	for _, server := range m.servers {
		server.Close()
	}
}

func parseSmtpUrl(s string) (u *url.URL, err error) {
	u, err = url.Parse(s)
	if err != nil {
		return
	}

	port := u.Port()
	if len(port) == 0 {
		switch u.Scheme {
		case "smtps":
			port = "587"
		default:
			port = "25"
		}
		u.Host = fmt.Sprintf("%v:%v", u.Hostname(), port)
	}

	return
}
