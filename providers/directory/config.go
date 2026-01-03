package directory

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/sortedmap"
	"mokapi/version"
	"net/url"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	ConfigPath string `yaml:"-" json:"-"`
	Info       Info
	Address    string
	SizeLimit  int64
	Entries    *sortedmap.LinkedHashMap[string, Entry]
	Files      []string

	root       map[string][]string
	Schema     *Schema
	oldEntries map[string]Entry
}

func (c *Config) Key() string {
	return c.ConfigPath
}

type Info struct {
	Name        string `yaml:"title"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

func (c *Config) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	c.Entries = &sortedmap.LinkedHashMap[string, Entry]{KeyNormalizer: normalizeDN}
	// copy from old format
	for k, v := range c.oldEntries {
		c.Entries.Set(k, v)
	}

	for it := c.Entries.Iter(); it.Next(); {
		entry := it.Value()
		err := entry.Parse(config, reader)
		if err != nil {
			return err
		}
	}

	for _, file := range c.Files {
		u, err := url.Parse(file)
		if err != nil {
			return fmt.Errorf("parse file %s failed: %v", file, err)
		}
		if !u.IsAbs() {
			if config.Info.Url != nil && config.Info.Url.Scheme == "file" {
				if !filepath.IsAbs(file) {
					path := config.Info.Url.Path
					if config.Info.Url.Opaque != "" {
						path = config.Info.Url.Opaque
					}
					dir := filepath.Dir(path)
					file = filepath.Join(dir, file)
				}
				file = filepath.Clean(file)
				u, err = url.Parse(fmt.Sprintf("file:%v", file))
			} else {
				u, err = config.Info.Url.Parse(file)
			}
			if err != nil {
				return fmt.Errorf("parse file %s failed: %v", file, err)
			}
		}

		d := &Ldif{}
		ref, err := reader.Read(u, d)
		if err != nil {
			return fmt.Errorf("read file %s failed: %v", file, err)
		}
		dynamic.AddRef(config, ref)
		err = d.Parse(ref, reader)
		if err != nil {
			return fmt.Errorf("parse file %s failed: %v", file, err)
		}
		for _, record := range d.Records {
			err = record.Apply(c.Entries, nil)
			if err != nil {
				return fmt.Errorf("apply file '%s' failed: %w", file, err)
			}
		}
	}

	root, ok := c.Entries.Get("")
	if !ok {
		root = Entry{
			Attributes: map[string][]string{
				"supportedLDAPVersion": {"3"},
				"vendorName":           {"Mokapi"},
				"vendorVersion":        {version.BuildVersion},
				"dsServiceName":        {c.Info.Name},
				"description":          {c.Info.Description},
				"namingContexts":       {"dc=mokapi,dc=io"},
				"subschemaSubentry":    {"cn=subschema"},
			},
		}
		schema := Entry{
			Dn: "cn=subschema",
			Attributes: map[string][]string{
				"cn":          {"subschema"},
				"objectClass": {"subschema"},
			},
		}
		c.Entries.Set("cn=subschema", schema)
	} else {
		if _, ok := root.Attributes["supportedLDAPVersion"]; !ok {
			root.Attributes["supportedLDAPVersion"] = []string{"3"}
		}

		if _, ok := root.Attributes["vendorName"]; !ok {
			root.Attributes["vendorName"] = []string{"Mokapi"}
		}

		if _, ok := root.Attributes["vendorVersion"]; !ok {
			root.Attributes["vendorVersion"] = []string{version.BuildVersion}
		}
		if _, ok := root.Attributes["dsServiceName"]; !ok && c.Info.Name != "" {
			root.Attributes["dsServiceName"] = []string{c.Info.Name}
		}
		if _, ok := root.Attributes["description"]; !ok && c.Info.Description != "" {
			root.Attributes["description"] = []string{c.Info.Description}
		}
		if _, ok := root.Attributes["namingContexts"]; !ok {
			root.Attributes["namingContexts"] = []string{"dc=mokapi,dc=io"}
		}
		if _, ok := root.Attributes["subschemaSubentry"]; !ok {
			root.Attributes["subschemaSubentry"] = []string{"cn=subschema"}
			schema := Entry{
				Dn: "cn=subschema",
				Attributes: map[string][]string{
					"cn":          {"subschema"},
					"objectClass": {"subschema"},
				},
			}
			c.Entries.Set("cn=subschema", schema)
		}
	}

	var v []string
	if v, ok = c.root["rootDomainNamingContext"]; ok {
		root.Attributes["rootDomainNamingContext"] = v
	}
	if v, ok = c.root["subschemaSubentry"]; ok {
		root.Attributes["subschemaSubentry"] = v
	}
	if v, ok = c.root["namingContexts"]; ok {
		root.Attributes["namingContexts"] = v
	}
	c.Entries.Set("", root)

	dn := root.Attributes["subschemaSubentry"]
	s, ok := c.Entries.Get(dn[0])
	if ok {
		var err error
		c.Schema, err = NewSchema(s)
		if err != nil {
			return err
		}
	}

	c.rebuildMemberOf()

	return nil
}

func (c *Config) getSizeLimit() int64 {
	return c.SizeLimit
}

func (c *Config) rebuildMemberOf() {
	// reset entries

	groups := map[string][]string{}
	for it := c.Entries.Iter(); it.Next(); {
		groupName := it.Value().Dn
		for k, v := range it.Value().Attributes {
			if k == "member" {
				if _, ok := groups[k]; !ok {
					groups[k] = []string{}
				}
				for _, member := range v {
					groups[groupName] = append(groups[groupName], member)
				}
			}
			if k == "memberOf" {
				delete(it.Value().Attributes, "memberOf")
			}
		}
	}

	for group, members := range groups {
		for _, member := range members {
			e, ok := c.Entries.Get(member)
			if !ok {
				log.Warnf("LDAP: member %s for group %s not found", member, group)
				continue
			}
			if e.Attributes == nil {
				e.Attributes = map[string][]string{}
			}
			e.Attributes["memberOf"] = append(e.Attributes["memberOf"], group)
			c.Entries.Set(member, e)
		}
	}
}

func (e *Entry) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if aList, ok := e.Attributes["thumbnailphoto"]; ok {
		for i, a := range aList {
			if !strings.HasPrefix(a, "file:") {
				continue
			}

			file := strings.TrimPrefix(a, "file:")
			if strings.HasPrefix(file, "./") {
				path := config.Info.Url.Opaque
				if len(path) == 0 {
					path = config.Info.Url.Path
				}
				dir := filepath.Dir(path)
				file = strings.Replace(file, ".", dir, 1)
			}
			u, err := url.Parse("file:" + file)
			if err != nil {
				continue
			}
			f, err := reader.Read(u, nil)
			if err != nil {
				return err
			}

			dynamic.AddRef(config, f)
			aList[i] = f.Data.(string)
		}
	}

	return nil
}

func normalizeDN(dn string) string {
	result := ""
	for _, rdn := range strings.Split(dn, ",") {
		if len(result) > 0 {
			result += ","
		}
		result += strings.TrimSpace(strings.ToLower(rdn))
	}
	return result
}
