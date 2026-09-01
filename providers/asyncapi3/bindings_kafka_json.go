package asyncapi3

import "encoding/json"

func (b *BrokerBindings) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.marshal())
}

func (b *BrokerBindings) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &b.configs)
	if err != nil {
		return err
	}
	return b.unmarshal()
}

func (t *TopicBindings) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.marshal())
}

func (t *TopicBindings) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &t.configs)
	if err != nil {
		return err
	}
	return t.unmarshal(t.configs)
}
