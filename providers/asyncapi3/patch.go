package asyncapi3

import "maps"

// Patch applies patch to c. Everything c takes over from patch is copied, so c never shares
// data with patch and patching does not change the config that was parsed from a file.
func (c *Config) Patch(patch *Config) {
	cp := copies{}
	c.patchHeader(patch)
	c.patchInfo(patch, cp)
	c.patchServer(patch, cp)
	c.patchChannels(patch, cp)
	c.patchOperations(patch, cp)
	c.patchComponents(patch, cp)
}

// patchHeader keeps the values of the first config: every file has them,
// DefaultContentType even by default.
func (c *Config) patchHeader(patch *Config) {
	if c.Version == "" {
		c.Version = patch.Version
	}
	if c.Id == "" {
		c.Id = patch.Id
	}
	if c.DefaultContentType == "" {
		c.DefaultContentType = patch.DefaultContentType
	}
}

func (c *Config) patchInfo(patch *Config, cp copies) {
	if c.Info.Name == "" {
		c.Info.Name = patch.Info.Name
	}
	if len(patch.Info.Description) > 0 {
		c.Info.Description = patch.Info.Description
	}
	if len(patch.Info.Version) > 0 {
		c.Info.Version = patch.Info.Version
	}
	if c.Info.Contact == nil {
		c.Info.Contact = clonePointer(patch.Info.Contact)
	} else {
		c.Info.Contact.patch(patch.Info.Contact)
	}
	if len(patch.Info.TermsOfService) > 0 {
		c.Info.TermsOfService = patch.Info.TermsOfService
	}
	if c.Info.License == nil {
		c.Info.License = clonePointer(patch.Info.License)
	} else {
		c.Info.License.patch(patch.Info.License)
	}
	if len(c.Info.ExternalDocs) == 0 {
		c.Info.ExternalDocs = cloneExternalDocRefs(patch.Info.ExternalDocs, cp)
	}
}

func (c *Contact) patch(patch *Contact) {
	if patch == nil {
		return
	}
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
	if patch == nil {
		return
	}
	if len(patch.Name) > 0 {
		l.Name = patch.Name
	}
	if len(patch.Url) > 0 {
		l.Url = patch.Url
	}
}

func (c *Config) patchServer(patch *Config, cp copies) {
	if c.Servers.Len() == 0 {
		c.Servers = cloneServers(patch.Servers, cp)
	} else {
		for it := patch.Servers.Iter(); it.Next(); {
			if s, ok := c.Servers.Get(it.Key()); ok && s != nil {
				s.patch(it.Value(), cp)
			} else {
				c.Servers.Set(it.Key(), it.Value().clone(cp))
			}
		}
	}
}

func (r *ServerRef) patch(patch *ServerRef, cp copies) {
	if patch == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value.clone(cp)
	} else {
		r.Value.patch(patch.Value)
	}
}

func (s *Server) patch(patch *Server) {
	if patch == nil {
		return
	}

	if len(patch.Host) > 0 {
		s.Host = patch.Host
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

	if s.Bindings != nil {
		s.Bindings.patch(patch.Bindings)
	} else {
		s.Bindings = patch.Bindings.clone()
	}
}

func (c *Config) patchChannels(patch *Config, cp copies) {
	c.Channels = patchMap(c.Channels, patch.Channels, cp, (*ChannelRef).clone, (*ChannelRef).patch)
	// the channels belong to c now, like after Parse
	for _, ch := range c.Channels {
		if ch != nil && ch.Value != nil {
			ch.Value.Config = c
		}
	}
}

func (r *ChannelRef) patch(patch *ChannelRef, cp copies) {
	if patch == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value.clone(cp)
	} else {
		r.Value.patch(patch.Value, cp)
	}
}

func (c *Channel) patch(patch *Channel, cp copies) {
	if patch == nil {
		return
	}
	if len(patch.Description) > 0 {
		c.Description = patch.Description
	}

	c.Messages = patchMap(c.Messages, patch.Messages, cp, (*MessageRef).clone, (*MessageRef).patch)

	if c.Bindings != nil {
		c.Bindings.patch(patch.Bindings)
	} else {
		c.Bindings = patch.Bindings.clone()
	}
}

func (c *Config) patchOperations(patch *Config, cp copies) {
	c.Operations = patchMap(c.Operations, patch.Operations, cp, (*OperationRef).clone, (*OperationRef).patch)
}

func (r *OperationRef) patch(patch *OperationRef, cp copies) {
	if patch == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value.clone(cp)
	} else {
		r.Value.patch(patch.Value)
	}
}

func (o *Operation) patch(patch *Operation) {
	if patch == nil {
		return
	}
	if len(patch.Title) > 0 {
		o.Title = patch.Title
	}
	if len(patch.Summary) > 0 {
		o.Summary = patch.Summary
	}
	if len(patch.Description) > 0 {
		o.Description = patch.Description
	}
	if o.Bindings != nil {
		o.Bindings.patch(patch.Bindings)
	} else {
		o.Bindings = patch.Bindings.clone()
	}
}

func (r *MessageRef) patch(patch *MessageRef, cp copies) {
	if patch == nil {
		return
	}
	if r.Value == nil {
		r.Value = patch.Value.clone(cp)
	} else {
		r.Value.patch(patch.Value, cp)
	}
}

func (m *Message) patch(patch *Message, cp copies) {
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
		m.Payload = patch.Payload.clone(cp)
	} else {
		m.Payload.Patch(patch.Payload)
	}
	if m.Headers == nil {
		m.Headers = patch.Headers.clone(cp)
	} else {
		m.Headers.Patch(patch.Headers)
	}
	if m.Bindings != nil {
		m.Bindings.patch(patch.Bindings)
	} else {
		m.Bindings = patch.Bindings.clone(cp)
	}
}

