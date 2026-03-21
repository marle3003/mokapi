package asyncapi3

import (
	"mokapi/config/dynamic"

	"gopkg.in/yaml.v3"
)

type ExternalDocRef struct {
	dynamic.Reference
	Value *ExternalDoc
}

type ExternalDoc struct {
	Description string `yaml:"description" json:"description"`
	Url         string `yaml:"url" json:"url"`
}

func (r *ExternalDocRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *ExternalDocRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *ExternalDocRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		var resolved *ExternalDocRef
		if err := dynamic.Resolve(r.Ref, &resolved, config, reader); err != nil {
			return err
		}
		r.Value = resolved.Value
	}

	return nil
}
