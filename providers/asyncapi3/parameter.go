package asyncapi3

import (
	"mokapi/config/dynamic"

	"gopkg.in/yaml.v3"
)

type ParameterRef struct {
	dynamic.Reference[*ParameterRef]
	Value *Parameter
}

type Parameter struct {
	Description string   `yaml:"description" json:"description"`
	Enum        []string `yaml:"enum" json:"enum"`
	Default     string   `yaml:"default" json:"default"`
	Examples    []string `yaml:"examples" json:"examples"`
	Location    string   `yaml:"location" json:"location"`
}

func (r *ParameterRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *ParameterRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *ParameterRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		resolved, err := r.Resolve(config, reader)
		if err != nil {
			return err
		}
		r.Value = resolved.Value
		return nil
	}

	return nil
}
