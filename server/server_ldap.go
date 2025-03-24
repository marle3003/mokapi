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
			m.servers[name].Close()
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
			m.servers[name].Close()
			m.start(info)
		}
	}

	if s, ok := m.servers[cfg.Info.Name]; ok {
		s.Handler = info.Handler(m.app.Monitor.Ldap)
	} else {
		m.start(info)
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
