package ldap

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mokapi/config/dynamic"
	"path/filepath"
	"strings"
)

func init() {
	dynamic.Register("ldap", &Config{}, func(path string, c dynamic.Config, r dynamic.ConfigReader) (bool, dynamic.Config) {
		switch config := c.(type) {
		case *Config:
			config.ConfigPath = path
			dir := filepath.Dir(path)
			for _, e := range config.Entries {
				resolveThumbnail(e, dir)
			}
		}

		return true, c
	})
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

func resolveThumbnail(e Entry, dir string) {
	if aList, ok := e.Attributes["thumbnailphoto"]; ok {
		for i, a := range aList {
			if strings.HasPrefix(a, "file:") {
				file := strings.TrimPrefix(a, "file:")
				if strings.HasPrefix(file, "./") {
					file = strings.Replace(file, ".", dir, 1)
				}
				data, err := ioutil.ReadFile(file)
				if err != nil {
					log.Errorf("unable to read thumbnail %v: %v", file, err.Error())
				} else {
					aList[i] = string(data)
				}
			}
		}
	}
}
