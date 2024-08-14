package asyncApi

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type Channel3Ref struct {
	dynamic.Reference
	Value *Channel3
}

type Channel3 struct {
	Title       string                    `yaml:"title" json:"title"`
	Address     string                    `yaml:"address" json:"address"`
	Summary     string                    `yaml:"summary" json:"summary"`
	Description string                    `yaml:"description" json:"description"`
	Servers     []*Server3Ref             `yaml:"servers" json:"servers"`
	Messages    map[string]*Message3Ref   `yaml:"messages" json:"messages"`
	Parameters  map[string]*Parameter3Ref `yaml:"parameters" json:"parameters"`
	Bindings    ChannelBindings           `yaml:"bindings" json:"bindings"`

	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

type ChannelTrait struct {
	Title       string                    `yaml:"title" json:"title"`
	Address     string                    `yaml:"address" json:"address"`
	Summary     string                    `yaml:"summary" json:"summary"`
	Description string                    `yaml:"description" json:"description"`
	Servers     []*Server3Ref             `yaml:"servers" json:"servers"`
	Messages    map[string]*Message3Ref   `yaml:"messages" json:"messages"`
	Parameters  map[string]*Parameter3Ref `yaml:"parameters" json:"parameters"`
	Bindings    ChannelBindings           `yaml:"bindings" json:"bindings"`

	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

func (r *Channel3Ref) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *Channel3Ref) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *Channel3Ref) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
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
