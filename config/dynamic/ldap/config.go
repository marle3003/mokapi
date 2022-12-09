package ldap

import (
	"mokapi/config/dynamic/common"
	"net/url"
	"path/filepath"
	"strings"
)

func init() {
	common.Register("ldap", &Config{})
}

type Config struct {
	ConfigPath string `yaml:"-" json:"-"`
	Info       Info
	Address    string
	Root       Entry
	Entries    map[string]Entry
}

func (c *Config) Key() string {
	return c.ConfigPath
}

type Info struct {
	Name        string `yaml:"title"`
	Description string `yaml:"description"`
}

type Entry struct {
	Dn         string
	Attributes map[string][]string
}

func (c *Config) Parse(config *common.Config, reader common.Reader) error {
	for _, entry := range c.Entries {
		err := entry.Parse(config, reader)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e Entry) Parse(config *common.Config, reader common.Reader) error {
	if aList, ok := e.Attributes["thumbnailphoto"]; ok {
		for i, a := range aList {
			if !strings.HasPrefix(a, "file:") {
				continue
			}

			file := strings.TrimPrefix(a, "file:")
			if strings.HasPrefix(file, "./") {
				dir := filepath.Dir(config.Url.Opaque)
				if len(dir) == 0 {
					dir = filepath.Dir(config.Url.Path)
				}
				file = strings.Replace(file, ".", dir, 1)
			}
			u, err := url.Parse("file:" + file)
			if err != nil {
				continue
			}
			f, err := reader.Read(u, common.WithParent(config))
			if err != nil {
				return err
			}

			aList[i] = f.Data.(string)
		}
	}

	return nil
}
