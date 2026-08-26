package asyncapi3

import (
	"encoding/json"
	"mokapi/config/dynamic"

	"gopkg.in/yaml.v3"
)

type CorrelationIdRef struct {
	dynamic.Reference[*CorrelationIdRef]
	Value *CorrelationId
}

type CorrelationId struct {
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Location    string `yaml:"location,omitempty" json:"location,omitempty"`
}

func (r *CorrelationIdRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
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

func (r *CorrelationIdRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *CorrelationIdRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *CorrelationIdRef) MarshalJSON() ([]byte, error) {
	if r.Value != nil {
		return json.Marshal(r.Value)
	}
	return json.Marshal(r.Reference)
}

func (r *CorrelationIdRef) MarshalYAML() (any, error) {
	if r.Value != nil {
		return r.Value, nil
	}
	return r.Reference, nil
}
