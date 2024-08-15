package asyncApi

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type Config3 struct {
	Version string `yaml:"asyncapi" json:"asyncapi"`
	Id      string `yaml:"id" json:"id"`
	Info    Info   `yaml:"info" json:"info"`

	// Default content type to use when encoding/decoding a message's payload.
	DefaultContentType string `yaml:"defaultContentType" json:"defaultContentType"`

	Servers map[string]*Server3Ref `yaml:"servers" json:"servers"`

	Channels   map[string]*Channel3Ref
	Operations map[string]*Operation3Ref `yaml:"operations" json:"operations"`

	Components *Components3 `yaml:"components,omitempty" json:"components,omitempty"`
}

type Info struct {
	Name           string           `yaml:"title" json:"title"`
	Description    string           `yaml:"description,omitempty" json:"description,omitempty"`
	Version        string           `yaml:"version" json:"version"`
	TermsOfService string           `yaml:"termsOfService,omitempty" json:"termsOfService,omitempty"`
	Contact        *Contact         `yaml:"contact,omitempty" json:"contact,omitempty"`
	License        *License         `yaml:"license,omitempty" json:"license,omitempty"`
	ExternalDocs   []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

type CorrelationIdRef struct {
	dynamic.Reference
	Value *CorrelationId
}

type CorrelationId struct {
	Description string `yaml:"description" json:"description"`
	Location    string `yaml:"location" json:"location"`
}

type ExternalDocRef struct {
	dynamic.Reference
	Value *ExternalDoc
}

type ExternalDoc struct {
	Description string `yaml:"description" json:"description"`
	Url         string `yaml:"url" json:"url"`
}

func (c *Config3) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	for _, server := range c.Servers {
		if len(server.Ref) == 0 {
			continue
		}
		if err := dynamic.Resolve(server.Ref, &server.Value, config, reader); err != nil {
			return err
		}
	}

	for _, ch := range c.Channels {
		if err := ch.parse(config, reader); err != nil {
			return err
		}
	}

	for _, op := range c.Operations {
		if err := op.parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (r *CorrelationIdRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
	}
	return nil
}

func (r *CorrelationIdRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *CorrelationIdRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}
