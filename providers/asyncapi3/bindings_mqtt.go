package asyncapi3

import "encoding/json"

type MqttChannelBindings struct {
	configs map[string]any
}

func (b *MqttChannelBindings) IsSet() bool {
	return b.configs != nil
}

func (b *MqttChannelBindings) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.configs)
}

func (b *MqttChannelBindings) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &b.configs)
}

type MqttOperationBindings struct {
	configs map[string]any
}

func (b *MqttOperationBindings) IsSet() bool {
	return b.configs != nil
}

func (b *MqttOperationBindings) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.configs)
}

func (b *MqttOperationBindings) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &b.configs)
}

type MqttMessageBindings struct {
	configs map[string]any
}

func (b *MqttMessageBindings) IsSet() bool {
	return b.configs != nil
}

func (b *MqttMessageBindings) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.configs)
}

func (b *MqttMessageBindings) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &b.configs)
}
