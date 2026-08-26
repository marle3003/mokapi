package asyncapi3

import (
	"encoding/json"
	"mokapi/config/dynamic"

	"gopkg.in/yaml.v3"
)

type ServerRef struct {
	dynamic.Reference[*ServerRef]
	Value *Server
}

type Server struct {
	Host            string                        `yaml:"host,omitempty" json:"host,omitempty"`
	Pathname        string                        `yaml:"pathname,omitempty" json:"pathname,omitempty"`
	Title           string                        `yaml:"title,omitempty" json:"title,omitempty"`
	Summary         string                        `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description     string                        `yaml:"description,omitempty" json:"description,omitempty"`
	Protocol        string                        `yaml:"protocol,omitempty" json:"protocol,omitempty"`
	ProtocolVersion string                        `yaml:"protocolVersion,omitempty" json:"protocolVersion,omitempty"`
	Variables       map[string]*ServerVariableRef `yaml:"variables,omitempty" json:"variables,omitempty"`
	Tags            []*TagRef                     `yaml:"tags,omitempty" json:"tags,omitempty"`
	Bindings        *ServerBindings               `yaml:"bindings,omitempty" json:"bindings,omitempty"`
	ExternalDocs    []ExternalDocRef              `yaml:"externalDocs,omitempty" json:"externalDocs,omitempty"`
}

type ServerVariableRef struct {
	dynamic.Reference[ServerVariableRef]
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
		resolved, err := r.Resolve(config, reader)
		if err != nil {
			return err
		}
		r.Value = resolved.Value
		return nil
	}

	return nil
}

func (r *ServerRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *ServerRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *ServerRef) MarshalJSON() ([]byte, error) {
	if r.Value != nil {
		return json.Marshal(r.Value)
	}
	return json.Marshal(r.Reference)
}

func (r *ServerRef) MarshalYAML() (any, error) {
	if r.Value != nil {
		return r.Value, nil
	}
	return r.Reference, nil
}

func (r *ServerVariableRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (r *ServerVariableRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}
