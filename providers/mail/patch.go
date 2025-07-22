package mail

func (c *Config) Patch(patch *Config) {
	if patch == nil {
		return
	}
	c.patchInfo(patch)
	c.patchSettings(patch.Settings)
	c.patchMailboxes(patch)
	c.Rules.patch(patch.Rules)
}

func (c *Config) patchInfo(patch *Config) {
	if len(patch.Info.Description) > 0 {
		c.Info.Description = patch.Info.Description
	}
	if len(patch.Info.Version) > 0 {
		c.Info.Version = patch.Info.Version
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

Loop:
	for _, p := range patch.Mailboxes {
		for i, v := range c.Mailboxes {
			if v.Name == p.Name {
				v.patch(&p)
				c.Mailboxes[i] = v
				continue Loop
			}
		}
		c.Mailboxes = append(c.Mailboxes, p)
	}
}

func (m *MailboxConfig) patch(patch *MailboxConfig) {
	if len(patch.Username) > 0 {
		m.Username = patch.Username
	}
	if len(patch.Password) > 0 {
		m.Password = patch.Password
	}
	if len(m.Folders) == 0 {
		m.Folders = patch.Folders
	} else {
		for _, folder := range patch.Folders {
			for i, f := range m.Folders {
				if f.Name == folder.Name {
				Flags:
					for _, flag := range folder.Flags {
						for _, orig := range f.Flags {
							if orig == flag {
								continue Flags
							}
						}
						f.Flags = append(f.Flags, flag)
					}
				}
				m.Folders[i] = f
			}
		}
	}
}

func (r *Rules) patch(patch Rules) {
Loop:
	for _, p := range patch {
		for i := range *r {
			v := &(*r)[i]
			if v.Name == p.Name {
				v.patch(p)
				continue Loop
			}
		}
		*r = append(*r, p)
	}
}

func (r *Rule) patch(patch Rule) {
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
