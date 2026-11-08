package server

import (
	"mokapi/config/dynamic"
	engine "mokapi/engine/common"
	"mokapi/imap"
	"mokapi/providers/mail"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/server/service"
	"mokapi/smtp"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
)

type MailServer interface {
	Stop()
	Start() error
}

type MailManager struct {
	servers      map[string]map[string]MailServer
	app          *runtime.App
	eventEmitter engine.MailEventEmitter
	certStore    *cert.Store
	m            sync.Mutex
}

func NewMailManager(app *runtime.App, eventEmitter engine.MailEventEmitter, store *cert.Store) *MailManager {
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

type mailProtocol struct {
	label       string
	defaultPort string
	newServer   func(port string, h runtime.MailHandler, store *cert.Store) MailServer
}

var mailProtocols = map[string]mailProtocol{
	"smtp": {"SMTP", "25", func(port string, h runtime.MailHandler, store *cert.Store) MailServer {
		return service.NewSmtpServer(port, h, store, smtp.StartTls)
	}},
	"smtps": {"SMTPS", "587", func(port string, h runtime.MailHandler, store *cert.Store) MailServer {
		return service.NewSmtpServer(port, h, store, smtp.Implicit)
	}},
	"imap": {"IMAP", "143", func(port string, h runtime.MailHandler, store *cert.Store) MailServer {
		return service.NewImapServer(port, h, store, imap.StartTls)
	}},
	"imaps": {"IMAPS", "993", func(port string, h runtime.MailHandler, store *cert.Store) MailServer {
		return service.NewImapServer(port, h, store, imap.Implicit)
	}},
}

func (m *MailManager) startServers(cfg *runtime.MailInfo) error {
	servers, ok := m.servers[cfg.Info.Name]
	if !ok {
		servers = map[string]MailServer{}
		m.servers[cfg.Info.Name] = servers
	}
	h := cfg.Handler(m.app.Monitor.Mail, m.eventEmitter, m.app.Events)
	for _, server := range cfg.Servers {
		protocol, ok := mailProtocols[server.Protocol]
		if !ok {
			continue
		}
		port := mailPort(server)
		if _, ok = servers[port]; ok {
			continue
		}

		log.Infof("adding new %v host on :%v", protocol.label, port)
		s := protocol.newServer(port, h, m.certStore)
		startAndLogError(s)
		servers[port] = s
	}
	return nil
}

func mailPort(server *mail.Server) string {
	_, port, _ := net.SplitHostPort(server.Host)
	if port == "" {
		return mailProtocols[server.Protocol].defaultPort
	}
	return port
}

func (m *MailManager) Stop() {
	m.m.Lock()
	defer m.m.Unlock()

	for _, ms := range m.servers {
		for _, server := range ms {
			server.Stop()
		}
	}
}

func (m *MailManager) removeService(name string) {
	for port, server := range m.servers[name] {
		log.Infof("removing service '%v' on binding :%v", name, port)
		server.Stop()
	}
	delete(m.servers, name)
}

func (m *MailManager) cleanupRemovedServers(cfg *runtime.MailInfo, old map[string]*mail.Server) {
	servers := m.servers[cfg.Info.Name]
	for key, val := range old {
		if val2, ok := cfg.Servers[key]; ok {
			if val2.Host == val.Host && val2.Protocol == val.Protocol {
				continue
			}
		}

		port := mailPort(val)
		if server, ok := servers[port]; ok {
			log.Infof("removing '%v' on binding :%v", cfg.Info.Name, port)
			server.Stop()
			delete(servers, port)
		}
	}
}
