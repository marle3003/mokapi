package asyncapi3

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
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

func (r *ExternalDocRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}

	return nil
}
