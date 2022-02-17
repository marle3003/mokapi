package server

import (
	"mokapi/config/dynamic/common"
	config "mokapi/config/dynamic/ldap"
	"mokapi/engine"
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

func (m LdapDirectoryManager) UpdateConfig(file *common.File) {
	config, ok := file.Data.(*config.Config)
	if !ok {
		return
	}

	if d, ok := m.Directories[config.Info.Name]; !ok {
		d = ldap.NewDirectory(config)
		d.Start()
	} else {
		d.Update(config)
	}
}

func (dirs LdapDirectories) Stop() {
	for _, d := range dirs {
		d.Close()
	}
}
