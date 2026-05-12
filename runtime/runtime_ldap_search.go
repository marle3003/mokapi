package runtime

import (
	"fmt"
	"mokapi/providers/directory"
	"mokapi/runtime/search"
	"strings"
)

type ldapSearchIndexData struct {
	Type          string            `json:"type"`
	Discriminator string            `json:"discriminator"`
	Api           string            `json:"api"`
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Server        string            `json:"server"`
	Description   string            `json:"description"`
	Meta          map[string]string `json:"meta"`
}

type ldapSearchIndexEntry struct {
	Type          string `json:"type"`
	Discriminator string `json:"discriminator"`
	Api           string `json:"api"`
	Dn            string `json:"dn"`
	// Bleve does not index map keys
	Attributes []ldapSearchIndexAttribute `json:"attributes"`
}

type ldapSearchIndexAttribute struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

func (s *LdapStore) addToIndex(cfg *directory.Config) {
	if cfg == nil || cfg.Info.Name == "" {
		return
	}

	c := ldapSearchIndexData{
		Type:          "ldap",
		Discriminator: "ldap",
		Api:           cfg.Info.Name,
		Name:          cfg.Info.Name,
		Version:       cfg.Info.Version,
		Description:   cfg.Info.Description,
		Server:        cfg.Address,
		Meta: map[string]string{
			"entries": fmt.Sprintf("%d", cfg.Entries.Len()),
		},
	}

	s.index.Add(fmt.Sprintf("ldap_%s", cfg.Info.Name), c)

	if cfg.Entries != nil {
		for it := cfg.Entries.Iter(); it.Next(); {
			e := it.Value()
			se := ldapSearchIndexEntry{
				Type:          "ldap",
				Discriminator: "ldap_entry",
				Api:           cfg.Info.Name,
				Dn:            e.Dn,
			}
			for name, values := range e.Attributes {
				se.Attributes = append(se.Attributes, ldapSearchIndexAttribute{
					Name:   name,
					Values: values,
				})
			}
			s.index.Add(fmt.Sprintf("ldap_%s_%s", cfg.Info.Name, e.Dn), se)
		}
	}
}

func getLdapSearchResult(fields map[string]string, discriminator []string) (search.ResultItem, error) {
	result := search.ResultItem{
		Type: "LDAP",
	}

	if len(discriminator) == 1 {
		result.Title = fields["name"]
		result.Params = map[string]string{
			"type":    strings.ToLower(result.Type),
			"service": result.Title,
		}
		for k, v := range fields {
			if strings.HasPrefix(k, "meta.") {
				k = strings.Replace(k, "meta.", "", 1)
				result.Params[k] = v
			}
		}
		return result, nil
	}

	switch discriminator[1] {
	case "entry":
		result.Domain = fields["api"]
		result.Title = fields["dn"]
		result.Params = map[string]string{
			"type":    strings.ToLower(result.Type),
			"service": result.Domain,
			"entry":   fields["dn"],
		}
	default:
		return result, fmt.Errorf("unsupported search result: %s", strings.Join(discriminator, "_"))
	}
	return result, nil
}

func (s *LdapStore) removeFromIndex(cfg *directory.Config) {
	s.index.Delete(fmt.Sprintf("ldap_%s", cfg.Info.Name))

	if cfg.Entries != nil {
		for it := cfg.Entries.Iter(); it.Next(); {
			e := it.Value()
			s.index.Delete(fmt.Sprintf("mail_%s_%s", cfg.Info.Name, e.Dn))
		}
	}
}
