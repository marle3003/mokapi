package mail

func (c *Config) Patch(patch *Config) {
	if patch == nil {
		return
	}
	c.patchInfo(patch)
	c.patchSettings(patch.Settings)
	c.patchMailboxes(patch)
	c.patchRules(patch.Rules)
}

func (c *Config) patchInfo(patch *Config) {
	if len(patch.Info.Description) > 0 {
		c.Info.Description = patch.Info.Description
	}
}

func (c *Config) patchServer(patch *Config) {
	if len(c.Servers) == 0 {
		c.Servers = patch.Servers
	} else {
		for name, ps := range patch.Servers {
			if s, ok := c.Servers[name]; ok {
				s.patch(ps)
			} else {
				c.Servers[name] = ps
			}
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
		c.Settings = patch
	} else {
		c.Settings.AutoCreateMailbox = patch.AutoCreateMailbox
		c.Settings.MaxRecipients = patch.MaxRecipients
	}
}

func (c *Config) patchMailboxes(patch *Config) {
	if len(patch.Mailboxes) == 0 {
		return
	}
	if len(c.Mailboxes) == 0 {
		c.Mailboxes = patch.Mailboxes
		return
	}

	for name, p := range patch.Mailboxes {
		if mb, ok := c.Mailboxes[name]; ok {
			mb.patch(p)
		} else {
			c.Mailboxes[name] = p
		}
	}
}

func (m *MailboxConfig) patch(patch *MailboxConfig) {
	if len(patch.Username) > 0 {
		m.Username = patch.Username
	}
	if len(patch.Password) > 0 {
		m.Password = patch.Password
	}

	for name, child := range patch.Folders {
		if c, ok := m.Folders[name]; ok {
			c.patch(child)
		} else {
			if m.Folders == nil {
				m.Folders = make(map[string]*FolderConfig)
			}
			m.Folders[name] = child
		}
	}
}

func (c *Config) patchRules(patch map[string]*Rule) {
	if c.Rules == nil {
		c.Rules = patch
		return
	}

	for name, rule := range patch {
		if r, ok := c.Rules[name]; ok {
			r.patch(rule)
		} else {
			c.Rules[name] = rule
		}
	}
}

func (r *Rule) patch(patch *Rule) {
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
	f.Flags = patch.Flags

	for name, child := range patch.Folders {
		if c, ok := f.Folders[name]; ok {
			c.patch(child)
		} else {
			f.Folders[name] = child
		}
	}
}
