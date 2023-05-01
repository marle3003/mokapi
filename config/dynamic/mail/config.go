package mail

import (
	"mokapi/config/dynamic/common"
	"regexp"
)

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
	Rules     []Rule    `yaml:"rules" json:"rules"`
}

type Info struct {
	Name        string `yaml:"title" json:"title"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string `yaml:"version" json:"version"`
}

type RuleAction string

const (
	Allow RuleAction = "allow"
	Deny  RuleAction = "deny"
)

type Rule struct {
	Sender    *RuleExpr  `yaml:"sender" json:"sender"`
	Recipient *RuleExpr  `yaml:"recipient" json:"recipient"`
	Subject   *RuleExpr  `yaml:"subject" json:"subject"`
	Body      *RuleExpr  `yaml:"body" json:"body"`
	Action    RuleAction `yaml:"action" json:"action"`
}

type RuleExpr struct {
	expr *regexp.Regexp
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

func NewRuleExpr(r *regexp.Regexp) *RuleExpr {
	return &RuleExpr{expr: r}
}

func (r *RuleExpr) String() string {
	return r.expr.String()
}

func (r *RuleExpr) Match(v string) bool {
	return r.expr.Match([]byte(v))
}
