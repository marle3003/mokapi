package models

import (
	"io/ioutil"
	"mokapi/config/dynamic"
	"path/filepath"
	"strings"
)

func (a *Application) ApplyLdap(config map[string]*dynamic.Ldap) {
	for key, item := range config {
		if len(item.Info.Name) > 0 {
			key = item.Info.Name
		}
		ldapServiceInfo, found := a.LdapServices[key]
		if !found {
			ldapServiceInfo = NewLdapServiceInfo()
		}
		ldapServiceInfo.Apply(item, key)
	}
}

func (l *LdapServiceInfo) Apply(config *dynamic.Ldap, file string) {
	l.Data.Name = config.Info.Name

	for k, v := range config.Server {
		if strings.ToLower(k) == "address" {
			if s, ok := v.(string); ok {
				l.Data.Address = s
			}
		} else {
			if s, ok := v.(string); ok {
				l.Data.Root.Attributes[k] = []string{s}
			} else if a, ok := v.([]interface{}); ok {
				list := make([]string, 0)
				for _, e := range a {
					if s, ok := e.(string); ok {
						list = append(list, s)
					}
				}
				l.Data.Root.Attributes[k] = list
			}
		}
	}

	for _, e := range config.Entries {
		entry := &Entry{Attributes: make(map[string][]string)}
		l.Data.Entries = append(l.Data.Entries, entry)
		for k, v := range e {
			if s, ok := v.(string); ok {
				if strings.ToLower(k) == "dn" {
					entry.Dn = s
					parts := strings.Split(s, ",")
					if len(parts) > 0 {
						kv := strings.Split(parts[0], "=")
						if len(kv) == 2 && strings.ToLower(kv[0]) == "cn" {
							entry.Attributes["cn"] = []string{strings.TrimSpace(kv[1])}
						}
					} else {
						entry.Attributes[k] = []string{s}
					}
				} else {
					entry.Attributes[k] = []string{s}
				}
			} else if a, ok := v.([]interface{}); ok {
				list := make([]string, 0)
				for _, el := range a {
					if s, ok := el.(string); ok {
						list = append(list, s)
					}
				}
				entry.Attributes[k] = list
			} else if ext, ok := v.(map[interface{}]interface{}); ok {
				for p, o := range ext {
					name := p.(string)
					switch strings.ToLower(name) {
					case "file":
						path := o.(string)
						if strings.HasPrefix(path, "./") {
							path = strings.Replace(path, ".", filepath.Dir(file), 1)
						}
						data, error := ioutil.ReadFile(path)
						if error != nil {
							l.Errors = append(l.Errors, error.Error())
						} else {
							entry.Attributes[k] = []string{string(data)}
						}
					}
				}
			}
		}
	}
}
