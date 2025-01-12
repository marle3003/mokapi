package directory

import (
	"fmt"
	"mokapi/config/dynamic"
	"mokapi/version"
	"net/url"
	"path/filepath"
	"strings"
)

const (
	defaultSizeLimit = 1000
	defaultTimeLimit = 3600
)

type Config struct {
	ConfigPath string `yaml:"-" json:"-"`
	Info       Info
	Address    string
	SizeLimit  int64
	Entries    map[string]Entry
	Files      []string

	root       map[string][]string
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

type Entry struct {
	Dn         string
	Attributes map[string][]string
}

func (c *Config) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	c.Entries = map[string]Entry{}
	// copy from old format
	for k, v := range c.oldEntries {
		c.Entries[k] = v
	}

	for _, entry := range c.Entries {
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
					dir := filepath.Dir(config.Info.Url.Path)
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
			err = record.Apply(c.Entries)
			if err != nil {
				return fmt.Errorf("apply file '%s' failed: %w", file, err)
			}
		}
	}

	root, ok := c.Entries[""]
	if !ok {
		root = Entry{
			Attributes: map[string][]string{
				"supportedLDAPVersion": {"3"},
				"vendorName":           {"Mokapi"},
				"vendorVersion":        {version.BuildVersion},
				/*"control": {
					"1.2.840.113556.1.4.319 true", // Paged Results Control
				},*/
			},
		}
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
	}

	var v []string
	if v, ok = c.root["RootDomainNamingContext"]; ok {
		root.Attributes["rootDomainNamingContext"] = v
	}
	if v, ok = c.root["SubSchemaSubentry"]; ok {
		root.Attributes["subSchemaSubentry"] = v
	}
	if v, ok = c.root["NamingContexts"]; ok {
		root.Attributes["namingContexts"] = v
	}
	c.Entries[""] = root

	return nil
}

func (c *Config) getSizeLimit() int64 {
	return c.SizeLimit
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

func formatDN(dn string) string {
	result := ""
	for _, rdn := range strings.Split(dn, ",") {
		if len(result) > 0 {
			result += ","
		}
		result += strings.TrimSpace(rdn)
	}
	return result
}
