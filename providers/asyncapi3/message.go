package asyncapi3

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type MessageRef struct {
	dynamic.Reference
	Value *Message
}

type Message struct {
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
	Name        string `yaml:"name" json:"name"`
	Title       string `yaml:"title" json:"title"`
	Summary     string `yaml:"summary" json:"summary"`
	Description string `yaml:"description" json:"description"`

	CorrelationId *CorrelationIdRef `yaml:"correlationId" json:"correlationId"`

	ContentType string         `yaml:"contentType" json:"contentType"`
	Headers     *SchemaRef     `yaml:"headers" json:"headers"`
	Bindings    MessageBinding `yaml:"bindings" json:"bindings"`

	Examples []*MessageExample `yaml:"examples" json:"examples"`

	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

func (r *MessageRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *MessageRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *MessageTraitRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *MessageTraitRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *MessageRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}

	if r.Value == nil {
		return nil
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

	for _, trait := range r.Value.Traits {
		if err := trait.parse(config, reader); err != nil {
			return err
		}
		r.Value.applyTrait(trait.Value)
	}

	return nil
}

func (r *MessageTraitRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
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

func (m *Message) applyTrait(trait *MessageTrait) {
	if trait == nil {
		return
	}

	if len(m.Name) == 0 {
		m.Name = trait.Name
	}
	if len(m.Title) == 0 {
		m.Title = trait.Title
	}
	if len(m.Summary) == 0 {
		m.Summary = trait.Summary
	}
	if len(m.Description) == 0 {
		m.Description = trait.Description
	}
	if m.CorrelationId == nil {
		m.CorrelationId = trait.CorrelationId
	}
	if len(m.ContentType) == 0 {
		m.ContentType = trait.ContentType
	}
	if m.Headers == nil {
		m.Headers = trait.Headers
	}

	m.Examples = append(m.Examples, trait.Examples)
	m.ExternalDocs = append(m.ExternalDocs, trait.ExternalDocs...)

	if m.Bindings.Kafka.Key == nil {
		m.Bindings.Kafka.Key = trait.Bindings.Kafka.Key
	}
}
