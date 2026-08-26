package asyncapi3

import (
	"encoding/json"
	"mokapi/config/dynamic"

	"gopkg.in/yaml.v3"
)

type OperationRef struct {
	dynamic.Reference[*OperationRef]
	Value *Operation
}

type Operation struct {
	Action      string               `yaml:"action,omitempty" json:"action,omitempty"`
	Channel     *ChannelRef          `yaml:"channel,omitempty" json:"channel,omitempty"`
	Title       string               `yaml:"title,omitempty" json:"title,omitempty"`
	Summary     string               `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string               `yaml:"description,omitempty" json:"description,omitempty"`
	Bindings    *OperationBindings   `yaml:"bindings,omitempty" json:"bindings,omitempty"`
	Traits      []*OperationTraitRef `yaml:"traits,omitempty" json:"traits,omitempty"`
	Messages    []*MessageRef        `yaml:"messages,omitempty" json:"messages,omitempty"`

	ExternalDocs []*ExternalDocRef `yaml:"externalDocs,omitempty" json:"externalDocs,omitempty"`
}

type OperationTraitRef struct {
	dynamic.Reference[*OperationTraitRef]
	Value *OperationTrait
}

type OperationTrait struct {
	Channel     *ChannelRef        `yaml:"channel,omitempty" json:"channel,omitempty"`
	Title       string             `yaml:"title,omitempty" json:"title,omitempty"`
	Summary     string             `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string             `yaml:"description,omitempty" json:"description,omitempty"`
	Bindings    *OperationBindings `yaml:"bindings,omitempty" json:"bindings,omitempty"`

	ExternalDocs []*ExternalDocRef `yaml:"externalDocs,omitempty" json:"externalDocs,omitempty"`
}

func (r *OperationRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *OperationRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *OperationRef) MarshalJSON() ([]byte, error) {
	if r.Value != nil {
		return json.Marshal(r.Value)
	}
	return json.Marshal(r.Reference)
}

func (r *OperationRef) MarshalYAML() (any, error) {
	if r.Value != nil {
		return r.Value, nil
	}
	return r.Reference, nil
}

func (r *OperationTraitRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *OperationTraitRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *OperationTraitRef) MarshalJSON() ([]byte, error) {
	if r.Value != nil {
		return json.Marshal(r.Value)
	}
	return json.Marshal(r.Reference)
}

func (r *OperationTraitRef) MarshalYAML() (any, error) {
	if r.Value != nil {
		return r.Value, nil
	}
	return r.Reference, nil
}

func (r *OperationRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 {
		resolved, err := r.Resolve(config, reader)
		if err != nil {
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

	if o.Channel != nil && len(o.Channel.Ref) > 0 {
		r := dynamic.Reference[ChannelRef]{Ref: o.Channel.Ref}
		resolved, err := r.Resolve(config, reader)
		if err != nil {
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
		resolved, err := r.Resolve(config, reader)
		if err != nil {
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

	if o.Bindings == nil {
		o.Bindings = trait.Bindings
	} else {
		if o.Bindings.Kafka.ClientId == nil {
			o.Bindings.Kafka.ClientId = trait.Bindings.Kafka.ClientId
		}
		if o.Bindings.Kafka.GroupId == nil {
			o.Bindings.Kafka.GroupId = trait.Bindings.Kafka.GroupId
		}
	}
}