func (c *Config) patchComponents(patch *Config, cp copies) {
	if patch.Components == nil {
		return
	}
	if c.Components == nil {
		c.Components = patch.Components.clone(cp)
		return
	}

	p := patch.Components
	c.Components.Servers = patchMap(c.Components.Servers, p.Servers, cp, (*ServerRef).clone, (*ServerRef).patch)
	c.Components.Tags = patchMap(c.Components.Tags, p.Tags, cp, (*TagRef).clone, nil)
	c.Components.Channels = patchMap(c.Components.Channels, p.Channels, cp, (*ChannelRef).clone, (*ChannelRef).patch)
	c.Components.Schemas = patchMap(c.Components.Schemas, p.Schemas, cp, (*SchemaRef).clone,
		func(r, patch *SchemaRef, _ copies) { r.Patch(patch) })
	c.Components.Messages = patchMap(c.Components.Messages, p.Messages, cp, (*MessageRef).clone, (*MessageRef).patch)
	c.Components.Operations = patchMap(c.Components.Operations, p.Operations, cp, (*OperationRef).clone, (*OperationRef).patch)
	c.Components.Parameters = patchMap(c.Components.Parameters, p.Parameters, cp, (*ParameterRef).clone, nil)
	c.Components.CorrelationIds = patchMap(c.Components.CorrelationIds, p.CorrelationIds, cp, (*CorrelationIdRef).clone, nil)
	c.Components.ExternalDocs = patchMap(c.Components.ExternalDocs, p.ExternalDocs, cp, (*ExternalDocRef).clone, nil)
	c.Components.OperationTraits = patchMap(c.Components.OperationTraits, p.OperationTraits, cp, (*OperationTraitRef).clone, nil)
	c.Components.MessageTraits = patchMap(c.Components.MessageTraits, p.MessageTraits, cp, (*MessageTraitRef).clone, nil)
	c.Components.ServerVariables = patchMap(c.Components.ServerVariables, p.ServerVariables, cp, (*ServerVariableRef).clone, nil)
}

// patchMap adds a copy of the entries of patch that m does not have (or has as nil) and merges the others
// with merge. Entries without merge function and nil entries of patch keep the value of m.
func patchMap[V comparable](m, patch map[string]V, cp copies, clone func(V, copies) V, merge func(V, V, copies)) map[string]V {
	if m == nil {
		return cloneMap(patch, cp, clone)
	}
	var nothing V
	for k, p := range patch {
		v := m[k]
		switch {
		case v == nothing:
			m[k] = clone(p, cp)
		case p != nothing && merge != nil:
			merge(v, p, cp)
		}
	}
	return m
}

func (b *BrokerBindings) Patch(patch BrokerBindings) {
	if patch.LogRetentionBytes != 0 {
		b.LogRetentionBytes = patch.LogRetentionBytes
	}
	if patch.LogRetentionMs != 0 {
		b.LogRetentionMs = patch.LogRetentionMs
	}
	if patch.LogRetentionCheckIntervalMs != 0 {
		b.LogRetentionCheckIntervalMs = patch.LogRetentionCheckIntervalMs
	}
	if patch.LogSegmentDeleteDelayMs != 0 {
		b.LogSegmentDeleteDelayMs = patch.LogSegmentDeleteDelayMs
	}
	if patch.LogRollMs != 0 {
		b.LogRollMs = patch.LogRollMs
	}
	if patch.LogSegmentBytes != 0 {
		b.LogSegmentBytes = patch.LogSegmentBytes
	}
	if patch.GroupInitialRebalanceDelayMs != 0 {
		b.GroupInitialRebalanceDelayMs = patch.GroupInitialRebalanceDelayMs
	}
	if patch.GroupMinSessionTimeoutMs != 0 {
		b.GroupMinSessionTimeoutMs = patch.GroupMinSessionTimeoutMs
	}
	if len(patch.SchemaRegistryUrl) > 0 {
		b.SchemaRegistryUrl = patch.SchemaRegistryUrl
	}
	if len(patch.SchemaRegistryVendor) > 0 {
		b.SchemaRegistryVendor = patch.SchemaRegistryVendor
	}
}

func (t *TopicBindings) Patch(patch TopicBindings) {
	if t.configs == nil {
		t.configs = maps.Clone(patch.configs)
	} else {
		maps.Copy(t.configs, patch.configs)
	}

	if patch.Partitions != 0 {
		t.Partitions = patch.Partitions
	}
	if patch.RetentionBytes != 0 {
		t.RetentionBytes = patch.RetentionBytes
	}
	if patch.RetentionMs != 0 {
		t.RetentionMs = patch.RetentionMs
	}
	if patch.SegmentBytes != 0 {
		t.SegmentBytes = patch.SegmentBytes
	}
	if patch.SegmentMs != 0 {
		t.SegmentMs = patch.SegmentMs
	}

	t.ValueSchemaValidation = patch.ValueSchemaValidation
}

func (m *KafkaMessageBinding) Patch(patch KafkaMessageBinding) {
	if m.Key != nil {
		m.Key.Patch(patch.Key)
	} else {
		m.Key = patch.Key.clone(copies{})
	}
}

func (o *KafkaOperationBinding) Patch(patch KafkaOperationBinding) {
	if o.ClientId != nil {
		o.ClientId.Patch(patch.ClientId)
	} else {
		o.ClientId = patch.ClientId.Clone()
	}

	if o.GroupId != nil {
		o.GroupId.Patch(patch.GroupId)
	} else {
		o.GroupId = patch.GroupId.Clone()
	}
}
