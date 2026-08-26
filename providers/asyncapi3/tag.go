package asyncapi3

import (
	"encoding/json"
	"mokapi/config/dynamic"

	"gopkg.in/yaml.v3"
)

type TagRef struct {
	dynamic.Reference[*TagRef]
	Value *Tag
}

type Tag struct {
	Name         string            `yaml:"name,omitempty" json:"name,omitempty"`
	Description  string            `yaml:"description,omitempty" json:"description,omitempty"`
	ExternalDocs []*ExternalDocRef `yaml:"externalDocs,omitempty" json:"externalDocs,omitempty"`
}

func (r *TagRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *TagRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *TagRef) MarshalJSON() ([]byte, error) {
	if r.Value != nil {
		return json.Marshal(r.Value)
	}
	return json.Marshal(r.Reference)
}

func (r *TagRef) MarshalYAML() (any, error) {
	if r.Value != nil {
		return r.Value, nil
	}
	return r.Reference, nil
}

func (r *TagRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		resolved, err := r.Resolve(config, reader)
		if err != nil {
			return err
		}
		r.Value = resolved.Value
		return nil
	}

	if r.Value == nil {
		return nil
	}

	for _, v := range r.Value.ExternalDocs {
		if err := v.Parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}
