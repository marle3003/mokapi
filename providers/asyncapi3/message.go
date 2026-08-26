package asyncapi3

import (
	"encoding/json"
	"mokapi/config/dynamic"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type MessageRef struct {
	dynamic.Reference[*MessageRef]
	Value *Message
}

type Message struct {
	Title       string `yaml:"title,omitempty" json:"title,omitempty"`
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Summary     string `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Deprecated  bool   `yaml:"deprecated,omitempty" json:"deprecated,omitempty"`

	CorrelationId *CorrelationIdRef `yaml:"correlationId,omitempty" json:"correlationId,omitempty"`

	ContentType string             `yaml:"contentType,omitempty" json:"contentType,omitempty"`
	Headers     *SchemaRef         `yaml:"headers,omitempty" json:"headers,omitempty"`
	Payload     *SchemaRef         `yaml:"payload,omitempty" json:"payload,omitempty"`
	Bindings    *MessageBinding    `yaml:"bindings,omitempty" json:"bindings,omitempty"`
	Traits      []*MessageTraitRef `yaml:"traits,omitempty" json:"traits,omitempty"`

	Examples []interface{} `yaml:"examples,omitempty" json:"examples,omitempty"`

	ExternalDocs []*ExternalDocRef `yaml:"externalDocs,omitempty" json:"externalDocs,omitempty"`
}

type MessageTraitRef struct {
	dynamic.Reference[*MessageTraitRef]
	Value *MessageTrait
}

type MessageTrait struct {
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Title       string `yaml:"title,omitempty" json:"title,omitempty"`
	Summary     string `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	CorrelationId *CorrelationIdRef `yaml:"correlationId,omitempty" json:"correlationId,omitempty"`

	ContentType string          `yaml:"contentType,omitempty" json:"contentType,omitempty"`
	Headers     *SchemaRef      `yaml:"headers,omitempty" json:"headers,omitempty"`
	Bindings    *MessageBinding `yaml:"bindings,omitempty" json:"bindings,omitempty"`

	Examples []*MessageExample `yaml:"examples,omitempty" json:"examples,omitempty"`

	ExternalDocs []*ExternalDocRef `yaml:"externalDocs,omitempty" json:"externalDocs,omitempty"`
}

func (r *MessageRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *MessageRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *MessageRef) MarshalJSON() ([]byte, error) {
	if r.Value != nil {
		return json.Marshal(r.Value)
	}
	return json.Marshal(r.Reference)
}

func (r *MessageRef) MarshalYAML() (any, error) {
	if r.Value != nil {
		return r.Value, nil
	}
	return r.Reference, nil
}

func (r *MessageTraitRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *MessageTraitRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *MessageTraitRef) MarshalJSON() ([]byte, error) {
	if r.Value != nil {
		return json.Marshal(r.Value)
	}
	return json.Marshal(r.Reference)
}

func (r *MessageTraitRef) MarshalYAML() (any, error) {
	if r.Value != nil {
		return r.Value, nil
	}
	return r.Reference, nil
}

func (r *MessageRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}
	if r.Ref != "" {
		resolved, err := r.Resolve(config, reader)
		if err != nil {
			return err
		}
		r.Value = resolved.Value
		return nil
	}
	return r.Value.Parse(config, reader)
}

func (m *Message) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if m == nil {
		return nil
	}

	if m.Payload != nil {
		if err := m.Payload.Parse(config, reader); err != nil {
			return err
		}
	}

	if m.Headers != nil {
		if err := m.Headers.Parse(config, reader); err != nil {
			return err
		}
	}

	if m.CorrelationId != nil {
		if err := m.CorrelationId.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, trait := range m.Traits {
		if err := trait.Parse(config, reader); err != nil {
			return err
		}
		m.applyTrait(trait.Value)
	}

	if m.ContentType == "" {
		cfg, ok := config.Data.(*Config)
		if ok {
			m.ContentType = cfg.DefaultContentType
		}
		if m.ContentType == "" {
			log.Warnf("content type is missing, using default %s", DefaultContentType)
			m.ContentType = DefaultContentType
		}
	}

	if m.Bindings != nil && m.Bindings.Kafka.Key != nil {
		err := m.Bindings.Kafka.Key.Parse(config, reader)
		if err != nil {
			return err
		}
	}

	for _, doc := range m.ExternalDocs {
		if err := doc.Parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (r *MessageTraitRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		resolved, err := r.Resolve(config, reader)
		if err != nil {
			return err
		}
		r.Value = resolved.Value
		return nil
	}

	if r.Value == nil {
		return nil
	}

	if r.Value.Headers != nil {
		if err := r.Value.Headers.Parse(config, reader); err != nil {
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

	if m.Bindings != nil && m.Bindings.Kafka.Key == nil {
		m.Bindings.Kafka.Key = trait.Bindings.Kafka.Key
	}
}
