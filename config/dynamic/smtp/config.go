package smtp

import "mokapi/config/dynamic"

func init() {
	dynamic.Register("smtp", &Config{}, func(path string, config dynamic.Config, r dynamic.ConfigReader) (bool, dynamic.Config) {
		switch c := config.(type) {
		case *Config:
			c.ConfigPath = path
			if len(c.Name) == 0 {
				c.Name = c.Address
			}
		}
		return true, config
	})
}

type Config struct {
	ConfigPath string `yaml:"-" json:"-"`
	Name       string
	Address    string
	Tls        *Tls
}

type Tls struct {
}
