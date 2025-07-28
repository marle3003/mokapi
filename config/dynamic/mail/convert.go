package mail

import (
	"fmt"
	"mokapi/providers/mail"
	"net/url"
)

func (c *Config) Convert() *mail.Config {
	result := &mail.Config{
		Version: "1.0",
		Info: mail.Info{
			Name:        c.Info.Name,
			Description: c.Info.Description,
			Version:     c.Info.Version,
		},
		Mailboxes: map[string]*mail.MailboxConfig{},
		Rules:     map[string]*mail.Rule{},
	}

	if c.Server != "" {
		name, server, err := getServerFromUrl(c.Server)
		if err == nil {
			result.Servers = map[string]*mail.Server{}
			result.Servers[name] = server
		}
	}

	for _, s := range c.Servers {
		name, server, err := getServerFromUrl(s.Url)
		if err == nil {
			if result.Servers == nil {
				result.Servers = map[string]*mail.Server{}
			}
			server.Description = s.Description
			result.Servers[name] = server
		}
	}

	for _, mb := range c.Mailboxes {
		m := &mail.MailboxConfig{
			Username: mb.Username,
			Password: mb.Password,
			Folders:  map[string]*mail.FolderConfig{},
		}

		for _, fs := range mb.Folders {
			m.Folders[fs.Name] = getFolder(fs)
		}

		result.Mailboxes[mb.Name] = m
	}

	for index, r := range c.Rules {
		rule := &mail.Rule{
			Name:   r.Name,
			Action: mail.RuleAction(r.Action),
		}
		if r.Sender != nil {
			rule.Sender = mail.NewRuleExpr(r.Sender.expr.String())
		}
		if r.Recipient != nil {
			rule.Recipient = mail.NewRuleExpr(r.Recipient.expr.String())
		}
		if r.Subject != nil {
			rule.Subject = mail.NewRuleExpr(r.Subject.String())
		}
		if r.Body != nil {
			rule.Body = mail.NewRuleExpr(r.Body.String())
		}

		if r.RejectResponse != nil {
			rule.RejectResponse = &mail.RejectResponse{
				StatusCode:         r.RejectResponse.StatusCode,
				EnhancedStatusCode: r.RejectResponse.EnhancedStatusCode,
				Message:            r.RejectResponse.Text,
			}
		}
		if rule.Name == "" {
			rule.Name = fmt.Sprintf("%d", index+1)
		}
		result.Rules[rule.Name] = rule
	}

	return result
}

func getServerFromUrl(serverUrl string) (string, *mail.Server, error) {
	u, err := url.Parse(serverUrl)
	if err == nil && u.Host != "" {
		return u.String(), &mail.Server{
			Host:     u.Host,
			Protocol: u.Scheme,
		}, nil
	}
	return "", nil, err
}

func getFolder(f FolderConfig) *mail.FolderConfig {
	r := &mail.FolderConfig{
		Flags:   f.Flags,
		Folders: map[string]*mail.FolderConfig{},
	}

	for _, folder := range f.Folders {
		r.Folders[folder.Name] = getFolder(folder)
	}
	return r
}
