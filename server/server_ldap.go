package server

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic"
	engine "mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/runtime"
	"mokapi/server/cert"
	"sync"
)

type LdapDirectoryManager struct {
	servers map[string]*ldap.Server

	eventEmitter engine.EventEmitter
	certStore    *cert.Store
	app          *runtime.App
	m            sync.Mutex
}

func NewLdapDirectoryManager(emitter engine.EventEmitter, store *cert.Store, app *runtime.App) *LdapDirectoryManager {
	return &LdapDirectoryManager{
		eventEmitter: emitter,
		certStore:    store,
		app:          app,
		servers:      map[string]*ldap.Server{},
	}
}

func (m *LdapDirectoryManager) UpdateConfig(e dynamic.ConfigEvent) {
	if !runtime.IsLdapConfig(e.Config) {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	name, cfg := m.app.GetLdap(e.Config)
	if e.Event == dynamic.Delete {
		m.app.RemoveLdap(e.Config)
		if cfg.Config == nil {
			log.Infof("removing LDAP host '%v' on binding %v", name, m.servers[name].Addr)
			m.servers[name].Close()
			return
		}
	} else if cfg == nil {
		cfg = m.app.AddLdap(e.Config, m.eventEmitter)
	} else {
		addr := cfg.Address
		cfg.AddConfig(e.Config)
		if addr != cfg.Address {
			s := m.servers[name]
			log.Infof("removing LDAP host '%v' on binding %v", name, s.Addr)
			m.servers[name].Close()
			m.start(cfg)
		}
	}

	if s, ok := m.servers[cfg.Info.Name]; ok {
		s.Handler = cfg.Handler(m.app.Monitor.Ldap)
	} else {
		m.start(cfg)
	}
}

func (m *LdapDirectoryManager) start(cfg *runtime.LdapInfo) {
	s := &ldap.Server{Addr: cfg.Address, Handler: cfg.Handler(m.app.Monitor.Ldap)}
	m.servers[cfg.Info.Name] = s
	log.Infof("adding LDAP host '%v' on binding %v", cfg.Info.Name, s.Addr)
	go func() {
		err := s.ListenAndServe()
		if !errors.Is(err, ldap.ErrServerClosed) {
			log.Errorf("adding LDAP host '%v' on binding %v failed: %v", cfg.Info.Name, s.Addr, err)
		}
	}()
}

func (m *LdapDirectoryManager) Stop() {
	for _, s := range m.servers {
		s.Close()
	}
}
