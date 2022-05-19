package ref

import (
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
