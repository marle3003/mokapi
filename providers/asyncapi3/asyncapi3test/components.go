package asyncapi3test

import "mokapi/providers/asyncapi3"

func WithComponentServer(name string, s *asyncapi3.Server) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Components == nil {
			c.Components = &asyncapi3.Components{}
		}
		if c.Components.Servers == nil {
			c.Components.Servers = map[string]*asyncapi3.ServerRef{}
		}
		c.Components.Servers[name] = &asyncapi3.ServerRef{
			Value: s,
		}
	}
}

func WithComponentTag(name string, tag *asyncapi3.Tag) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Components == nil {
			c.Components = &asyncapi3.Components{}
		}
		if c.Components.Tags == nil {
			c.Components.Tags = make(map[string]*asyncapi3.TagRef)
		}
		c.Components.Tags[name] = &asyncapi3.TagRef{
			Value: tag,
		}
	}
}

func WithComponentChannel(name string, ch *asyncapi3.Channel) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Components == nil {
			c.Components = &asyncapi3.Components{}
		}
		if c.Components.Channels == nil {
			c.Components.Channels = map[string]*asyncapi3.ChannelRef{}
		}
		c.Components.Channels[name] = &asyncapi3.ChannelRef{
			Value: ch,
		}
	}
}

func WithComponentSchema(name string, s asyncapi3.Schema) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Components == nil {
			c.Components = &asyncapi3.Components{}
		}
		if c.Components.Schemas == nil {
			c.Components.Schemas = map[string]*asyncapi3.SchemaRef{}
		}
		c.Components.Schemas[name] = &asyncapi3.SchemaRef{
			Value: &asyncapi3.MultiSchemaFormat{Schema: s},
		}
	}
}

func WithComponentMessage(name string, m *asyncapi3.Message) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Components == nil {
			c.Components = &asyncapi3.Components{}
		}
		if c.Components.Messages == nil {
			c.Components.Messages = map[string]*asyncapi3.MessageRef{}
		}
		c.Components.Messages[name] = &asyncapi3.MessageRef{
			Value: m,
		}
	}
}

func WithComponentOperation(name string, o *asyncapi3.Operation) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Components == nil {
			c.Components = &asyncapi3.Components{}
		}
		if c.Components.Operations == nil {
			c.Components.Operations = map[string]*asyncapi3.OperationRef{}
		}
		c.Components.Operations[name] = &asyncapi3.OperationRef{
			Value: o,
		}
	}
}

func WithComponentParameter(name string, p *asyncapi3.Parameter) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Components == nil {
			c.Components = &asyncapi3.Components{}
		}
		if c.Components.Parameters == nil {
			c.Components.Parameters = map[string]*asyncapi3.ParameterRef{}
		}
		c.Components.Parameters[name] = &asyncapi3.ParameterRef{
			Value: p,
		}
	}
}

func WithComponentCorrelationId(name string, cId *asyncapi3.CorrelationId) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Components == nil {
			c.Components = &asyncapi3.Components{}
		}
		if c.Components.CorrelationIds == nil {
			c.Components.CorrelationIds = map[string]*asyncapi3.CorrelationIdRef{}
		}
		c.Components.CorrelationIds[name] = &asyncapi3.CorrelationIdRef{
			Value: cId,
		}
	}
}

func WithComponentExternalDoc(name string, d *asyncapi3.ExternalDoc) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Components == nil {
			c.Components = &asyncapi3.Components{}
		}
		if c.Components.ExternalDocs == nil {
			c.Components.ExternalDocs = map[string]*asyncapi3.ExternalDocRef{}
		}
		c.Components.ExternalDocs[name] = &asyncapi3.ExternalDocRef{
			Value: d,
		}
	}
}

func WithComponentOperationTrait(name string, trait *asyncapi3.OperationTrait) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Components == nil {
			c.Components = &asyncapi3.Components{}
		}
		if c.Components.OperationTraits == nil {
			c.Components.OperationTraits = map[string]*asyncapi3.OperationTraitRef{}
		}
		c.Components.OperationTraits[name] = &asyncapi3.OperationTraitRef{
			Value: trait,
		}
	}
}

func WithComponentMessageTrait(name string, trait *asyncapi3.MessageTrait) ConfigOptions {
	return func(c *asyncapi3.Config) {
		if c.Components == nil {
			c.Components = &asyncapi3.Components{}
		}
		if c.Components.MessageTraits == nil {
			c.Components.MessageTraits = map[string]*asyncapi3.MessageTraitRef{}
		}
		c.Components.MessageTraits[name] = &asyncapi3.MessageTraitRef{
			Value: trait,
		}
	}
}
