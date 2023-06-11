package mail

func (c *Config) Patch(patch *Config) {
	if patch == nil {
		return
	}
	c.patchInfo(patch)
	if len(c.Server) == 0 {
		c.Server = patch.Server
	}
	if c.MaxRecipients == 0 {
		c.MaxRecipients = patch.MaxRecipients
	}
	c.patchMailboxes(patch)
	c.Rules.patch(patch.Rules)
}

func (c *Config) patchInfo(patch *Config) {
	if len(c.Info.Description) == 0 {
		c.Info.Description = patch.Info.Description
	}
	if len(c.Info.Version) == 0 {
		c.Info.Version = patch.Info.Version
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
		for _, v := range c.Mailboxes {
			if v.Name == p.Name {
				v.patch(p)
				continue Loop
			}
		}
		c.Mailboxes = append(c.Mailboxes, p)
	}
}

func (m *MailboxConfig) patch(patch MailboxConfig) {
	if len(m.Username) == 0 {
		m.Username = patch.Username
	}
	if len(m.Password) == 0 {
		m.Password = patch.Password
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
	if r.Sender == nil {
		r.Sender = patch.Sender
	}
	if r.Recipient == nil {
		r.Recipient = patch.Recipient
	}
	if r.Subject == nil {
		r.Subject = patch.Subject
	}
	if r.Body == nil {
		r.Body = patch.Body
	}
	if r.RejectResponse == nil {
		r.RejectResponse = patch.RejectResponse
	}
}
