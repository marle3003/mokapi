package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	engine "mokapi/engine/common"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/server/service"
	"net"
	"sync"
)

type MailServer interface {
	Stop()
	Start()
}

type MailManager struct {
	servers      map[string]map[string]MailServer
	app          *runtime.App
	eventEmitter engine.EventEmitter
	certStore    *cert.Store
	m            sync.Mutex
}

func NewMailManager(app *runtime.App, eventEmitter engine.EventEmitter, store *cert.Store) *MailManager {
	return &MailManager{
		servers:      make(map[string]map[string]MailServer),
		app:          app,
		eventEmitter: eventEmitter,
		certStore:    store,
	}
}

func (m *MailManager) UpdateConfig(e dynamic.ConfigEvent) {
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

func (m *MailManager) startServers(cfg *runtime.MailInfo) error {
	servers, ok := m.servers[cfg.Info.Name]
	if !ok {
		servers = map[string]MailServer{}
		m.servers[cfg.Info.Name] = servers
	}
	h := cfg.Handler(m.app.Monitor.Smtp, m.eventEmitter, m.app.Events)
	for _, server := range cfg.Servers {
		_, port, err := net.SplitHostPort(server.Host)
		if err != nil {
			return err
		}

		switch server.Protocol {
		case "smtp":
			if port == "" {
				port = "25"
			}
			if _, ok = servers[port]; ok {
				continue
			}

			log.Infof("adding new SMTP host on :%v", port)
			s := service.NewSmtpServer(port, h)
			s.Start()
			servers[port] = s
		case "smtps":
			if port == "" {
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
			if port == "" {
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

func (m *MailManager) Stop() {
	for _, ms := range m.servers {
		for _, server := range ms {
			server.Stop()
		}
	}
}

func (m *MailManager) removeService(name string) {
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

func (m *MailManager) cleanupRemovedServers(cfg *runtime.MailInfo, old map[string]*mail.Server) {
	for key, val := range old {
		if val2, ok := cfg.Servers[key]; ok {
			if val2.Host == val.Host && val2.Protocol == val.Protocol {
				continue
			}
		}

		_, port, err := net.SplitHostPort(val.Host)
		if err != nil {
			continue
		}
		if port == "" {
			switch val.Protocol {
			case "smtp":
				port = "25"
			case "smtps":
				port = "587"
			case "imap":
				port = "143"
			case "imaps":
				port = "993"
			}
		}

		addr := fmt.Sprintf(":%v", port)

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
