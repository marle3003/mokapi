package asyncApi

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type Message3Ref struct {
	dynamic.Reference
	Value *Message3
}

type Message3 struct {
	Title       string `yaml:"title" json:"title"`
	Name        string `yaml:"name" json:"name"`
	Summary     string `yaml:"summary" json:"summary"`
	Description string `yaml:"description" json:"description"`
	Deprecated  bool   `yaml:"deprecated" json:"deprecated"`

	CorrelationId *CorrelationIdRef `yaml:"correlationId" json:"correlationId"`

	ContentType string             `yaml:"contentType" json:"contentType"`
	Headers     *SchemaRef         `yaml:"headers" json:"headers"`
	Payload     *SchemaRef         `yaml:"payload" json:"payload"`
	Bindings    MessageBinding     `yaml:"bindings" json:"bindings"`
	Traits      []*MessageTraitRef `yaml:"traits" json:"traits"`

	Examples []interface{} `yaml:"examples" json:"examples"`

	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

type MessageTraitRef struct {
	dynamic.Reference
	Value *MessageTrait
}

type MessageTrait struct {
	Title       string `yaml:"title" json:"title"`
	Name        string `yaml:"name" json:"name"`
	Summary     string `yaml:"summary" json:"summary"`
	Description string `yaml:"description" json:"description"`
	Deprecated  bool   `yaml:"deprecated" json:"deprecated"`

	CorrelationId string `yaml:"correlationId" json:"correlationId"`

	ContentType string         `yaml:"contentType" json:"contentType"`
	Headers     *SchemaRef     `yaml:"headers" json:"headers"`
	Bindings    MessageBinding `yaml:"bindings" json:"bindings"`

	Examples []interface{} `yaml:"examples" json:"examples"`

	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

func (r *Message3Ref) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *Message3Ref) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *MessageTraitRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *MessageTraitRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *Message3Ref) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
	}

	if r.Value == nil {
		return nil
	}

	for _, trait := range r.Value.Traits {
		if err := trait.parse(config, reader); err != nil {
			return err
		}
	}

	if r.Value.Payload != nil {
		if err := r.Value.Payload.parse(config, reader); err != nil {
			return err
		}
	}

	if r.Value.Headers != nil {
		if err := r.Value.Headers.parse(config, reader); err != nil {
			return err
		}
	}

	if r.Value.CorrelationId != nil {
		if err := r.Value.CorrelationId.parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (r *MessageTraitRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
	}

	if r.Value == nil {
		return nil
	}

	if r.Value.Headers != nil {
		if err := r.Value.Headers.parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}
