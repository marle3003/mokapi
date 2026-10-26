package asyncapi3

import (
	"bytes"
	"encoding/json"
)

type ServerBindings struct {
	Kafka BrokerBindings `yaml:"kafka" json:"kafka"`
}

type ChannelBindings struct {
	Kafka     TopicBindings            `yaml:"kafka" json:"kafka"`
	Mqtt      MqttChannelBindings      `yaml:"mqtt" json:"mqtt"`
	Websocket WebsocketChannelBindings `yaml:"ws" json:"ws"`
}

type OperationBindings struct {
	Kafka     KafkaOperationBinding      `yaml:"kafka" json:"kafka"`
	Mqtt      MqttOperationBindings      `yaml:"mqtt" json:"mqtt"`
	Websocket WebsocketOperationBindings `yaml:"ws" json:"ws"`
}

type MessageBinding struct {
	Kafka     KafkaMessageBinding      `yaml:"kafka" json:"kafka"`
	Mqtt      MqttMessageBindings      `yaml:"mqtt" json:"mqtt"`
	Websocket WebsocketMessageBindings `yaml:"ws" json:"ws"`
}

func (s *ServerBindings) MarshalJSON() ([]byte, error) {
	result := new(bytes.Buffer)
	result.WriteString("{")

	b, err := json.Marshal(&s.Kafka)
	if err != nil {
		return nil, err
	}
	if len(b) > 2 {
		result.WriteString(`"kafka": `)
		result.Write(b)
	}

	result.WriteString("}")
	return result.Bytes(), nil
}

func (s *ServerBindings) MarshalYAML() (any, error) {
	m := map[string]any{}

	kafka := s.Kafka.marshal()
	if len(kafka) > 0 {
		m["kafka"] = kafka
	}

	return m, nil
}

func (c *ChannelBindings) MarshalJSON() ([]byte, error) {
	result := new(bytes.Buffer)
	result.WriteString("{")

	b, err := json.Marshal(&c.Kafka)
	if err != nil {
		return nil, err
	}
	if len(b) > 2 {
		result.WriteString(`"kafka": `)
		result.Write(b)
	}

	b, err = json.Marshal(&c.Websocket)
	if err != nil {
		return nil, err
	}
	if len(b) > 2 {
		if result.Len() > 1 {
			result.WriteString(`,`)
		}
		result.WriteString(`"websocket": `)
		result.Write(b)
	}

	result.WriteString("}")
	return result.Bytes(), nil
}

func (c *ChannelBindings) MarshalYAML() (any, error) {
	m := map[string]any{}

	kafka := c.Kafka.marshal()
	if len(kafka) > 0 {
		m["kafka"] = kafka
	}

	if len(c.Websocket.Method) > 0 || c.Websocket.Query != nil || c.Websocket.Headers != nil {
		m["websocket"] = c.Websocket
	}

	return m, nil
}

func (s *ServerBindings) patch(patch *ServerBindings) {
	if patch == nil {
		return
	}

	s.Kafka.Patch(patch.Kafka)
}

func (c *ChannelBindings) patch(patch *ChannelBindings) {
	if patch == nil {
		return
	}

	c.Kafka.Patch(patch.Kafka)
	c.Websocket.patch(patch.Websocket)
}

func (o *OperationBindings) patch(patch *OperationBindings) {
	if patch == nil {
		return
	}

	o.Kafka.Patch(patch.Kafka)
}

func (m *MessageBinding) patch(patch *MessageBinding) {
	if patch == nil {
		return
	}

	m.Kafka.Patch(patch.Kafka)
}
