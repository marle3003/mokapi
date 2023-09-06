package server

import (
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/mail"
	engine "mokapi/engine/common"
	"mokapi/imap"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/smtp"
	"net/url"
	"sync"
)

type MailServer interface {
	Close()
}

type SmtpManager struct {
	servers      map[string][]MailServer
	app          *runtime.App
	eventEmitter engine.EventEmitter
	certStore    *cert.Store
	m            sync.Mutex
}

func NewSmtpManager(app *runtime.App, eventEmitter engine.EventEmitter, store *cert.Store) *SmtpManager {
	return &SmtpManager{
		servers:      make(map[string][]MailServer),
		app:          app,
		eventEmitter: eventEmitter,
		certStore:    store,
	}
}

func (m *SmtpManager) UpdateConfig(c *common.Config) {
	if !runtime.IsSmtpConfig(c) {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	cfg := m.app.AddSmtp(c)

	if _, ok := m.servers[cfg.Info.Name]; !ok {
		if len(cfg.Server) > 0 {
			log.Warnf("Deprecated server configuration. Please use smtp configuration")
			cfg.Servers = append(cfg.Servers, mail.Server{Url: cfg.Server})
		}

		var err error
		m.servers[cfg.Info.Name], err = m.newMailServers(cfg)
		if err != nil {
			log.Errorf(err.Error())
		}
		/*if m.certStore != nil {
			server.TLSConfig = &tls.Config{
				GetCertificate: m.certStore.GetCertificate,
			}
		}

		m.servers[cfg.Info.Name] = server
		if u.Scheme == "smtps" {
			startServer(server.ListenAndServeTLS)
		} else {
			startServer(server.ListenAndServe)
		}*/
	} else {
	}
}

func (m *SmtpManager) newMailServers(cfg *runtime.SmtpInfo) ([]MailServer, error) {
	var servers []MailServer
	h := cfg.Handler(m.app.Monitor.Smtp, m.eventEmitter)
	for _, server := range cfg.Servers {
		u, err := url.Parse(server.Url)
		if err != nil {
			return nil, err
		}
		switch u.Scheme {
		case "smtp":
			port := u.Port()
			if len(port) == 0 {
				port = "25"
			}
			server := &smtp.Server{
				Addr:    fmt.Sprintf(":%v", port),
				Handler: h,
			}
			log.Infof("adding new smtp host on :%v", port)
			startServer(server.ListenAndServe)
			servers = append(servers, server)
		case "smtps":
			port := u.Port()
			if len(port) == 0 {
				port = "587"
			}
			log.Infof("adding new smtps host on %v", port)
			server := &smtp.Server{
				Addr:    fmt.Sprintf(":%v", port),
				Handler: h,
				TLSConfig: &tls.Config{
					GetCertificate: m.certStore.GetCertificate,
				},
			}
			startServer(server.ListenAndServeTLS)
			servers = append(servers, server)
		case "imap":
			port := u.Port()
			if len(port) == 0 {
				port = "143"
			}
			log.Infof("adding new imap host on %v", port)
			server := &imap.Server{
				Addr:    fmt.Sprintf(":%v", port),
				Handler: h,
			}
			startServer(server.ListenAndServe)
			servers = append(servers, server)
		}
	}
	return servers, nil
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
	for _, ms := range m.servers {
		for _, server := range ms {
			server.Close()
		}
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
