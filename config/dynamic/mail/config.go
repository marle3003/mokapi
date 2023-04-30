package mail

import "mokapi/config/dynamic/common"

func init() {
	common.Register("smtp", &Config{})
}

type Config struct {
	ConfigPath    string `yaml:"-" json:"-"`
	Info          Info   `yaml:"info" json:"info"`
	Server        string `yaml:"server" json:"server"`
	MaxRecipients int    `yaml:"maxRecipients,omitempty" json:"maxRecipients,omitempty"`
	//MaxMessageBytes   int       `yaml:"maxMessageBytes,omitempty" json:"maxMessageBytes,omitempty"`
	//AllowInsecureAuth bool      `yaml:"allowInsecureAuth,omitempty" json:"allowInsecureAuth,omitempty"`
	Mailboxes []Mailbox `yaml:"mailboxes" json:"mailboxes"`
	AllowList []Rule    `yaml:"allowList" json:"allowList"`
	DenyList  []Rule    `yaml:"denyList" json:"denyList"`
}

type Info struct {
	Name        string `yaml:"title" json:"title"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string `yaml:"version" json:"version"`
}

type Rule struct {
	Sender    string `yaml:"sender" json:"sender"`
	Recipient string `yaml:"recipient" json:"recipient"`
	Subject   string `yaml:"subject" json:"subject"`
	Body      string `yaml:"body" json:"body"`
}

type Mailbox struct {
	Name     string `yaml:"name" json:"name"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

func (c *Config) getMailbox(name string) (Mailbox, bool) {
	for _, m := range c.Mailboxes {
		if m.Name == name {
			return m, true
		}
	}
	return Mailbox{}, false
}
