package asyncapi3

import "mokapi/schema/json/schema"

type WebsocketChannelBindings struct {
	Method  string         `yaml:"method" json:"method"`
	Query   *schema.Schema `yaml:"query" json:"query"`
	Headers *schema.Schema `yaml:"headers" json:"headers"`
}
