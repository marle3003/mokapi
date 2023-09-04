package mail

import (
	"mokapi/config/dynamic/common"
	"mokapi/smtp"
	"regexp"
)

func init() {
	common.Register("smtp", &Config{})
}

type Config struct {
	ConfigPath    string   `yaml:"-" json:"-"`
	Info          Info     `yaml:"info" json:"info"`
	Server        string   `yaml:"server" json:"server"`
	Servers       []Server `yaml:"servers" json:"servers"`
	Imap          string   `yaml:"imap" json:"imap"`
	MaxRecipients int      `yaml:"maxRecipients,omitempty" json:"maxRecipients,omitempty"`
	//MaxMessageBytes   int       `yaml:"maxMessageBytes,omitempty" json:"maxMessageBytes,omitempty"`
	//AllowInsecureAuth bool      `yaml:"allowInsecureAuth,omitempty" json:"allowInsecureAuth,omitempty"`
	Mailboxes []MailboxConfig `yaml:"mailboxes" json:"mailboxes"`
	Rules     Rules           `yaml:"rules" json:"rules"`
}

type Info struct {
	Name        string `yaml:"title" json:"title"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string `yaml:"version" json:"version"`
}

type Server struct {
	Url         string `yaml:"url" json:"url"`
	Description string `yaml:"description" json:"description"`
}

type RuleAction string

type Rules []Rule

const (
	Allow RuleAction = "allow"
	Deny  RuleAction = "deny"
)

type Rule struct {
	Name           string          `yaml:"name" json:"name"`
	Sender         *RuleExpr       `yaml:"sender" json:"sender"`
	Recipient      *RuleExpr       `yaml:"recipient" json:"recipient"`
	Subject        *RuleExpr       `yaml:"subject" json:"subject"`
	Body           *RuleExpr       `yaml:"body" json:"body"`
	Action         RuleAction      `yaml:"action" json:"action"`
	RejectResponse *RejectResponse `yaml:"rejectResponse" json:"rejectResponse"`
}

type RejectResponse struct {
	StatusCode         smtp.StatusCode         `yaml:"statusCode" json:"statusCode"`
	EnhancedStatusCode smtp.EnhancedStatusCode `yaml:"enhancedStatusCode" json:"enhancedStatusCode"`
	Text               string                  `yaml:"text" json:"text"`
}

type RuleExpr struct {
	expr *regexp.Regexp
}

type MailboxConfig struct {
	Name     string `yaml:"name" json:"name"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

func (c *Config) getMailbox(name string) (MailboxConfig, bool) {
	for _, m := range c.Mailboxes {
		if m.Name == name {
			return m, true
		}
	}
	return MailboxConfig{}, false
}

func NewRuleExpr(r *regexp.Regexp) *RuleExpr {
	return &RuleExpr{expr: r}
}

func (r *RuleExpr) String() string {
	if r == nil {
		return ""
	}
	return r.expr.String()
}

func (r *RuleExpr) Match(v string) bool {
	return r.expr.Match([]byte(v))
}
