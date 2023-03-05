package parameter

import (
	"bytes"
	"encoding/json"
	"gopkg.in/yaml.v3"
	"mokapi/config/dynamic/openapi/schema"
)

type parameter struct {
	Name        string
	Type        Location `yaml:"in" json:"in"`
	Schema      *schema.Ref
	Required    bool
	Description string
	Style       string
	Explode     bool
}

func (p *Parameter) UnmarshalYAML(value *yaml.Node) error {
	tmp := &parameter{Explode: true}
	err := value.Decode(tmp)
	if err != nil {
		return err
	}

	p.Name = tmp.Name
	p.Type = tmp.Type
	p.Schema = tmp.Schema
	p.Required = tmp.Required
	p.Description = tmp.Description
	p.Style = tmp.Style
	p.Explode = tmp.Explode

	return nil
}

func (r *Ref) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}

func (r *NamedParameters) UnmarshalYAML(node *yaml.Node) error {
	return r.Reference.Unmarshal(node, &r.Value)
}

func (p *Parameter) UnmarshalJSON(b []byte) error {
	tmp := &parameter{Explode: true}
	dec := json.NewDecoder(bytes.NewReader(b))
	err := dec.Decode(&tmp)
	if err != nil {
		return err
	}

	p.Name = tmp.Name
	p.Type = tmp.Type
	p.Schema = tmp.Schema
	p.Required = tmp.Required
	p.Description = tmp.Description
	p.Style = tmp.Style
	p.Explode = tmp.Explode

	return nil
}

func (r *Ref) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}

func (r *NamedParameters) UnmarshalJSON(b []byte) error {
	return r.Reference.UnmarshalJson(b, &r.Value)
}
