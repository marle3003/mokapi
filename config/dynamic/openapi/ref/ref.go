package ref

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
)

type Resolver interface {
	Resolve() interface{}
}

type Reference struct {
	Ref string `yaml:"$ref" json:"$ref"`
}

func (r *Reference) Unmarshal(node *yaml.Node, val interface{}) error {
	err := node.Decode(r)
	if err == nil && len(r.Ref) > 0 {
		return nil
	}

	return node.Decode(val)
}

func (r *Reference) UnmarshalJson(b []byte, val interface{}) error {
	if err := json.Unmarshal(b, r); err == nil && len(r.Ref) > 0 {
		return nil
	}
	return json.Unmarshal(b, val)
}
