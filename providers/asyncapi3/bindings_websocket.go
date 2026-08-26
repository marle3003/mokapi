package asyncapi3

import "mokapi/schema/json/schema"

type WebsocketChannelBindings struct {
	Method  string         `yaml:"method,omitempty" json:"method,omitempty"`
	Query   *schema.Schema `yaml:"query,omitempty" json:"query,omitempty"`
	Headers *schema.Schema `yaml:"headers,omitempty" json:"headers,omitempty"`
}
