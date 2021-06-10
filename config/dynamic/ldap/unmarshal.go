package ldap

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

type ldap struct {
	Info    Info
	Server  server
	Entries []map[string]interface{}
}

type server struct {
	Address                 string
	RootDomainNamingContext string   `yaml:"rootDomainNamingContext" json:"rootDomainNamingContext"`
	SubSchemaSubentry       string   `yaml:"subSchemaSubentry" json:"subSchemaSubentry"`
	NamingContexts          []string `yaml:"namingContexts" json:"namingContexts"`
}

type entry map[string]interface{}

func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	tmp := &ldap{}
	err := value.Decode(tmp)
	if err != nil {
		return err
	}

	c.Info = tmp.Info
	c.Address = tmp.Server.Address
	c.Root = Entry{
		Attributes: map[string][]string{
			"rootDomainNamingContext": {tmp.Server.RootDomainNamingContext},
			"subSchemaSubentry":       {tmp.Server.SubSchemaSubentry},
			"namingContexts":          tmp.Server.NamingContexts,
		},
	}
	c.Entries = make(map[string]Entry)

	for _, e := range tmp.Entries {
		entry := buildEntry(e)
		c.Entries[entry.Dn] = entry
	}

	return nil
}

func buildEntry(e entry) (r Entry) {
	r = Entry{Attributes: make(map[string][]string)}
	for k, v := range e {
		if s, ok := v.(string); ok {
			if strings.ToLower(k) == "dn" {
				r.Dn = s
				parts := strings.Split(s, ",")
				if len(parts) > 0 {
					kv := strings.Split(parts[0], "=")
					if len(kv) == 2 && strings.ToLower(kv[0]) == "cn" {
						r.Attributes["cn"] = []string{strings.TrimSpace(kv[1])}
					}
				} else {
					r.Attributes[k] = []string{s}
				}
			} else {
				r.Attributes[k] = []string{s}
			}
		} else if a, ok := v.([]interface{}); ok {
			list := make([]string, 0)
			for _, el := range a {
				if s, ok := el.(string); ok {
					list = append(list, s)
				}
			}
			r.Attributes[k] = list
		} else if ext, ok := v.(map[string]interface{}); ok {
			list := make([]string, 0, len(ext))
			for p, o := range ext {
				list = append(list, fmt.Sprintf("%v:%v", p, o))
			}
			r.Attributes[k] = list
		}
	}
	return
}
