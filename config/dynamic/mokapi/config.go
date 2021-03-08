package mokapi

import (
	"mokapi/config/dynamic"
)

func init() {
	dynamic.Register("mokapi", &Config{}, func(path string, config dynamic.Config, r dynamic.ConfigReader) (bool, dynamic.Config) {
		switch c := config.(type) {
		case *Config:
			c.ConfigPath = path
		}
		return true, config
	})
}

type Config struct {
	Name       string
	Pipelines  []Pipeline
	Reference  string `yaml:"$ref" json:"$ref"`
	ConfigPath string `yaml:"-" json:"-"`
	Schedules  []*Schedule
}

type Pipeline struct {
	Name   string
	Stages []Stage
	Steps  string
}

type Stage struct {
	Steps string
}

type Schedule struct {
	Name     string
	Cron     string
	Pipeline string
}
