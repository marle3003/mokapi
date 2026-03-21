package asyncapi3

import (
	"mokapi/config/dynamic"

	"gopkg.in/yaml.v3"
)

type CorrelationIdRef struct {
	dynamic.Reference
	Value *CorrelationId
}

type CorrelationId struct {
	Description string `yaml:"description" json:"description"`
	Location    string `yaml:"location" json:"location"`
}

func (r *CorrelationIdRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		var resolved *CorrelationIdRef
		if err := dynamic.Resolve(r.Ref, &resolved, config, reader); err != nil {
			return err
		}
		r.Value = resolved.Value
	}
	return nil
}

func (r *CorrelationIdRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *CorrelationIdRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}
