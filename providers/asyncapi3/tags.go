package asyncapi3

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type TagRef struct {
	dynamic.Reference
	Value *Tag
}

type Tag struct {
	Name         string           `yaml:"name" json:"name"`
	Description  string           `yaml:"description" json:"description"`
	ExternalDocs []ExternalDocRef `yaml:"externalDocs" json:"externalDocs"`
}

func (r *TagRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *TagRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *TagRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}

	if r.Value == nil {
		return nil
	}

	for _, v := range r.Value.ExternalDocs {
		if err := v.parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}
