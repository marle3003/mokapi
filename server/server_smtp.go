package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/mail"
	engine "mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/server/service"
	"net/url"
	"slices"
	"sync"
)

type MailServer interface {
	Stop()
	Start()
}

type SmtpManager struct {
	servers      map[string]map[string]MailServer
	app          *runtime.App
	eventEmitter engine.EventEmitter
	certStore    *cert.Store
	m            sync.Mutex
}

func NewSmtpManager(app *runtime.App, eventEmitter engine.EventEmitter, store *cert.Store) *SmtpManager {
	return &SmtpManager{
		servers:      make(map[string]map[string]MailServer),
		app:          app,
		eventEmitter: eventEmitter,
		certStore:    store,
	}
}

func (m *SmtpManager) UpdateConfig(e dynamic.ConfigEvent) {
	cfg, ok := runtime.IsSmtpConfig(e.Config)
	if !ok {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	info := m.app.Mail.Get(cfg.Info.Name)
	if e.Event == dynamic.Delete {
		m.app.Mail.Remove(e.Config)
		if info.Config == nil {
			m.removeService(cfg.Info.Name)
		}
		return
	} else if info == nil {
		info = m.app.Mail.Add(e.Config)
	} else {
		oldServers := info.Servers
		info.AddConfig(e.Config)
		m.cleanupRemovedServers(info, oldServers)
	}

	err := m.startServers(info)
	if err != nil {
		log.Errorf("starting '%v' failed: %v", cfg.Info.Name, err)
	}
}

func (m *SmtpManager) startServers(cfg *runtime.MailInfo) error {
	servers, ok := m.servers[cfg.Info.Name]
	if !ok {
		servers = map[string]MailServer{}
		m.servers[cfg.Info.Name] = servers
	}
	h := cfg.Handler(m.app.Monitor.Smtp, m.eventEmitter)
	for _, server := range cfg.Servers {
		u, err := url.Parse(server.Url)
		if err != nil {
			return err
		}

		switch u.Scheme {
		case "smtp":
			port := u.Port()
			if len(port) == 0 {
				port = "25"
			}
			if _, ok = servers[port]; ok {
				continue
			}

			log.Infof("adding new smtp host on :%v", port)
			s := service.NewSmtpServer(port, h)
			s.Start()
			servers[port] = s
		case "smtps":
			port := u.Port()
			if len(port) == 0 {
				port = "587"
			}

			addr := fmt.Sprintf(":%v", port)
			if _, ok = servers[addr]; ok {
				continue
			}

			log.Infof("adding new SMTPS host on %v", port)
			s := service.NewSmtpServerTls(port, h, m.certStore)
			s.Start()
			servers[port] = s
		case "imap":
			port := u.Port()
			if len(port) == 0 {
				port = "143"
			}

			addr := fmt.Sprintf(":%v", port)
			if _, ok = servers[addr]; ok {
				continue
			}

			log.Infof("adding new IMAP host on %v", port)
			s := service.NewImapServer(port, h)
			s.Start()
			servers[port] = s
		}
	}
	return nil
}

func (m *SmtpManager) Stop() {
	for _, ms := range m.servers {
		for _, server := range ms {
			server.Stop()
		}
	}
}

func (m *SmtpManager) removeService(name string) {
	servers := m.servers[name]
	for _, server := range servers {
		switch s := server.(type) {
		case *service.SmtpServer:
			log.Infof("removing service '%v' on IMAP binding %v", name, s.Addr())
		case *service.ImapServer:
			log.Infof("removing service '%v' on SMTP binding %v", name, s.Addr())
		}
		server.Stop()
	}
}

func (m *SmtpManager) cleanupRemovedServers(cfg *runtime.MailInfo, old []mail.Server) {
	for _, o := range old {
		if !slices.ContainsFunc(cfg.Servers, func(s mail.Server) bool {
			return s.Url == o.Url
		}) {
			u, err := url.Parse(o.Url)
			if err != nil {
				continue
			}
			addr := fmt.Sprintf(":%v", u.Port())

			servers := m.servers[cfg.Info.Name]
			for _, server := range servers {
				switch s := server.(type) {
				case *service.SmtpServer:
					if s.Addr() == addr {
						log.Infof("removing '%v' on SMTP binding %v", cfg.Info.Name, addr)
						s.Stop()
						delete(m.servers, addr)
					}
				case *service.ImapServer:
					if s.Addr() == addr {
						log.Infof("removing '%v' on IMAP binding %v", cfg.Info.Name, addr)
						s.Start()
						delete(m.servers, addr)
					}
				}
			}
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
