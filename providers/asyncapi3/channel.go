package asyncapi3

import (
	"fmt"
	"mokapi/config/dynamic"
	"regexp"

	"gopkg.in/yaml.v3"
)

type ChannelRef struct {
	dynamic.Reference[*ChannelRef]
	Value *Channel
}

type Channel struct {
	Name        string                   `yaml:"-" json:"-"`
	Title       string                   `yaml:"title" json:"title"`
	Address     string                   `yaml:"address" json:"address"`
	Summary     string                   `yaml:"summary" json:"summary"`
	Description string                   `yaml:"description" json:"description"`
	Servers     []*ServerRef             `yaml:"servers" json:"servers"`
	Messages    map[string]*MessageRef   `yaml:"messages" json:"messages"`
	Parameters  map[string]*ParameterRef `yaml:"parameters" json:"parameters"`
	Bindings    ChannelBindings          `yaml:"bindings" json:"bindings"`

	Tags         []*TagRef        `yaml:"tags" json:"tags"`
	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
	Config       *Config
}

func (r *ChannelRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *ChannelRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *ChannelRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
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

func (c *Channel) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if c == nil {
		return nil
	}

	for _, s := range c.Servers {
		if err := s.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, msg := range c.Messages {
		if err := msg.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, p := range c.Parameters {
		if err := p.Parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (c *Channel) UnmarshalYAML(node *yaml.Node) error {
	// set default
	c.Bindings.Kafka.ValueSchemaValidation = true
	c.Bindings.Kafka.Partitions = 1

	type alias Channel
	a := alias(*c)
	err := node.Decode(&a)
	if err != nil {
		return err
	}
	*c = Channel(a)
	return nil
}

func (c *Channel) UnmarshalJSON(b []byte) error {
	// set default
	c.Bindings.Kafka.ValueSchemaValidation = true
	c.Bindings.Kafka.KeySchemaValidation = true
	c.Bindings.Kafka.Partitions = 1

	type alias Channel
	a := alias(*c)
	err := dynamic.UnmarshalJSON(b, &a)
	if err != nil {
		return err
	}
	*c = Channel(a)
	return nil
}

func (c *Channel) GetName() string {
	if c.Address != "" {
		return c.Address
	}
	if c.Name != "" {
		return c.Name
	}
	return c.Title
}

func (c *Channel) IsChannelAvailable(protocol string) bool {
	if len(c.Servers) == 0 {
		return true
	}

	for _, v := range c.Servers {
		if v.Value == nil {
			continue
		}
		if protocol == v.Value.Protocol {
			return true
		}
	}
	return false
}

func (c *Channel) ResolveAddress() string {
	if c.Address != "" {
		return c.Address
	}
	return c.Name
}

// IsNameValid use if channel contains parameters
func (c *Channel) IsNameValid(topic string) error {
	address := c.ResolveAddress()

	// Find all {param} names
	re := regexp.MustCompile(`\{([^}]+)\}`)

	// Replace {param} with regex group
	pattern := "^" + re.ReplaceAllString(address, `([^/]+)`) + "$"
	re = regexp.MustCompile(pattern)

	match := re.FindStringSubmatch(topic)
	if match == nil {
		return fmt.Errorf("topic name does not match channel address expression")
	}
	return nil
}

func (c *Channel) ExtractParams(topicName string) (map[string]string, error) {
	// Find all {param} names
	re := regexp.MustCompile(`\{([^}]+)\}`)
	names := re.FindAllStringSubmatch(c.ResolveAddress(), -1)

	// Replace {param} with regex group
	pattern := "^" + re.ReplaceAllString(c.ResolveAddress(), `([^/]+)`) + "$"
	re = regexp.MustCompile(pattern)

	match := re.FindStringSubmatch(topicName)
	if match == nil {
		// path parameters are always required
		return nil, fmt.Errorf("topic name does not match channel address expression")
	}

	// Build result map
	params := map[string]string{}
	for i, name := range names {
		params[name[1]] = match[i+1]
	}

	return params, nil
}
