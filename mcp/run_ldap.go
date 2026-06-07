package mcp

import "mokapi/runtime"

type LdapAPI struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Type        string `json:"type"`

	info *runtime.LdapInfo
}

type LdapServer struct {
	Host        string `json:"host"`
	Description string `json:"description"`
}

type LdapEntry struct {
	Dn         string              `json:"dn"`
	Attributes map[string][]string `json:"attributes"`
}

func (m *mokapi) getLdapApi(name string) any {
	for _, api := range m.app.Ldap.List() {
		if api.Info.Name == name {
			return &LdapAPI{
				Name:        name,
				Description: api.Info.Description,
				Type:        "ldap",
				Address:     api.Address,
				info:        api,
			}
		}
	}
	return nil
}

func (l *LdapAPI) GetEntries() []LdapEntry {
	var entries []LdapEntry
	if l.info.Entries == nil {
		return entries
	}
	for it := l.info.Entries.Iter(); it.Next(); {
		entries = append(entries, LdapEntry{
			Dn:         it.Value().Dn,
			Attributes: it.Value().Attributes,
		})
	}
	return entries
}
