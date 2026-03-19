package asyncapi3

import (
	"mokapi/config/dynamic"

	"gopkg.in/yaml.v3"
)

type ServerRef struct {
	dynamic.Reference
	Value *Server
}

type Server struct {
	Host            string                        `yaml:"host" json:"host"`
	Pathname        string                        `yaml:"pathname" json:"pathname"`
	Title           string                        `yaml:"title" json:"title"`
	Summary         string                        `yaml:"summary" json:"summary"`
	Description     string                        `yaml:"description" json:"description"`
	Protocol        string                        `yaml:"protocol" json:"protocol"`
	ProtocolVersion string                        `yaml:"protocolVersion" json:"protocolVersion"`
	Variables       map[string]*ServerVariableRef `yaml:"variables" json:"variables"`
	Tags            []*TagRef                     `yaml:"tags" json:"tags"`
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

func (r *ServerRef) Parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		var resolved *ServerRef
		if err := dynamic.Resolve(r.Ref, &resolved, config, reader); err != nil {
			return err
		}
		r.Value = resolved.Value
		return nil
	}

	if r.Value == nil {
		return nil
	}

	for _, v := range r.Value.Variables {
		if err := v.parse(config, reader); err != nil {
			return err
		}
	}

	for _, v := range r.Value.Tags {
		if err := v.parse(config, reader); err != nil {
			return err
		}
	}

	for _, v := range r.Value.ExternalDocs {
		if err := v.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, t := range r.Value.Tags {
		if err := t.parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (r *ServerVariableRef) parse(config *dynamic.Config, reader dynamic.Reader) error {
	if len(r.Ref) > 0 {
		return dynamic.Resolve(r.Ref, &r.Value, config, reader)
	}

	return nil
}

func (r *ServerRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *ServerRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *ServerVariableRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *ServerVariableRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}
