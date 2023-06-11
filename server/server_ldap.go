package server

import (
	"mokapi/config/dynamic/common"
	engine "mokapi/engine/common"
	"mokapi/ldap"
	"mokapi/runtime"
	"mokapi/server/cert"
)

type LdapDirectoryManager struct {
	servers map[string]*ldap.Server

	eventEmitter engine.EventEmitter
	certStore    *cert.Store
	app          *runtime.App
}

func NewLdapDirectoryManager(emitter engine.EventEmitter, store *cert.Store, app *runtime.App) *LdapDirectoryManager {
	return &LdapDirectoryManager{
		eventEmitter: emitter,
		certStore:    store,
		app:          app,
		servers:      map[string]*ldap.Server{},
	}
}

func (m LdapDirectoryManager) UpdateConfig(c *common.Config) {
	if !runtime.IsLdapConfig(c) {
		return
	}

	li := m.app.AddLdap(c, m.eventEmitter)

	if s, ok := m.servers[li.Info.Name]; ok {
		s.Handler = li.Handler(m.app.Monitor.Ldap)
	} else {
		s := &ldap.Server{Addr: li.Config.Address, Handler: li.Handler(m.app.Monitor.Ldap)}
		m.servers[li.Info.Name] = s
		go func() {
			s.ListenAndServe()
		}()
	}
}

func (m LdapDirectoryManager) Stop() {
	for _, s := range m.servers {
		s.Close()
	}
}
