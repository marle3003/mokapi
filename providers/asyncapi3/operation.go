package asyncapi3

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type OperationRef struct {
	dynamic.Reference
	Value *Operation
}

type Operation struct {
	Action      string              `yaml:"action" json:"action"`
	Channel     ChannelRef          `yaml:"channel" json:"channel"`
	Title       string              `yaml:"title" json:"title"`
	Summary     string              `yaml:"summary" json:"summary"`
	Description string              `yaml:"description" json:"description"`
	Bindings    OperationBindings   `yaml:"bindings" json:"bindings"`
	Traits      []OperationTraitRef `yaml:"traits" json:"traits"`
	Messages    []MessageRef        `yaml:"messages" json:"messages"`

	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

type OperationTraitRef struct {
	dynamic.Reference
	Value *OperationTrait
}

type OperationTrait struct {
	Channel     ChannelRef        `yaml:"channel" json:"channel"`
	Title       string            `yaml:"title" json:"title"`
	Summary     string            `yaml:"summary" json:"summary"`
	Description string            `yaml:"description" json:"description"`
	Bindings    OperationBindings `yaml:"bindings" json:"bindings"`

	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

func (r *OperationRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *OperationRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *OperationTraitRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *OperationTraitRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *OperationRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
	}

	if r.Value == nil {
		return nil
	}

	if len(r.Value.Channel.Ref) > 0 {
		if err := dynamic.Resolve(r.Value.Channel.Ref, &r.Value.Channel.Value, config, reader); err != nil {
			return err
		}
	}

	for _, msg := range r.Value.Messages {
		if err := msg.parse(config, reader); err != nil {
			return err
		}
	}

	for _, trait := range r.Value.Traits {
		if err := trait.parse(config, reader); err != nil {
			return err
		}
		r.Value.applyTrait(trait.Value)
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

func (o *Operation) applyTrait(trait *OperationTrait) {
	if trait == nil {
		return
	}

	if len(o.Title) == 0 {
		o.Title = trait.Title
	}
	if len(o.Summary) == 0 {
		o.Summary = trait.Summary
	}
	if len(o.Description) == 0 {
		o.Description = trait.Description
	}

	o.ExternalDocs = append(o.ExternalDocs, trait.ExternalDocs...)

	if o.Bindings.Kafka.ClientId == nil {
		o.Bindings.Kafka.ClientId = trait.Bindings.Kafka.ClientId
	}
	if o.Bindings.Kafka.GroupId == nil {
		o.Bindings.Kafka.GroupId = trait.Bindings.Kafka.GroupId
	}
}
