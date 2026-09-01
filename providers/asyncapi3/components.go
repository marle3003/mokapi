package asyncapi3

import (
	"mokapi/config/dynamic"
)

type Components struct {
	Servers         map[string]*ServerRef         `yaml:"servers,omitempty" json:"servers,omitempty"`
	Tags            map[string]*TagRef            `yaml:"tags,omitempty" json:"tags,omitempty"`
	Channels        map[string]*ChannelRef        `yaml:"channels,omitempty" json:"channels,omitempty"`
	Schemas         map[string]*SchemaRef         `yaml:"schemas,omitempty" json:"schemas,omitempty"`
	Messages        map[string]*MessageRef        `yaml:"messages,omitempty" json:"messages,omitempty"`
	Operations      map[string]*OperationRef      `yaml:"operations,omitempty" json:"operations,omitempty"`
	Parameters      map[string]*ParameterRef      `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	CorrelationIds  map[string]*CorrelationIdRef  `yaml:"correlationIds,omitempty" json:"correlationIds,omitempty"`
	ExternalDocs    map[string]*ExternalDocRef    `yaml:"externalDocs,omitempty" json:"externalDocs,omitempty"`
	OperationTraits map[string]*OperationTraitRef `yaml:"operationTraits,omitempty" json:"operationTraits,omitempty"`
	MessageTraits   map[string]*MessageTraitRef   `yaml:"messageTraits,omitempty" json:"messageTraits,omitempty"`
	ServerVariables map[string]*ServerVariableRef `yaml:"serverVariables,omitempty" json:"serverVariables,omitempty"`
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
