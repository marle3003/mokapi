package dynamic

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
)

type Reference struct {
	Ref         string `yaml:"$ref" json:"$ref"`
	Summary     string `yaml:"summary" json:"summary"`
	Description string `yaml:"description" json:"description"`
}

func (r *Reference) UnmarshalYaml(node *yaml.Node, val interface{}) error {
	err := node.Decode(r)
	if err == nil && len(r.Ref) > 0 {
		return nil
	}

	return node.Decode(val)
}

func (r *Reference) UnmarshalJson(b []byte, val interface{}) error {
	var m map[string]string
	_ = json.Unmarshal(b, &m)
	if _, ok := m["$ref"]; ok {
		return UnmarshalJSON(b, r)
	}

	err := UnmarshalJSON(b, val)
	return err
}
