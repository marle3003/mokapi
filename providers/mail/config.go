package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
	"mokapi/smtp"
	"regexp"
)

var serverNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type Config struct {
	Version   string                    `yaml:"mail" json:"mail"`
	Info      Info                      `yaml:"info" json:"info"`
	Servers   map[string]*Server        `yaml:"servers" json:"servers"`
	Settings  *Settings                 `yaml:"settings" json:"settings"`
	Mailboxes map[string]*MailboxConfig `yaml:"mailboxes" json:"mailboxes"`
	Rules     Rules                     `yaml:"rules" json:"rules"`
}

type Info struct {
	Name        string `yaml:"title" json:"title"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

type Server struct {
	Host        string `yaml:"host" json:"host"`
	Protocol    string `yaml:"protocol" json:"protocol"`
	Description string `yaml:"description" json:"description"`
}

type Settings struct {
	MaxRecipients     int  `yaml:"maxRecipients" json:"maxRecipients"`
	AutoCreateMailbox bool `yaml:"autoCreateMailbox" json:"autoCreateMailbox"`
}

type MailboxConfig struct {
	Username    string                   `yaml:"username" json:"username"`
	Password    string                   `yaml:"password" json:"password"`
	Description string                   `yaml:"description,omitempty" json:"description,omitempty"`
	Folders     map[string]*FolderConfig `yaml:"folders" json:"folders"`
}

type FolderConfig struct {
	Flags   []string                 `yaml:"flags" json:"flags"`
	Folders map[string]*FolderConfig `yaml:"folders" json:"folders"`
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

type RuleExpr struct {
	expr *regexp.Regexp
}

type RejectResponse struct {
	StatusCode         smtp.StatusCode         `yaml:"statusCode" json:"statusCode"`
	EnhancedStatusCode smtp.EnhancedStatusCode `yaml:"enhancedStatusCode" json:"enhancedStatusCode"`
	Text               string                  `yaml:"text" json:"text"`
}

func (c *Config) Parse(_ *dynamic.Config, _ dynamic.Reader) error {
	for name := range c.Servers {
		if !serverNamePattern.MatchString(name) {
			return fmt.Errorf("server name '%s' does not match valid pattern", name)
		}
	}

	for name := range c.Mailboxes {
		if name == "" {
			return fmt.Errorf("mailbox name is required")
		}
	}

	return nil
}

func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	type alias Config
	tmp := alias(*c)
	tmp.Settings = &Settings{AutoCreateMailbox: true}
	err := value.Decode(&tmp)
	*c = Config(tmp)
	return err
}

func (c *Config) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	type alias Config
	tmp := alias(*c)
	tmp.Settings = &Settings{AutoCreateMailbox: true}
	err := dec.Decode(&tmp)
	*c = Config(tmp)
	return err
}

func (r *RuleExpr) UnmarshalYAML(value *yaml.Node) error {
	var err error
	r.expr, err = regexp.Compile(value.Value)
	return err
}

func (r *RuleExpr) UnmarshalJSON(b []byte) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	t, err := dec.Token()
	if err != nil {
		return err
	}
	r.expr, err = regexp.Compile(t.(string))
	return err
}
