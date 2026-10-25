package asyncapi3

import (
	"mokapi/config/dynamic"
	"mokapi/sortedmap"

	"gopkg.in/yaml.v3"
)

// DefaultContentType set default: https://github.com/asyncapi/spec/issues/319
const DefaultContentType = "application/json"

var supportedProtocols = []string{"kafka", "mqtt", "ws"}

type mockability int

const (
	mockable mockability = iota
	notMockableNoContent
	notMockableUnsupportedServers
	notMockableNoSupportedBindings
)

type Config struct {
	Version string `yaml:"asyncapi" json:"asyncapi"`
	Id      string `yaml:"id,omitempty" json:"id,omitempty"`
	Info    Info   `yaml:"info" json:"info"`

	// Default content type to use when encoding/decoding a message's payload.
	DefaultContentType string `yaml:"defaultContentType,omitempty" json:"defaultContentType,omitempty"`

	Servers *sortedmap.LinkedHashMap[string, *ServerRef] `yaml:"servers,omitempty" json:"servers,omitempty"`

	Channels   map[string]*ChannelRef   `yaml:"channels,omitempty" json:"channels,omitempty"`
	Operations map[string]*OperationRef `yaml:"operations,omitempty" json:"operations,omitempty"`

	Components *Components `yaml:"components,omitempty" json:"components,omitempty"`
}

type Info struct {
	Name           string           `yaml:"title" json:"title"`
	Description    string           `yaml:"description,omitempty" json:"description,omitempty"`
	Version        string           `yaml:"version,omitempty" json:"version,omitempty"`
	TermsOfService string           `yaml:"termsOfService,omitempty" json:"termsOfService,omitempty"`
	Contact        *Contact         `yaml:"contact,omitempty" json:"contact,omitempty"`
	License        *License         `yaml:"license,omitempty" json:"license,omitempty"`
	ExternalDocs   []ExternalDocRef `yaml:"externalDocs,omitempty" json:"externalDocs,omitempty"`
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
	if c.Servers != nil {
		for it := c.Servers.Iter(); it.Next(); {
			server := it.Value()
			if err := server.Parse(config, reader); err != nil {
				return err
			}
		}
	}

	for name, ch := range c.Channels {
		if err := ch.Parse(config, reader); err != nil {
			return err
		}
		if ch.Value != nil {
			ch.Value.Name = name
			ch.Value.Config = c
		}
	}

	for _, op := range c.Operations {
		if err := op.Parse(config, reader); err != nil {
			return err
		}
	}

	return c.Components.parse(config, reader)
}

func (c *Config) UnmarshalYAML(node *yaml.Node) error {
	// set default: https://github.com/asyncapi/spec/issues/319
	c.DefaultContentType = DefaultContentType

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
	c.DefaultContentType = DefaultContentType

	type alias Config
	a := alias(*c)
	err := dynamic.UnmarshalJSON(b, &a)
	if err != nil {
		return err
	}
	*c = Config(a)
	return nil
}

func (c *Config) Mockability() dynamic.Mockability {
	a := c.Analyze()
	for _, u := range a.Protocols {
		if len(u.Servers) > 0 || (!a.HasServers && len(u.Channels) > 0) {
			return dynamic.Mockability{}
		}
	}

	switch {
	case a.HasServers:
		return dynamic.Mockability{
			Level:  dynamic.MockabilityInfo,
			Reason: "contains only unsupported servers",
		}
	case a.HasChannels:
		return dynamic.Mockability{
			Level:  dynamic.MockabilityWarn,
			Reason: "has no servers and no supported bindings; it will not be served",
		}
	default:
		return dynamic.Mockability{}
	}
}
