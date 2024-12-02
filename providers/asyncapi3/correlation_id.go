package asyncapi3

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type CorrelationIdRef struct {
	dynamic.Reference
	Value *CorrelationId
}

type CorrelationId struct {
	Description string `yaml:"description" json:"description"`
	Location    string `yaml:"location" json:"location"`
}

func (r *CorrelationIdRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
	}
	return nil
}

func (r *CorrelationIdRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *CorrelationIdRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}
