package asyncapi3

import (
	"encoding/json"
	"mokapi/schema/json/schema"
)

type WebsocketChannelBindings struct {
	Method  string         `yaml:"method,omitempty" json:"method,omitempty"`
	Query   *schema.Schema `yaml:"query,omitempty" json:"query,omitempty"`
	Headers *schema.Schema `yaml:"headers,omitempty" json:"headers,omitempty"`
}

func (b *WebsocketChannelBindings) IsSet() bool {
	return b.Method != "" || b.Query != nil || b.Headers != nil
}

func (b *WebsocketChannelBindings) patch(patch WebsocketChannelBindings) {
	if patch.Method != "" {
		b.Method = patch.Method
	}
	if b.Query != nil {
		b.Query.Patch(patch.Query)
	} else {
		b.Query = patch.Query
	}
	if b.Headers != nil {
		b.Headers.Patch(patch.Headers)
	} else {
		b.Headers = patch.Headers
	}
}

type WebsocketOperationBindings struct {
	configs map[string]any
}

func (b *WebsocketOperationBindings) IsSet() bool {
	return b.configs != nil
}

func (b *WebsocketOperationBindings) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.configs)
}

func (b *WebsocketOperationBindings) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &b.configs)
}

type WebsocketMessageBindings struct {
	configs map[string]any
}

func (b *WebsocketMessageBindings) IsSet() bool {
	return b.configs != nil
}

func (b *WebsocketMessageBindings) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.configs)
}

func (b *WebsocketMessageBindings) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &b.configs)
}
