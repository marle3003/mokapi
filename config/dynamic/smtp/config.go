package smtp

import "mokapi/config/dynamic/common"

func init() {
	common.Register("smtp", &Config{})
}

type Config struct {
	ConfigPath        string    `yaml:"-" json:"-"`
	Name              string    `yaml:"name" json:"name"`
	Description       string    `yaml:"description,omitempty" json:"description,omitempty"`
	Server            string    `yaml:"server" json:"server"`
	MaxRecipients     int       `yaml:"maxRecipients,omitempty" json:"maxRecipients,omitempty"`
	MaxMessageBytes   int       `yaml:"maxMessageBytes,omitempty" json:"maxMessageBytes,omitempty"`
	AllowInsecureAuth bool      `yaml:"allowInsecureAuth,omitempty" json:"allowInsecureAuth,omitempty"`
	Accounts          []Account `yaml:"accounts,omitempty" json:"accounts,omitempty"`
}

type Account struct {
	Username string
	Password string
}
