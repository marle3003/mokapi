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
	Name        string
	Description string
	Stages      []*Stage
	Stage       *Stage
	Steps       string
	Variables   Variables
}

type Stage struct {
	Name        string
	Description string
	Steps       string
	Condition   string
}

type Schedule struct {
	Name  string
	Every string

	// Number of iterations, or less than 1 for unlimited
	Iterations int
	Pipeline   string
}

type Variables []Variable

type Variable struct {
	Name  string
	Value string
	Expr  string
}
