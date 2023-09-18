package parameter

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
	if param.Style == "" {
		p.SetDefaultStyle()
	}
	*p = Parameter(param)

	return nil
}

func (r *Ref) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}

func (p *Parameter) UnmarshalJSON(b []byte) error {
	type alias Parameter
	param := alias{}
	err := json.Unmarshal(b, &param)
	if err != nil {
		return err
	}
	if param.Style == "" {
		p.SetDefaultStyle()
	}
	*p = Parameter(param)
	return nil
}

func (r *Ref) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}
