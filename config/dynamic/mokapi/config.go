package mokapi

import (
	"mokapi/config/tls"
)

//func init() {
//	dynamic.Register("mokapi", &Config{}, func(config *dynamic.Config, r dynamic.ConfigReader) bool {
//		return true
//	})
//}

type Config struct {
	ConfigPath   string `yaml:"-" json:"-"`
	Certificates []Certificate
}

type Certificate struct {
	CertFile tls.FileOrContent `yaml:"certFile" json:"certFile"`
	KeyFile  tls.FileOrContent `yaml:"keyFile" json:"keyFile"`
}
