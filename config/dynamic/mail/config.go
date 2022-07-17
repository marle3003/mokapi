package mail

import "mokapi/config/dynamic/common"

func init() {
	common.Register("smtp", &Config{})
}

type Config struct {
	ConfigPath        string    `yaml:"-" json:"-"`
	Info              Info      `yaml:"info" json:"info"`
	Server            string    `yaml:"server" json:"server"`
	MaxRecipients     int       `yaml:"maxRecipients,omitempty" json:"maxRecipients,omitempty"`
	MaxMessageBytes   int       `yaml:"maxMessageBytes,omitempty" json:"maxMessageBytes,omitempty"`
	AllowInsecureAuth bool      `yaml:"allowInsecureAuth,omitempty" json:"allowInsecureAuth,omitempty"`
	Accounts          []Account `yaml:"accounts,omitempty" json:"accounts,omitempty"`
}

type Info struct {
	Name        string `yaml:"title" json:"title"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string `yaml:"version" json:"version"`
}

type Account struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}
