package openapi

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
)

func (p *Parameter) UnmarshalYAML(value *yaml.Node) error {
	type alias Parameter
	param := alias{}
	err := value.Decode(&param)
	if err != nil {
		return err
	}
	*p = Parameter(param)
	if p.Style == "" {
		p.SetDefaultStyle()
	}

	return nil
}

func (r *ParameterRef) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.UnmarshalYaml(node, &r.Value)
}

func (p *Parameter) UnmarshalJSON(b []byte) error {
	type alias Parameter
	param := alias{}
	err := json.Unmarshal(b, &param)
	if err != nil {
		return err
	}
	*p = Parameter(param)
	if p.Style == "" {
		p.SetDefaultStyle()
	}
	return nil
}

func (r *ParameterRef) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}
