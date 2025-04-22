package schema

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

// The gopkg.in/yaml.v3 package automatically interprets date-like strings as time.Time
// Therefore, we must use a custom parser, otherwise we will get a time.Time object that contains,
// for example, the time part in a string value, even if the value was defined as 2022-11-21

type Example struct {
	Value
}

type Value interface{}

func (e *Example) UnmarshalJSON(b []byte) error {
	return dynamic.UnmarshalJSON(b, &e.Value)
}

func (e *Example) UnmarshalYAML(node *yaml.Node) error {
	v, err := dynamic.ParseYamlPlain(node)
	if err != nil {
		return err
	}
	e.Value = v
	return nil
}

func (e *Example) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Value)
}

func (e *Example) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(e.Value)
}
