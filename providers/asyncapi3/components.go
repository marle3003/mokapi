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

	for _, t := range c.Tags {
		if err := t.parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}
