package smtp

import "mokapi/config/dynamic/common"

func init() {
	common.Register("smtp", &Config{})
}

type Config struct {
	ConfigPath  string `yaml:"-" json:"-"`
	Name        string
	Description string
	Address     string
	Tls         *Tls
}

type Tls struct {
}
