package server

import (
	"mokapi/config/dynamic/common"
	"mokapi/config/dynamic/directory"
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
	ldapConfig, ok := c.Data.(*directory.Config)
	if !ok {
		return
	}

	m.app.AddLdap(ldapConfig)
	d := directory.NewHandler(ldapConfig, m.eventEmitter)
	d = runtime.NewLdapHandler(m.app.Monitor.Ldap, d)

	if s, ok := m.servers[ldapConfig.Info.Name]; ok {
		s.Handler = d
	} else {
		s := &ldap.Server{Addr: ldapConfig.Address, Handler: d}
		m.servers[ldapConfig.Info.Name] = s
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
