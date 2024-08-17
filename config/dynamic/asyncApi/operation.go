package asyncApi

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type Operation3Ref struct {
	dynamic.Reference
	Value *Operation3
}

type Operation3 struct {
	Action      string              `yaml:"action" json:"action"`
	Channel     Channel3Ref         `yaml:"channel" json:"channel"`
	Title       string              `yaml:"title" json:"title"`
	Summary     string              `yaml:"summary" json:"summary"`
	Description string              `yaml:"description" json:"description"`
	Bindings    OperationBindings   `yaml:"bindings" json:"bindings"`
	Traits      []OperationTraitRef `yaml:"traits" json:"traits"`
	Messages    []Message3Ref       `yaml:"messages" json:"messages"`

	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

type OperationTraitRef struct {
	dynamic.Reference
	Value *OperationTrait
}

type OperationTrait struct {
	Channel     Channel3Ref       `yaml:"channel" json:"channel"`
	Title       string            `yaml:"title" json:"title"`
	Summary     string            `yaml:"summary" json:"summary"`
	Description string            `yaml:"description" json:"description"`
	Bindings    OperationBindings `yaml:"bindings" json:"bindings"`

	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

func (r *Operation3Ref) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *Operation3Ref) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *OperationTraitRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *OperationTraitRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *Operation3Ref) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
	}

	if r.Value == nil {
		return nil
	}

	for _, msg := range r.Value.Messages {
		if err := msg.parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (r *OperationTraitRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}

	if r.Value == nil {
		return nil
	}

	return nil
}
