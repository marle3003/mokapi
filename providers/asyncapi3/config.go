package asyncapi3

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type Config struct {
	Version string `yaml:"asyncapi" json:"asyncapi"`
	Id      string `yaml:"id" json:"id"`
	Info    Info   `yaml:"info" json:"info"`

	// Default content type to use when encoding/decoding a message's payload.
	DefaultContentType string `yaml:"defaultContentType" json:"defaultContentType"`

	Servers map[string]*ServerRef `yaml:"servers" json:"servers"`

	Channels   map[string]*ChannelRef
	Operations map[string]*OperationRef `yaml:"operations" json:"operations"`

	Components *Components `yaml:"components,omitempty" json:"components,omitempty"`
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

type Contact struct {
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	Url   string `yaml:"url,omitempty" json:"url,omitempty"`
	Email string `yaml:"email,omitempty" json:"email,omitempty"`
}

type License struct {
	Name string `yaml:"name" json:"name"`
	Url  string `yaml:"url" json:"url"`
}

func (c *Config) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	for _, server := range c.Servers {
		if len(server.Ref) > 0 {
			return dynamic.Resolve(server.Ref, &server.Value, config, reader)
		}
		if server.Value == nil {
			return nil
		}
		if err := server.parse(config, reader); err != nil {
			return err
		}
	}

	for name, ch := range c.Channels {
		if err := ch.parse(config, reader); err != nil {
			return err
		}
		if ch.Value != nil {
			ch.Value.Name = name
		}
	}

	for _, op := range c.Operations {
		if err := op.parse(config, reader); err != nil {
			return err
		}
	}

	return c.Components.parse(config, reader)
}

func (c *Config) UnmarshalYAML(node *yaml.Node) error {
	// set default: https://github.com/asyncapi/spec/issues/319
	c.DefaultContentType = "application/json"

	type alias Config
	a := alias(*c)
	err := node.Decode(&a)
	if err != nil {
		return err
	}
	*c = Config(a)
	return nil
}

func (c *Config) UnmarshalJSON(b []byte) error {
	// set default: https://github.com/asyncapi/spec/issues/319
	c.DefaultContentType = "application/json"

	type alias Config
	a := alias(*c)
	err := dynamic.UnmarshalJSON(b, &a)
	if err != nil {
		return err
	}
	*c = Config(a)
	return nil
}
