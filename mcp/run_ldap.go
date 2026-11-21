package mcp

import (
	"fmt"
	"mokapi/runtime"
)

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
		cfg := api.Config()
		if cfg != nil && cfg.Info.Name == name {
			return &LdapAPI{
				Name:        name,
				Description: cfg.Info.Description,
				Type:        "ldap",
				Address:     cfg.Address,
				info:        api,
			}
		}
	}
	return nil
}

// notFound is the error for an API whose last file was removed after the object was created.
func (l *LdapAPI) notFound() error {
	return fmt.Errorf("ldap API '%s' not found", l.Name)
}

func (l *LdapAPI) GetEntries() ([]LdapEntry, error) {
	cfg := l.info.Config()
	if cfg == nil {
		return nil, l.notFound()
	}

	entries := []LdapEntry{}
	for _, e := range cfg.Entries.Entries() {
		entries = append(entries, LdapEntry{
			Dn:         e.Dn,
			Attributes: e.Attributes,
		})
	}
	return entries, nil
}
