package mail

func (c *Config) Patch(patch *Config) {
	if patch == nil {
		return
	}
	c.patchInfo(patch)
	if len(c.Server) == 0 {
		c.Server = patch.Server
	}
	if patch.MaxRecipients > 0 {
		c.MaxRecipients = patch.MaxRecipients
	}
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
	c.AutoCreateMailbox = patch.AutoCreateMailbox
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
