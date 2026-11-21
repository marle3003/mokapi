package mail

import "slices"

func (c *Config) Patch(patch *Config) {
	if patch == nil {
		return
	}
	c.patchInfo(patch)
	c.patchServer(patch)
	c.patchSettings(patch.Settings)
	c.patchMailboxes(patch)
	c.patchRules(patch.Rules)
}

func (c *Config) patchInfo(patch *Config) {
	if len(patch.Info.Description) > 0 {
		c.Info.Description = patch.Info.Description
	}
	if len(patch.Info.Version) > 0 {
		c.Info.Version = patch.Info.Version
	}
	if patch.Info.Contact != nil {
		c.Info.Contact = copyOf(patch.Info.Contact)
	}
}

func (c *Config) patchServer(patch *Config) {
	if len(patch.Servers) == 0 {
		return
	}
	if len(c.Servers) == 0 {
		c.Servers = cloneMap(patch.Servers, copyOf)
		return
	}
	for name, ps := range patch.Servers {
		if s, ok := c.Servers[name]; ok {
			s.patch(ps)
		} else {
			c.Servers[name] = copyOf(ps)
		}
	}
}

func (s *Server) patch(patch *Server) {
	if patch == nil {
		return
	}
	if patch.Host != "" {
		s.Host = patch.Host
	}
	if patch.Protocol != "" {
		s.Protocol = patch.Protocol
	}
	if patch.Description != "" {
		s.Description = patch.Description
	}
}

func (c *Config) patchSettings(patch *Settings) {
	if patch == nil {
		return
	}
	if c.Settings == nil {
		c.Settings = copyOf(patch)
	} else {
		c.Settings.AutoCreateMailbox = patch.AutoCreateMailbox
		c.Settings.MaxRecipients = patch.MaxRecipients
		c.Settings.MaxInboxMails = patch.MaxInboxMails
		c.Settings.AllowUnknownSenders = patch.AllowUnknownSenders
	}
}

func (c *Config) patchMailboxes(patch *Config) {
	if len(patch.Mailboxes) == 0 {
		return
	}
	if len(c.Mailboxes) == 0 {
		c.Mailboxes = cloneMap(patch.Mailboxes, (*MailboxConfig).clone)
		return
	}

	for name, p := range patch.Mailboxes {
		if mb, ok := c.Mailboxes[name]; ok {
			mb.patch(p)
		} else {
			c.Mailboxes[name] = p.clone()
		}
	}
}

func (m *MailboxConfig) patch(patch *MailboxConfig) {
	if patch == nil {
		return
	}
	if len(patch.Username) > 0 {
		m.Username = patch.Username
	}
	if len(patch.Password) > 0 {
		m.Password = patch.Password
	}
	if len(patch.Description) > 0 {
		m.Description = patch.Description
	}

	for name, child := range patch.Folders {
		if c, ok := m.Folders[name]; ok {
			c.patch(child)
		} else {
			if m.Folders == nil {
				m.Folders = make(map[string]*FolderConfig)
			}
			m.Folders[name] = child.clone()
		}
	}
}

func (c *Config) patchRules(patch map[string]*Rule) {
	if c.Rules == nil {
		c.Rules = cloneMap(patch, copyOf)
		return
	}

	for name, rule := range patch {
		if r, ok := c.Rules[name]; ok {
			r.patch(rule)
		} else {
			c.Rules[name] = copyOf(rule)
		}
	}
}

func (r *Rule) patch(patch *Rule) {
	if patch == nil {
		return
	}
	if patch.Sender != nil {
		r.Sender = patch.Sender
	}
	if patch.Recipient != nil {
		r.Recipient = patch.Recipient
	}
	if patch.Subject != nil {
		r.Subject = patch.Subject
	}
	if patch.Body != nil {
		r.Body = patch.Body
	}
	if patch.RejectResponse != nil {
		r.RejectResponse = patch.RejectResponse
	}
}

func (f *FolderConfig) patch(patch *FolderConfig) {
	if patch == nil {
		return
	}
	f.Flags = slices.Clone(patch.Flags)

	for name, child := range patch.Folders {
		if c, ok := f.Folders[name]; ok {
			c.patch(child)
		} else {
			if f.Folders == nil {
				f.Folders = make(map[string]*FolderConfig)
			}
			f.Folders[name] = child.clone()
		}
	}
}
