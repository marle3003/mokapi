package server

import (
	log "github.com/sirupsen/logrus"
	"mokapi/config/dynamic/common"
	config "mokapi/config/dynamic/ldap"
	engine "mokapi/engine/common"
	"mokapi/runtime"
	"mokapi/server/cert"
	"mokapi/server/ldap"
)

type LdapDirectories map[string]*ldap.Directory

type LdapDirectoryManager struct {
	Directories LdapDirectories

	eventEmitter engine.EventEmitter
	certStore    *cert.Store
	app          *runtime.App
}

func NewLdapDirectoryManager(directories LdapDirectories, emitter engine.EventEmitter, store *cert.Store, app *runtime.App) *LdapDirectoryManager {
	return &LdapDirectoryManager{
		eventEmitter: emitter,
		certStore:    store,
		app:          app,
		Directories:  directories,
	}
}

func (m LdapDirectoryManager) UpdateConfig(c *common.Config) {
	ldapConfig, ok := c.Data.(*config.Config)
	if !ok {
		return
	}

	if d, ok := m.Directories[ldapConfig.Info.Name]; !ok {
		d = ldap.NewDirectory(ldapConfig, m.app.Monitor.Ldap)
		d.Start()
	} else {
		d.Update(ldapConfig)
	}
}

func (dirs LdapDirectories) Stop() {
	if len(dirs) > 0 {
		log.Debug("stopping ldap directories")
	}
	for _, d := range dirs {
		d.Close()
	}
}
