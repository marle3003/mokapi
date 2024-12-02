package asyncapi3

import (
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
			if err := dynamic.Resolve(server.Ref, &server.Value, config, reader); err != nil {
				return err
			}
		}
		if server.Value == nil {
			return nil
		}
		if err := server.parse(config, reader); err != nil {
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

	return c.Components.parse(config, reader)
}
