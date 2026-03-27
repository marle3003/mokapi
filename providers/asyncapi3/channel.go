package asyncapi3

import (
	"mokapi/config/dynamic"

	"gopkg.in/yaml.v3"
)

type ChannelRef struct {
	dynamic.Reference
	Value *Channel
}

type Channel struct {
	Name        string                   `yaml:"-" json:"-"`
	Title       string                   `yaml:"title" json:"title"`
	Address     string                   `yaml:"address" json:"address"`
	Summary     string                   `yaml:"summary" json:"summary"`
	Description string                   `yaml:"description" json:"description"`
	Servers     []*ServerRef             `yaml:"servers" json:"servers"`
	Messages    map[string]*MessageRef   `yaml:"messages" json:"messages"`
	Parameters  map[string]*ParameterRef `yaml:"parameters" json:"parameters"`
	Bindings    ChannelBindings          `yaml:"bindings" json:"bindings"`

	Tags         []*TagRef        `yaml:"tags" json:"tags"`
	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
	Config       *Config
}

func (r *ChannelRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *ChannelRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *ChannelRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}
	if len(r.Ref) > 0 {
		var resolved *ChannelRef
		if err := dynamic.Resolve(r.Ref, &resolved, config, reader); err != nil {
			return err
		}
		r.Value = resolved.Value
		return nil
	}
	return r.Value.Parse(config, reader)
}

func (c *Channel) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if c == nil {
		return nil
	}

	for _, s := range c.Servers {
		if err := s.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, msg := range c.Messages {
		if err := msg.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, p := range c.Parameters {
		if err := p.Parse(config, reader); err != nil {
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

func (c *Channel) GetName() string {
	if c.Address != "" {
		return c.Address
	}
	if c.Name != "" {
		return c.Name
	}
	return c.Title
}
