package asyncapi3

import (
	"mokapi/config/dynamic"

	"gopkg.in/yaml.v3"
)

type OperationRef struct {
	dynamic.Reference
	Value *Operation
}

type Operation struct {
	Action      string               `yaml:"action" json:"action"`
	Channel     ChannelRef           `yaml:"channel" json:"channel"`
	Title       string               `yaml:"title" json:"title"`
	Summary     string               `yaml:"summary" json:"summary"`
	Description string               `yaml:"description" json:"description"`
	Bindings    OperationBindings    `yaml:"bindings" json:"bindings"`
	Traits      []*OperationTraitRef `yaml:"traits" json:"traits"`
	Messages    []*MessageRef        `yaml:"messages" json:"messages"`

	ExternalDocs []*ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

type OperationTraitRef struct {
	dynamic.Reference
	Value *OperationTrait
}

type OperationTrait struct {
	Channel     *ChannelRef       `yaml:"channel" json:"channel"`
	Title       string            `yaml:"title" json:"title"`
	Summary     string            `yaml:"summary" json:"summary"`
	Description string            `yaml:"description" json:"description"`
	Bindings    OperationBindings `yaml:"bindings" json:"bindings"`

	ExternalDocs []*ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
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

func (r *OperationRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		var resolved *OperationRef
		if err := dynamic.Resolve(r.Ref, &resolved, config, reader); err != nil {
			return err
		}
		r.Value = resolved.Value
		return nil
	}
	return r.Value.Parse(config, reader)
}

func (o *Operation) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if o == nil {
		return nil
	}

	if len(o.Channel.Ref) > 0 {
		var resolved *ChannelRef
		if err := dynamic.Resolve(o.Channel.Ref, &resolved, config, reader); err != nil {
			return err
		}
		o.Channel.Value = resolved.Value
	}

	for _, msg := range o.Messages {
		if err := msg.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, trait := range o.Traits {
		if err := trait.Parse(config, reader); err != nil {
			return err
		}
		o.applyTrait(trait.Value)
	}

	return nil
}

func (r *OperationTraitRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		var resolved *OperationTraitRef
		if err := dynamic.Resolve(r.Ref, &resolved, config, reader); err != nil {
			return err
		}
		r.Value = resolved.Value
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
