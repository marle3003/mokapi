package asyncapi

import (
	"mokapi/config/dynamic/asyncApi"
	"mokapi/config/dynamic/openapi"
	"sort"
	"strings"
)

type Service struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Servers     []Server  `json:"servers"`
	Channels    []Channel `json:"channels"`
	Type        string    `json:"type"`
}

type Server struct {
	Url     string   `json:"url"`
	Configs []Config `json:"configs"`
}

type Channel struct {
	Name  string  `json:"name"`
	Key   *Schema `json:"key"`
	Value *Schema `json:"value"`
}

type Schema struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Properties  []*Schema `json:"properties"`
	Items       *Schema   `json:"items"`
	Ref         string    `json:"ref"`
	Description string    `json:"description"`
	Required    []string  `json:"required"`
	Format      string    `json:"format"`
	Faker       string    `json:"faker"`
	Nullable    bool      `json:"nullable"`
}

type Config struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewService(c *asyncApi.Config) Service {
	s := Service{
		Name:        c.Info.Name,
		Description: c.Info.Description,
		Version:     c.Info.Version,
		Type:        "Kafka",
	}

	for _, server := range c.Servers {
		m := server.Bindings.Kafka.ToMap()
		configs := make([]Config, 0, len(m))
		for k, v := range m {
			configs = append(configs, Config{k, v})
		}
		s.Servers = append(s.Servers, Server{
			Url:     server.Url,
			Configs: configs,
		})
	}

	for n, c := range c.Channels {
		ch := Channel{
			Name:  n,
			Key:   newSchema("", c.Value.Subscribe.Message.Value.Bindings.Kafka.Key, 0),
			Value: newSchema("", c.Value.Subscribe.Message.Value.Payload, 0),
		}
		s.Channels = append(s.Channels, ch)
	}

	return s
}

func newSchema(name string, s *openapi.SchemaRef, level int) *Schema {
	if s == nil {
		return nil
	}

	v := &Schema{
		Name:        name,
		Type:        s.Value.Type,
		Properties:  make([]*Schema, 0),
		Ref:         s.Ref,
		Description: s.Value.Description,
		Required:    s.Value.Required,
		Format:      s.Value.Format,
		Faker:       s.Value.Faker,
		Nullable:    s.Value.Nullable,
	}

	if s.Value.Items != nil {
		v.Items = newSchema("", s.Value.Items, level+1)
	}

	if level > 10 {
		return v
	}

	if s.Value.Properties != nil {
		for s, p := range s.Value.Properties.Value {
			v.Properties = append(v.Properties, newSchema(s, p, level+1))
		}
	}

	sort.Slice(v.Properties, func(i int, j int) bool {
		return strings.Compare(v.Properties[i].Name, v.Properties[j].Name) < 0
	})

	return v
}
