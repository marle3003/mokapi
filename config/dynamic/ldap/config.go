package ldap

import (
	"mokapi/config/dynamic/common"
	"net/url"
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

func (e Entry) Parse(config *common.Config, reader common.Reader) error {
	if aList, ok := e.Attributes["thumbnailphoto"]; ok {
		for i, a := range aList {
			u, err := url.Parse(a)
			if err != nil {
				continue
			}
			f, err := reader.Read(u, common.WithParent(config))
			if err != nil {
				return err
			}

			if b, ok := f.Data.([]byte); ok {
				aList[i] = string(b)
			}
		}
	}

	return nil
}
