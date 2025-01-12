package directory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

type config struct {
	Info    Info
	Server  server
	Entries []map[string]interface{}
	Files   []string
}

type server struct {
	Address                 string
	RootDomainNamingContext string   `yaml:"rootDomainNamingContext" json:"rootDomainNamingContext"`
	SubSchemaSubentry       string   `yaml:"subSchemaSubentry" json:"subSchemaSubentry"`
	NamingContexts          []string `yaml:"namingContexts" json:"namingContexts"`
}

type entry map[string]interface{}

func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	tmp := &config{}
	err := value.Decode(tmp)
	if err != nil {
		return err
	}

	c.init(tmp)

	return nil
}

func (c *Config) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	tmp := &config{}
	err := dec.Decode(&tmp)
	if err != nil {
		return err
	}

	c.init(tmp)

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

func (c *Config) init(raw *config) {
	c.Info = raw.Info
	c.Address = raw.Server.Address

	c.root = map[string][]string{}
	if raw.Server.RootDomainNamingContext != "" {
		c.root["rootDomainNamingContext"] = []string{raw.Server.RootDomainNamingContext}
	}
	if raw.Server.SubSchemaSubentry != "" {
		c.root["subSchemaSubentry"] = []string{raw.Server.SubSchemaSubentry}
	}
	if raw.Server.NamingContexts != nil {
		c.root["namingContexts"] = raw.Server.NamingContexts
	}

	c.Files = raw.Files

	if raw.Entries != nil {
		c.oldEntries = map[string]Entry{}
		for _, e := range raw.Entries {
			entry := buildEntry(e)
			c.oldEntries[entry.Dn] = entry
		}
	}
}
