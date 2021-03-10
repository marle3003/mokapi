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
	ConfigPath string `yaml:"-" json:"-"`
	Schedules  []Schedule
}

type ConfigRef struct {
	Ref   string
	Value Config
}

type Pipeline struct {
	Name   string
	Stages []*Stage
	Stage  *Stage
	Steps  string
}

type Stage struct {
	Name      string
	Steps     string
	Condition string
}

type Schedule struct {
	Name     string
	Every    string
	Pipeline string
}
