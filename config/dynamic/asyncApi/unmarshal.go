package asyncApi

import (
	"gopkg.in/yaml.v3"
)

type refProp struct {
	Ref string `yaml:"$ref" json:"$ref"`
}

func (r *MessageRef) UnmarshalYAML(node *yaml.Node) error {
	return unmarshalRef(node, &r.Ref, &r.Value)
}

func (r *ChannelRef) UnmarshalYAML(node *yaml.Node) error {
	return unmarshalRef(node, &r.Ref, &r.Value)
}

func (r *ServerRef) UnmarshalYAML(node *yaml.Node) error {
	return unmarshalRef(node, &r.Ref, &r.Value)
}

func unmarshalRef(node *yaml.Node, ref *string, val interface{}) error {
	r := &refProp{}
	if err := node.Decode(r); err == nil && len(r.Ref) > 0 {
		*ref = r.Ref
		return nil
	}

	return node.Decode(val)
}

func (r *ParameterRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *MessageTraitRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *MessageTraitRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}
