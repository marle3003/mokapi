package asyncApi

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type ChannelTrait struct {
	Title       string                   `yaml:"title" json:"title"`
	Address     string                   `yaml:"address" json:"address"`
	Summary     string                   `yaml:"summary" json:"summary"`
	Description string                   `yaml:"description" json:"description"`
	Servers     []*ServerRef             `yaml:"servers" json:"servers"`
	Messages    map[string]*MessageRef   `yaml:"messages" json:"messages"`
	Parameters  map[string]*ParameterRef `yaml:"parameters" json:"parameters"`
	Bindings    ChannelBindings          `yaml:"bindings" json:"bindings"`
}

func (c *Channel) UnmarshalYAML(node *yaml.Node) error {
	// set default
	c.Bindings.Kafka.ValueSchemaValidation = true
	c.Bindings.Kafka.KeySchemaValidation = true
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
