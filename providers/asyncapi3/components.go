package asyncapi3

import "mokapi/config/dynamic"

type Components struct {
	Servers         map[string]*ServerRef         `yaml:"servers" json:"servers"`
	Tags            map[string]*TagRef            `yaml:"tags" json:"tags"`
	Channels        map[string]*ChannelRef        `yaml:"channels" json:"channels"`
	Schemas         map[string]*SchemaRef         `yaml:"schemas" json:"schemas"`
	Messages        map[string]*MessageRef        `yaml:"messages" json:"messages"`
	Operations      map[string]*OperationRef      `yaml:"operations" json:"operations"`
	Parameters      map[string]*ParameterRef      `yaml:"parameters" json:"parameters"`
	CorrelationIds  map[string]*CorrelationIdRef  `yaml:"correlationIds" json:"correlationIds"`
	ExternalDocs    map[string]*ExternalDocRef    `yaml:"externalDocs" json:"externalDocs"`
	OperationTraits map[string]*OperationTraitRef `yaml:"operationTraits" json:"operationTraits"`
	MessageTraits   map[string]*MessageTraitRef   `yaml:"messageTraits" json:"messageTraits"`
}

func (c *Components) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if c == nil {
		return nil
	}

	for _, s := range c.Servers {
		if err := s.parse(config, reader); err != nil {
			return err
		}
	}

	for _, t := range c.Tags {
		if err := t.parse(config, reader); err != nil {
			return err
		}
	}

	for name, ch := range c.Channels {
		if err := ch.parse(config, reader); err != nil {
			return err
		}
		if ch.Value != nil {
			ch.Value.Name = name
		}
	}

	for _, s := range c.Schemas {
		if err := s.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, m := range c.Messages {
		if err := m.parse(config, reader); err != nil {
			return err
		}
	}

	for _, o := range c.Operations {
		if err := o.parse(config, reader); err != nil {
			return err
		}
	}

	for _, p := range c.Parameters {
		if err := p.parse(config, reader); err != nil {
			return err
		}
	}

	for _, cId := range c.CorrelationIds {
		if err := cId.parse(config, reader); err != nil {
			return err
		}
	}

	for _, d := range c.ExternalDocs {
		if err := d.parse(config, reader); err != nil {
			return err
		}
	}

	for _, trait := range c.OperationTraits {
		if err := trait.parse(config, reader); err != nil {
			return err
		}
	}

	for _, trait := range c.MessageTraits {
		if err := trait.parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}
