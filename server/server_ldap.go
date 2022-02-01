package server

import (
	"mokapi/config/dynamic/common"
	config "mokapi/config/dynamic/ldap"
	"mokapi/server/ldap"
)

type LdapDirectories map[string]*ldap.Directory

func (dirs LdapDirectories) UpdateConfig(file *common.File) {
	config, ok := file.Data.(*config.Config)
	if !ok {
		return
	}

	if d, ok := dirs[config.Info.Name]; !ok {
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
