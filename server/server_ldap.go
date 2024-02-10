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

func (m LdapDirectoryManager) UpdateConfig(c *dynamic.Config) {
	if !runtime.IsLdapConfig(c) {
		return
	}

	m.m.Lock()
	defer m.m.Unlock()

	li := m.app.AddLdap(c, m.eventEmitter)

	if s, ok := m.servers[li.Info.Name]; ok {
		s.Handler = li.Handler(m.app.Monitor.Ldap)
	} else {
		s := &ldap.Server{Addr: li.Config.Address, Handler: li.Handler(m.app.Monitor.Ldap)}
		m.servers[li.Info.Name] = s
		go func() {
			err := s.ListenAndServe()
			if !errors.Is(err, ldap.ErrServerClosed) {
				log.Errorf("unable to start ldap server %v: %v", s.Addr, err)
			}
		}()
	}
}

func (m LdapDirectoryManager) Stop() {
	for _, s := range m.servers {
		s.Close()
	}
}
