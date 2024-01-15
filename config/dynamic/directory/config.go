package directory

import (
	"mokapi/config/dynamic"
	"net/url"
	"path/filepath"
	"strings"
)

func init() {
	dynamic.Register("ldap", &Config{})
}

const (
	defaultSizeLimit = 1000
	defaultTimeLimit = 3600
)

type Config struct {
	ConfigPath string `yaml:"-" json:"-"`
	Info       Info
	Address    string
	Root       Entry
	SizeLimit  int64
	Entries    map[string]Entry
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
	for _, entry := range c.Entries {
		err := entry.Parse(config, reader)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) getSizeLimit() int64 {
	if c.SizeLimit == 0 {
		return defaultSizeLimit
	}
	return c.SizeLimit
}

func (e Entry) Parse(config *dynamic.Config, reader dynamic.Reader) error {
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
