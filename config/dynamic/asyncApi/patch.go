package asyncApi

func (c *Config) Patch(patch *Config) {
	c.patchInfo(patch)
	c.patchServer(patch)
	c.patchChannels(patch)
	c.patchComponents(patch)
}

func (c *Config) patchInfo(patch *Config) {
	if len(patch.Info.Description) > 0 {
		c.Info.Description = patch.Info.Description
	}
	if len(patch.Info.Version) > 0 {
		c.Info.Version = patch.Info.Version
	}
	if c.Info.Contact == nil {
		c.Info.Contact = patch.Info.Contact
	} else {
		c.Info.Contact.patch(patch.Info.Contact)
	}
	if len(patch.Info.TermsOfService) > 0 {
		c.Info.TermsOfService = patch.Info.TermsOfService
	}
	if c.Info.License == nil {
		c.Info.License = patch.Info.License
	} else {
		c.Info.License.patch(patch.Info.License)
	}
}

func (c *Contact) patch(patch *Contact) {
	if len(patch.Name) > 0 {
		c.Name = patch.Name
	}
	if len(patch.Url) > 0 {
		c.Url = patch.Url
	}
	if len(patch.Email) > 0 {
		c.Email = patch.Email
	}
}

func (l *License) patch(patch *License) {
	if len(patch.Name) > 0 {
		l.Name = patch.Name
	}
	if len(patch.Url) > 0 {
		l.Url = patch.Url
	}
}

func (c *Config) patchServer(patch *Config) {
	if len(c.Servers) == 0 {
		c.Servers = patch.Servers
	} else {
		for name, ps := range patch.Servers {
			if s, ok := c.Servers[name]; ok {
				s.patch(ps)
				c.Servers[name] = s
			} else {
				c.Servers[name] = ps
			}
		}
	}
}

func (s *ServerRef) patch(patch *ServerRef) {
	if patch == nil {
		return
	}
	if s.Value == nil {
		s.Value = patch.Value
	} else {
		s.Value.patch(patch.Value)
	}
}

func (s *Server) patch(patch *Server) {
	if patch == nil {
		return
	}

	if len(patch.Url) > 0 {
		s.Url = patch.Url
	}

	if len(patch.Description) > 0 {
		s.Description = patch.Description
	}

	if len(patch.Protocol) > 0 {
		s.Protocol = patch.Protocol
	}
	if len(patch.ProtocolVersion) > 0 {
		s.ProtocolVersion = patch.ProtocolVersion
	}

	s.Bindings.Kafka.Patch(patch.Bindings.Kafka)
}

func (c *Config) patchChannels(patch *Config) {
	if patch.Channels == nil {
		return
	}
	if c.Channels == nil {
		c.Channels = map[string]*ChannelRef{}
	}

	for k, v := range patch.Channels {
		if ch, ok := c.Channels[k]; ok {
			ch.patch(v)
		} else {
			c.Channels[k] = v
		}
	}
}

func (c *ChannelRef) patch(patch *ChannelRef) {
	if patch == nil {
		return
	}
	if c.Value == nil {
		c.Value = patch.Value
	} else {
		c.Value.patch(patch.Value)
	}
}

func (c *Channel) patch(patch *Channel) {
	if patch == nil {
		return
	}
	if len(patch.Description) > 0 {
		c.Description = patch.Description
	}
	if c.Subscribe == nil {
		c.Subscribe = patch.Subscribe
	} else {
		c.Subscribe.patch(patch.Subscribe)
	}
	if c.Publish == nil {
		c.Publish = patch.Publish
	} else {
		c.Publish.patch(patch.Publish)
	}
	c.Bindings.Kafka.Patch(patch.Bindings.Kafka)
}

func (o *Operation) patch(patch *Operation) {
	if patch == nil {
		return
	}
	if len(patch.OperationId) > 0 {
		o.OperationId = patch.OperationId
	}
	if len(patch.Summary) > 0 {
		o.Summary = patch.Summary
	}
	if len(patch.Description) > 0 {
		o.Description = patch.Description
	}
	if o.Message == nil {
		o.Message = patch.Message
	} else {
		o.Message.patch(patch.Message)
	}
}

func (r *MessageRef) patch(patch *MessageRef) {
	if patch == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value
	} else {
		r.Value.patch(patch.Value)
	}
}

func (m *Message) patch(patch *Message) {
	if patch == nil {
		return
	}
	if len(patch.Name) > 0 {
		m.Name = patch.Name
	}
	if len(patch.Title) > 0 {
		m.Title = patch.Title
	}
	if len(patch.Summary) > 0 {
		m.Summary = patch.Summary
	}
	if len(patch.Description) > 0 {
		m.Description = patch.Description
	}
	if len(patch.ContentType) > 0 {
		m.ContentType = patch.ContentType
	}
	if m.Payload == nil {
		m.Payload = patch.Payload
	} else {
		m.Payload.Patch(patch.Payload)
	}
	if m.Headers == nil {
		m.Headers = patch.Headers
	} else {
		m.Headers.Patch(patch.Headers)
	}
	m.Bindings.Kafka.Patch(patch.Bindings.Kafka)
}

func (c *Config) patchComponents(patch *Config) {
	if patch.Components == nil {
		return
	}
	if c.Components == nil {
		c.Components = patch.Components
		return
	}

	if c.Components.Servers == nil {
		c.Components.Servers = patch.Components.Servers
	} else {
		for k, p := range patch.Components.Servers {
			if v, ok := c.Components.Servers[k]; ok {
				v.patch(p)
			} else {
				c.Components.Servers[k] = p
			}
		}
	}

	if c.Components.Schemas == nil {
		c.Components.Schemas = patch.Components.Schemas
	} else {
		c.Components.Schemas.Patch(patch.Components.Schemas)
	}

	if c.Components.Messages == nil {
		c.Components.Messages = patch.Components.Messages
	} else {
		for k, p := range patch.Components.Messages {
			if v, ok := c.Components.Messages[k]; ok {
				v.patch(p)
			} else {
				c.Components.Messages[k] = p
			}
		}
	}
}
