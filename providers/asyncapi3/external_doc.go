package asyncapi3

import (
	"encoding/json"
	"mokapi/config/dynamic"

	"gopkg.in/yaml.v3"
)

type ExternalDocRef struct {
	dynamic.Reference[*ExternalDocRef]
	Value *ExternalDoc
}

type ExternalDoc struct {
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Url         string `yaml:"url,omitempty" json:"url,omitempty"`
}

func (r *ExternalDocRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *ExternalDocRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *ExternalDocRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
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

func (r *ExternalDocRef) MarshalJSON() ([]byte, error) {
	if r.Value != nil {
		return json.Marshal(r.Value)
	}
	return json.Marshal(r.Reference)
}

func (r *ExternalDocRef) MarshalYAML() (any, error) {
	if r.Value != nil {
		return r.Value, nil
	}
	return r.Reference, nil
}
