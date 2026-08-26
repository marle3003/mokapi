package asyncapi3

import (
	"gopkg.in/yaml.v3"
)

func (b *BrokerBindings) UnmarshalYAML(value *yaml.Node) error {
	b.configs = make(map[string]any)
	err := value.Decode(b.configs)
	if err != nil {
		return err
	}

	return b.unmarshal()
}

func (b *BrokerBindings) MarshalYAML() (any, error) {
	return b.configs, nil
}

func (t *TopicBindings) UnmarshalYAML(value *yaml.Node) error {
	t.configs = make(map[string]any)
	err := value.Decode(t.configs)
	if err != nil {
		return err
	}

	return t.unmarshal(t.configs)
}

func (t *TopicBindings) MarshalYAML() (any, error) {
	return t.marshal(), nil
}
