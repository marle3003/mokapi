package asyncApi

import (
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic"
)

type Server3Ref struct {
	dynamic.Reference
	Value *Server3
}

type Server3 struct {
	Host            string                        `yaml:"host" json:"host"`
	Pathname        string                        `yaml:"pathname" json:"pathname"`
	Title           string                        `yaml:"title" json:"title"`
	Summary         string                        `yaml:"summary" json:"summary"`
	Description     string                        `yaml:"description" json:"description"`
	Protocol        string                        `yaml:"protocol" json:"protocol"`
	ProtocolVersion string                        `yaml:"protocolVersion" json:"protocolVersion"`
	Variables       map[string]*ServerVariableRef `yaml:"variables" json:"variables"`
	Bindings        ServerBindings                `yaml:"bindings" json:"bindings"`
	ExternalDocs    []ExternalDocRef              `yaml:"externalDocs" json:"externalDocs"`
}

type ServerVariableRef struct {
	dynamic.Reference
	Value *ServerVariable
}

type ServerVariable struct {
	Description string   `yaml:"description" json:"description"`
	Enum        []string `yaml:"enum" json:"enum"`
	Default     string   `yaml:"default" json:"default"`
	Examples    []string `yaml:"examples" json:"examples"`
}

func (r *Server3Ref) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
	}

	for _, v := range r.Value.Variables {
		if err := v.parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (r *ServerVariableRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		if err := dynamic.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (r *Server3Ref) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *Server3Ref) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *ServerVariableRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *ServerVariableRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}
