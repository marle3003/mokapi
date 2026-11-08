package server

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	engine "mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/runtime"
	"mokapi/server/cert"
	"net"
	"sync"
)

type LdapDirectoryManager struct {
	servers map[string]*ldapHost

	eventEmitter engine.EventEmitter
	certStore    *cert.Store
	app          *runtime.App
	m            sync.Mutex
}

// ldapHost keeps the listener next to the server because ldap.Server only learns about it once
// Serve runs, so Close alone can leave the port open.
type ldapHost struct {
	*ldap.Server
	ln net.Listener
}

func (h *ldapHost) close() {
	h.Close()
	if h.ln != nil {
		_ = h.ln.Close()
	}
}

func NewLdapDirectoryManager(emitter engine.EventEmitter, store *cert.Store, app *runtime.App) *LdapDirectoryManager {
	return &LdapDirectoryManager{
		eventEmitter: emitter,
		certStore:    store,
		app:          app,
		servers:      map[string]*ldapHost{},
	}
}

func (m *LdapDirectoryManager) UpdateConfig(e dynamic.ConfigEvent) {
	cfg, ok := runtime.IsLdapConfig(e.Config)
	if !ok {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	name := cfg.Info.Name
	info := m.app.Ldap.Get(cfg.Info.Name)
	if e.Event == dynamic.Delete {
		m.app.Ldap.Remove(e.Config)
		if info.Config == nil {
			log.Infof("removing LDAP host '%v' on binding %v", name, m.servers[name].Addr)
			m.servers[name].close()
			return
		}
	} else if info == nil {
		info = m.app.Ldap.Add(e.Config, m.eventEmitter)
	} else {
		addr := info.Address
		info.AddConfig(e.Config)
		if addr != cfg.Address {
			s := m.servers[name]
			log.Infof("removing LDAP host '%v' on binding %v", name, s.Addr)
			m.servers[name].close()
			m.start(info)
		}
	}

	if s, ok := m.servers[cfg.Info.Name]; ok {
		s.Handler = info.Handler(m.app.Monitor.Ldap, m.app.Events)
	} else {
		m.start(info)
	}
}

func (m *LdapDirectoryManager) start(cfg *runtime.LdapInfo) {
	s := &ldapHost{Server: &ldap.Server{Addr: cfg.Address, Handler: cfg.Handler(m.app.Monitor.Ldap, m.app.Events)}}
	m.servers[cfg.Info.Name] = s
	log.Infof("adding LDAP host '%v' on binding %v", cfg.Info.Name, s.Addr)

	logFailure := func(err error) {
		log.Errorf("adding LDAP host '%v' on binding %v failed: %v", cfg.Info.Name, s.Addr, err)
	}
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		logFailure(err)
		return
	}
	s.ln = ln
	go func() {
		if err := s.Serve(ln); !errors.Is(err, ldap.ErrServerClosed) {
			logFailure(err)
		}
	}()
}

func (m *LdapDirectoryManager) Stop() {
	for _, s := range m.servers {
		s.close()
	}
}
