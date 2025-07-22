package mail

import (
	"mokapi/providers/mail"
	"net/url"
)

func (c *Config) Convert() *mail.Config {
	result := &mail.Config{Version: "1.0"}

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
		m := mail.MailboxConfig{
			Name:     mb.Name,
			Username: mb.Username,
			Password: mb.Password,
		}

		for _, fs := range mb.Folders {
			m.Folders = append(m.Folders, getFolder(fs))
		}

		result.Mailboxes = append(result.Mailboxes, m)
	}

	for _, r := range c.Rules {
		rule := mail.Rule{
			Name:      r.Name,
			Sender:    mail.NewRuleExpr(r.Sender.expr.String()),
			Recipient: mail.NewRuleExpr(r.Recipient.String()),
			Subject:   mail.NewRuleExpr(r.Subject.String()),
			Body:      mail.NewRuleExpr(r.Body.String()),
			Action:    mail.RuleAction(r.Action),
		}
		if r.RejectResponse != nil {
			rule.RejectResponse = &mail.RejectResponse{
				StatusCode:         r.RejectResponse.StatusCode,
				EnhancedStatusCode: r.RejectResponse.EnhancedStatusCode,
				Text:               r.RejectResponse.Text,
			}
		}
		result.Rules = append(result.Rules, rule)
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

func getFolder(f FolderConfig) mail.FolderConfig {
	r := mail.FolderConfig{
		Name:  f.Name,
		Flags: f.Flags,
	}

	for _, folder := range f.Folders {
		r.Folders = append(r.Folders, getFolder(folder))
	}
	return r
}
