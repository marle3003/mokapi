package ref

import "gopkg.in/yaml.v3"

type Reference struct {
	Value string `yaml:"$ref" json:"$ref"`
}

func (r *Reference) Ref() string {
	return r.Value
}

func (r *Reference) Unmarshal(node *yaml.Node, val interface{}) error {
	err := node.Decode(r)
	if err == nil && len(r.Value) > 0 {
		return nil
	}

	return node.Decode(val)
}
