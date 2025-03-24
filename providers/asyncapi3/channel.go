package asyncapi3

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type ChannelRef struct {
	dynamic.Reference
	Value *Channel
}

type Channel struct {
	Title       string                   `yaml:"title" json:"title"`
	Address     string                   `yaml:"address" json:"address"`
	Summary     string                   `yaml:"summary" json:"summary"`
	Description string                   `yaml:"description" json:"description"`
	Servers     []*ServerRef             `yaml:"servers" json:"servers"`
	Messages    map[string]*MessageRef   `yaml:"messages" json:"messages"`
	Parameters  map[string]*ParameterRef `yaml:"parameters" json:"parameters"`
	Bindings    ChannelBindings          `yaml:"bindings" json:"bindings"`

	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

type ChannelTrait struct {
	Title       string                   `yaml:"title" json:"title"`
	Address     string                   `yaml:"address" json:"address"`
	Summary     string                   `yaml:"summary" json:"summary"`
	Description string                   `yaml:"description" json:"description"`
	Servers     []*ServerRef             `yaml:"servers" json:"servers"`
	Messages    map[string]*MessageRef   `yaml:"messages" json:"messages"`
	Parameters  map[string]*ParameterRef `yaml:"parameters" json:"parameters"`
	Bindings    ChannelBindings          `yaml:"bindings" json:"bindings"`

	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

func (r *ChannelRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *ChannelRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *ChannelRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}

	if r.Value == nil {
		return nil
	}

	for _, s := range r.Value.Servers {
		if err := s.parse(config, reader); err != nil {
			return err
		}
	}

	for _, msg := range r.Value.Messages {
		if err := msg.parse(config, reader); err != nil {
			return err
		}
	}

	for _, p := range r.Value.Parameters {
		if err := p.parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (c *Channel) UnmarshalYAML(node *yaml.Node) error {
	// set default
	c.Bindings.Kafka.ValueSchemaValidation = true
	c.Bindings.Kafka.Partitions = 1

	type alias Channel
	a := alias(*c)
	err := node.Decode(&a)
	if err != nil {
		return err
	}
	*c = Channel(a)
	return nil
}

func (c *Channel) UnmarshalJSON(b []byte) error {
	// set default
	c.Bindings.Kafka.ValueSchemaValidation = true
	c.Bindings.Kafka.KeySchemaValidation = true
	c.Bindings.Kafka.Partitions = 1

	type alias Channel
	a := alias(*c)
	err := dynamic.UnmarshalJSON(b, &a)
	if err != nil {
		return err
	}
	*c = Channel(a)
	return nil
}
