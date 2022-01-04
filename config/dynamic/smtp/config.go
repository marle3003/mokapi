package smtp

import "mokapi/config/dynamic/common"

func init() {
	common.Register("smtp", &Config{})
}

type Config struct {
	ConfigPath        string `yaml:"-" json:"-"`
	Name              string
	Description       string
	Server            string
	MaxRecipients     int
	MaxMessageBytes   int
	AllowInsecureAuth bool
	Accounts          []Account
}

type Account struct {
	Username string
	Password string
}
