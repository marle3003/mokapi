package asyncapi3

import (
	"mokapi/config/dynamic"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
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

	ExternalDocs []*ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
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

	ExternalDocs []*ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
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

func (r *MessageRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if r == nil {
		return nil
	}
	if r.Ref != "" {
		var resolved *MessageRef
		if err := dynamic.Resolve(r.Ref, &resolved, config, reader); err != nil {
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

	if m.Bindings.Kafka.Key != nil {
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
		var resolved *MessageTraitRef
		if err := dynamic.Resolve(r.Ref, &resolved, config, reader); err != nil {
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

	if m.Bindings.Kafka.Key == nil {
		m.Bindings.Kafka.Key = trait.Bindings.Kafka.Key
	}
}
